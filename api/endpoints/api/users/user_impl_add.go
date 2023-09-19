package users

import (
	"encoding/json"
	"fmt"
	"spay/models"
	"spay/utils"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v12"
	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ResUserAPICreateSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		User         models.UserModel `json:"user"`
		Token        string           `json:"token"`
		RefreshToken string           `json:"refresh_token"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type AddFormData struct {
	FirstName            string `json:"first_name" form:"first_name" validate:"required"`                                        // Prénoms
	LastName             string `json:"last_name" form:"last_name" validate:"required"`                                          // Nom
	PhonePrefix          string `json:"phone_prefix" form:"phone_prefix" validate:"required"`                                    // Prefix téléphonique (ex: 225)
	PhoneNumber          string `json:"phone_number" form:"phone_number" validate:"required"`                                    // Numéro de mobile
	Email                string `json:"email" form:"email" validate:"omitempty,email"`                                           // Adresse e-mail valide
	Country              string `json:"country" form:"country" validate:"omitempty"`                                             // Pays (ex: civ)
	Password             string `json:"password" form:"password" validate:"required,min=4,max=18"`                               // Mot de passe
	PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation" validate:"required,eqfield=Password"` // Confirmation du mot de passe
}

// Add
// @Summary      	Add new user
// @Description  	Création d'un nouvel utilisateur
// @Tags         	Users
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData AddFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResUserAPICreateSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/users/ [post]
func (u *UserApiRessource) Add() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		// Récuperation des données du formulaire
		data := new(AddFormData)
		if err := c.Bind(data); err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Validation du formulaire
		if err := c.Validate(data); err != nil {
			log.Error().Err(err).Msgf("")

			errorsApi := models.ResErrorAPI{}

			// Traitement de la reponse de validator
			for _, err := range err.(validator.ValidationErrors) {
				str := stringy.New(err.Field())
				errorsApi = append(errorsApi, models.ErrorAPI{
					Code:    "400",
					Message: str.SnakeCase().ToLower(),
					Data:    err.Tag(),
				})
			}

			return resp.SendError(c, "Formulaire invalide", models.TransformErr(errorsApi))
		}

		// Encodage des donnée en json pour faciliter le traitement
		dataJson, err := json.Marshal(data)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Decodage de dataJson vers models.UserModel
		newUser := models.UserModel{}
		err = json.Unmarshal(dataJson, &newUser)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Création de l'utilisateur
		token, err := createUserDb(&newUser, data.Password)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type resData struct {
			User         models.UserModel `json:"user"`
			Token        string           `json:"token"`
			RefreshToken string           `json:"refresh_token"`
		}

		resp.SetData(resData{
			User:         newUser,
			Token:        token.AccessToken,
			RefreshToken: token.RefreshToken,
		})

		return resp.Send(c)
	}
}

func createUserDb(newUser *models.UserModel, password string) (*gocloak.JWT, error) {
	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return nil, err
	}

	user, token, err := registerKeycloakUser(newUser, password)
	if err != nil {
		return nil, err
	}

	newUser.AuthId = gocloak.PString(user.ID)

	result := db.Create(&newUser)
	if result.Error != nil {
		// Suppression de l'utilisateur créer dans keycloak
		delKeycloakUser()

		log.Error().Err(result.Error).Msgf("")

		if strings.Contains(result.Error.Error(), "duplicate key value violates") {
			return nil, fmt.Errorf("utilisateur existe déjà")
		}

		return nil, result.Error
	}

	return token, nil
}

func registerKeycloakUser(newUser *models.UserModel, password string) (*gocloak.User, *gocloak.JWT, error) {
	config, err := models.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	// Connexion en tant qu'admin
	_, kcClientToken, err := utils.KeycloakLoginClient(config)
	if err != nil {
		return nil, nil, err
	}

	// Configuration de l'utilisateur
	newKcUser := gocloak.User{
		FirstName: gocloak.StringP(newUser.FirstName),
		LastName:  gocloak.StringP(newUser.LastName),
		Email:     gocloak.StringP(newUser.Email),
		Enabled:   gocloak.BoolP(true),
		Username:  gocloak.StringP(fmt.Sprintf("+%v%v", newUser.PhonePrefix, newUser.PhoneNumber)),
		Attributes: &map[string][]string{
			"prefix_phone_number": {
				newUser.PhonePrefix,
			},

			"phone_number": {
				newUser.PhoneNumber,
			},
		},
	}

	// Création d'un nouvel utilisateur
	userToken, err := utils.KeycloakNewUser(config, kcClientToken.AccessToken, &newKcUser, password)
	if err != nil {
		return nil, nil, err
	}

	return &newKcUser, userToken, nil
}

func delKeycloakUser() error {
	return nil
}
