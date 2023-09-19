package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"spay/models"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ResUserAPIUpdateSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		User models.UserModel `json:"user"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type UpdateFormData struct {
	FirstName         string `json:"first_name,omitempty" form:"first_name" validate:"omitempty"`                     // Prénoms
	LastName          string `json:"last_name,omitempty" form:"last_name" validate:"omitempty"`                       // Nom
	PhonePrefix       string `json:"phone_prefix,omitempty" form:"phone_prefix" validate:"omitempty"`                 // Prefix téléphonique (ex: 225)
	PhoneNumber       string `json:"phone_number,omitempty" form:"phone_number" validate:"omitempty"`                 // Numéro de mobile
	Email             string `json:"email,omitempty" form:"email" validate:"omitempty,email"`                         // Adresse e-mail valide
	Country           string `json:"country,omitempty" form:"country" validate:"omitempty"`                           // Pays (ex: civ)
	ApplyForAllSystem bool   `json:"apply_for_all_system,omitempty" form:"apply_for_all_system" validate:"omitempty"` // Appliquer la mise à jour au système complet
}

// UpdateInfo
// @Summary      	UpdateInfo user information
// @Description  	Mise à jour des données de l'utilisateur
// @Tags         	Users
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData UpdateFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResUserAPIUpdateSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/users/:id [put]
func (u *UserApiRessource) UpdateInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		claims, ok := c.Get("JWT_CLAIMS").(models.GrantedData)
		if !ok {
			err := fmt.Errorf("authentification obligatoire")
			resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation de l'utilisateur à modifier
		updateUser, ok := c.Get("USER").(*models.UserModel)
		if !ok {
			err := fmt.Errorf("utilisateur non valide")
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation des informations du formulaires
		data := new(UpdateFormData)
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

		err := checkForUpdateUser(claims.Claims, updateUser, *data, c)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type resData struct {
			User models.UserModel `json:"user"`
		}

		resp.SetData(resData{
			User: *updateUser,
		})

		return resp.Send(c)
	}
}

func checkForUpdateUser(claims jwt.MapClaims, updateUser *models.UserModel, data UpdateFormData, c echo.Context) error {
	// Récuperation de l'utilisateur
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return err
	}

	result := db.
		Preload("ShopPermissions").
		Where(&loginUser).First(&loginUser)
	if result.Error != nil {
		return result.Error
	}

	if !loginUser.IsGrant(models.USER_MANAGER) && (loginUser.IsGrant(models.USER_MERCHANT) && loginUser.AuthId != updateUser.AuthId) {
		err := fmt.Errorf("permission non accordé")
		return err
	}

	// Convert to json
	userFormJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(userFormJson, updateUser)
	if err != nil {
		return err
	}

	// Check
	// Validation du formulaire
	if err := c.Validate(updateUser); err != nil {
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

		return errorsApi
	}

	// Mise à jour des informations utilisateurs
	result = db.Save(updateUser)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("")

		if strings.Contains(result.Error.Error(), "duplicate key value violates") {
			return fmt.Errorf("impossible d'effectuer la modification, donnée dupliquée detecter")
		}

		return err
	}

	return nil
}
