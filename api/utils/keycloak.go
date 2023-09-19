package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/mail"
	"spay/models"
	"strconv"
	"strings"

	"github.com/Nerzal/gocloak/v12"
	uuid "github.com/satori/go.uuid"
)

const (
	USER_NOT_EXIST = "utilisateur inexistant"
)

func KeycloakLoginClient(config models.Config) (*gocloak.GoCloak, *gocloak.JWT, error) {
	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	// Connexion connexion en tant qu'admin
	kcClientToken, err := kcClient.LoginClient(ctx, config.KeyCloakClientID, config.KeyCloakClientSecret, config.KeyCloakClientRealm)
	if err != nil {
		return kcClient, nil, err
	}

	return kcClient, kcClientToken, nil
}

func KeycloakLogin(config models.Config, username string, password string) (*gocloak.GoCloak, *gocloak.JWT, error) {
	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	// Connexion utilisateur
	kcUserToken, err := kcClient.Login(ctx, config.KeyCloakClientID, config.KeyCloakClientSecret, config.KeyCloakClientRealm, username, password)
	if err != nil {
		return kcClient, nil, err
	}

	return kcClient, kcUserToken, nil
}

func KeycloakGetOneUser(config models.Config, adminAccessToken string, username string, isEmail bool) (*gocloak.User, error) {
	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	userParams := gocloak.GetUsersParams{
		Username: gocloak.StringP(username),
		Exact:    gocloak.BoolP(true),
	}

	if isEmail {
		userParams = gocloak.GetUsersParams{
			Email: gocloak.StringP(username),
			Exact: gocloak.BoolP(true),
		}
	}

	users, err := kcClient.GetUsers(ctx, adminAccessToken, config.KeyCloakClientRealm, userParams)

	if err != nil {
		return nil, err
	}

	if len(users) <= 0 {
		err := fmt.Errorf(USER_NOT_EXIST)

		return nil, err
	}

	return users[0], nil
}

