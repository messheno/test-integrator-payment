package services

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

type ResServiceAPIUpdateSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Service models.ServiceModel `json:"service"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type UpdateFormData struct {
	Name        string `json:"name,omitempty" form:"name" xml:"name" validate:"omitempty"`                      // Nom de la boutique
	Description string `json:"description,omitempty" form:"description" xml:"description" validate:"omitempty"` // Une courte description de la boutiuque
	SiteWeb     string `json:"site_web,omitempty" form:"site_web" xml:"site_web" validate:"omitempty"`          // Site web (ex: https://www.maboutique.ci)
	Country     string `json:"country,omitempty" form:"country" xml:"country" validate:"omitempty"`             // Pays (ex: civ)
}

// UpdateInfo
// @Summary      	UpdateInfo service information
// @Description  	Mise à jour des données de la boutique
// @Tags         	Services
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData UpdateFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResServiceAPIUpdateSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/services/:id/ [put]
func (s *ServiceApiRessource) UpdateInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		claims, ok := c.Get("JWT_CLAIMS").(jwt.MapClaims)
		if !ok {
			err := fmt.Errorf("authentification obligatoire")
			resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation de l'utilisateur à modifier
		updateService, ok := c.Get("SERVICE").(*models.ServiceModel)
		if !ok {
			err := fmt.Errorf("boutique non valide")
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

		err := checkForUpdateService(claims, updateService, data, c)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type resData struct {
			Service models.ServiceModel `json:"service"`
		}

		resp.SetData(resData{
			Service: *updateService,
		})

		return resp.Send(c)
	}
}

func checkForUpdateService(claims jwt.MapClaims, updateService *models.ServiceModel, formData *UpdateFormData, c echo.Context) error {
	// Récuperation de l'utilisateur
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return err
	}

	result := db.
		Preload("ServicePermissions").
		Where(&loginUser).First(&loginUser)
	if result.Error != nil {
		return result.Error
	}

	if !loginUser.IsGrant(models.USER_MANAGER) && !loginUser.IsServiceGrant(updateService.ID, models.SERVICE_MANAGER) {
		err := fmt.Errorf("permission non accordé")
		return err
	}

	// Convert to json
	serviceFormJson, err := json.Marshal(formData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(serviceFormJson, updateService)
	if err != nil {
		return err
	}

	// Check
	// Validation du formulaire
	if err := c.Validate(updateService); err != nil {
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
	result = db.Save(updateService)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates") {
			return fmt.Errorf("impossible d'effectuer la modification, donnée dupliquée detecter")
		}

		return err
	}

	return nil
}
