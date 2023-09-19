package users

import (
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

type ResUserAPILoginSuccess struct {
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

type LoginFormData struct {
	Username string `json:"username" form:"username" validate:"required"`              // User name
	Password string `json:"password" form:"password" validate:"required,min=4,max=18"` // Mot de passe
}

// Login
// @Summary      	Login user
// @Description  	Connexion d'un utilisateur
// @Tags         	Users
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData LoginFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResUserAPILoginSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/users/login [post]
func (u *UserApiRessource) Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		// Récuperation des données du formulaire
		data := new(LoginFormData)
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

		// Connexion via service d'authentification
		user, token, err := loginUserDb(*data)
		if err != nil {
			log.Error().Err(err).Msgf("")
			msgErr := checkErrMsg(err)

			return resp.SendError(c, msgErr, models.TransformErr(err))
		}

		// Récuperation des information de l'utilisateur

		type resData struct {
			User         models.UserModel `json:"user"`
			Token        string           `json:"token"`
			RefreshToken string           `json:"refresh_token"`
		}

		resp.SetData(resData{
			User:         *user,
			Token:        token.AccessToken,
			RefreshToken: token.RefreshToken,
		})

		return resp.Send(c)
	}
}

func checkErrMsg(err error) string {
	msgErr := err.Error()
	if strings.Contains(err.Error(), "401 Unauthorized: invalid_grant: Invalid user credentials") {
		msgErr = "username ou mot de passe invalide"
	}

	if strings.Contains(err.Error(), "record not found") {
		msgErr = "compte inexistant"
	}

	return msgErr
}

func loginUserDb(data LoginFormData) (*models.UserModel, *gocloak.JWT, error) {
	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return nil, nil, err
	}

	userKeycloak, token, err := loginKeycloakUser(data)
	if err != nil {
		return nil, nil, err
	}

	user := models.UserModel{}

	result := db.Where("auth_id = ?", gocloak.PString(userKeycloak.ID)).First(&user)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("")

		return nil, nil, result.Error
	}

	return &user, token, nil
}

func loginKeycloakUser(data LoginFormData) (*gocloak.User, *gocloak.JWT, error) {
	config, err := models.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	// Connexion au serveur d'authentification keycloak
	_, kcClientToken, err := utils.KeycloakLoginClient(config)
	if err != nil {
		return nil, nil, err
	}

	// Connexion de l'utilisateur
	kcUserToken, user, errToken := utils.KeycloakGetUserAndToken(config, kcClientToken.AccessToken, data.Username, data.Password)

	if errToken != nil {
		return nil, nil, errToken
	}

	return user, kcUserToken, nil
}