func KeycloakGetUsers(config models.Config, adminAccessToken string, params gocloak.GetUsersParams) ([]*gocloak.User, error) {
	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	users, err := kcClient.GetUsers(ctx, adminAccessToken, config.KeyCloakClientRealm, params)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func KeycloakGetUserByID(config models.Config, adminAccessToken string, username string) (*gocloak.User, error) {
	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	user, err := kcClient.GetUserByID(ctx, adminAccessToken, config.KeyCloakClientRealm, username)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func KeycloakGetUserAndToken(config models.Config, adminAccessToken string, username string, password string) (*gocloak.JWT, *gocloak.User, error) {
	var kcUserToken *gocloak.JWT
	var user *gocloak.User
	var errToken error

	// Si numéro de téléphone
	if _, err := strconv.Atoi(username); err == nil {
		_, kcUserToken, errToken = KeycloakLogin(config, username, password)

		if errToken == nil {
			user, err = KeycloakGetOneUser(config, adminAccessToken, username, false)

			if err != nil {
				return nil, nil, err
			}
		}
	} else if _, err := mail.ParseAddress(username); err == nil { // Si email
		user, err = KeycloakGetOneUser(config, adminAccessToken, username, true)
		if err != nil {
			return nil, nil, err
		}

		_, kcUserToken, errToken = KeycloakLogin(config, *user.Username, password)
	} else if _, err := uuid.FromString(username); err == nil { // Si id
		user, err = KeycloakGetUserByID(config, adminAccessToken, username)
		if err != nil {
			return nil, nil, err
		}

		_, kcUserToken, errToken = KeycloakLogin(config, *user.Username, password)
	} else {
		kcUserToken = nil
		errToken = fmt.Errorf(USER_NOT_EXIST)
	}

	return kcUserToken, user, errToken
}

// KeycloakNewUser fonction de création de compte avec le client keycloak
func KeycloakNewUser(config models.Config, clientAccessToken string, newUser *gocloak.User, password string) (*gocloak.JWT, error) {
	// Vérification de l'email
	if len(*newUser.Email) > 0 {
		// Récuperation de la liste des utilisateur
		users, _ := KeycloakGetUsers(config, clientAccessToken, gocloak.GetUsersParams{
			Email: gocloak.StringP(*newUser.Email),
			Exact: gocloak.BoolP(true),
		})

		if len(users) > 0 {
			return nil, fmt.Errorf("email déjà utilisé par un autre utilisateur")
		}
	}

	// Vérification du phone number
	if len(*newUser.Username) > 0 {
		// Récuperation de la liste des utilisateur
		users, _ := KeycloakGetUsers(config, clientAccessToken, gocloak.GetUsersParams{
			Username: gocloak.StringP(*newUser.Username),
			Exact:    gocloak.BoolP(true),
		})

		if len(users) > 0 {
			return nil, fmt.Errorf("numéro déjà utilisé par un autre utilisateur")
		}
	}

	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	newUserId, err := kcClient.CreateUser(ctx, clientAccessToken, config.KeyCloakClientRealm, *newUser)
	if err != nil {
		// Si utilisateur existe déjà
		if strings.Contains(err.Error(), "User exists") {
			return nil, fmt.Errorf("utilisateur existe déjà")
		}

		return nil, err
	}

	newUser.ID = gocloak.StringP(newUserId)

	err = KeycloakSetPassword(config, clientAccessToken, newUserId, password)
	if err != nil {
		// Suppression de l'utilisateur
		_ = KeycloakDeleteUser(config, clientAccessToken, newUserId)

		return nil, err
	}

	if *newUser.Enabled {
		_, userToken, err := KeycloakLogin(config, *newUser.Username, password)
		if err != nil {
			// Suppression de l'utilisateur
			_ = KeycloakDeleteUser(config, clientAccessToken, newUserId)

			return nil, err
		}

		return userToken, nil
	}

	return nil, nil
}

func KeycloakSetPassword(config models.Config, clientAccessToken string, userId string, password string) error {
	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	return kcClient.SetPassword(ctx, clientAccessToken, userId, config.KeyCloakClientRealm, password, false)
}

func KeycloakDeleteUser(config models.Config, clientAccessToken string, userId string) error {
	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	return kcClient.DeleteUser(ctx, clientAccessToken, config.KeyCloakClientRealm, userId)
}

func KeycloackUserExist(config models.Config, adminAccessToken string, username string) (bool, *gocloak.User, error) {
	if _, err := strconv.Atoi(username); err == nil {
		users, err := KeycloakGetUsers(config, adminAccessToken, gocloak.GetUsersParams{
			Username: gocloak.StringP(username),
			Exact:    gocloak.BoolP(true),
		})

		if err != nil {
			return false, nil, err
		}

		if len(users) <= 0 {
			err := fmt.Errorf(USER_NOT_EXIST)

			return false, nil, err
		}

		return true, users[0], nil
	} else if _, err := mail.ParseAddress(username); err == nil { // Si email
		users, err := KeycloakGetUsers(config, adminAccessToken, gocloak.GetUsersParams{
			Email: gocloak.StringP(username),
			Exact: gocloak.BoolP(true),
		})

		if err != nil {
			return false, nil, err
		}

		if len(users) <= 0 {
			err := fmt.Errorf(USER_NOT_EXIST)

			return false, nil, err
		}

		return true, users[0], nil
	} else if _, err := uuid.FromString(username); err == nil { // Si id
		user, err := KeycloakGetUserByID(config, adminAccessToken, username)
		if err != nil {
			return false, nil, err
		}

		return true, user, nil
	} else {
		err := fmt.Errorf("utilisateur inexistant")

		return false, nil, err
	}
}

func KeycloakGetClient(config models.Config) *gocloak.GoCloak {
	// Connexion au serveur d'authentification keycloak
	return gocloak.NewClient(config.KeyCloakHost, func(gc *gocloak.GoCloak) {
		gc.SetRestyClient(gc.RestyClient().SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		}))
	})
}

func KeycloakIntrospect(config models.Config, accessToken string) (*gocloak.IntroSpectTokenResult, error) {
	// Connexion de l'utilisateur
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
	defer ctxCancelFunc()

	// Connexion au serveur d'authentification keycloak
	kcClient := KeycloakGetClient(config)

	introspectResult, err := kcClient.RetrospectToken(ctx, accessToken, config.KeyCloakClientID, config.KeyCloakClientSecret, config.KeyCloakClientRealm)
	if err != nil {
		return nil, err
	}

	return introspectResult, nil
}
