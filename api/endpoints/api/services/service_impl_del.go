package services

import (
	"fmt"
	"net/http"
	"spay/models"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ResServiceAPIDeleteSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Service models.ServiceModel `json:"service"`
		Deleted bool                `json:"deleted"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Delete
// @Summary      	Delete service
// @Description  	Suppression de la boutique
// @Tags         	Services
// @Product       	json
// @response      	200 {object} ResServiceAPIDeleteSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/services/:id/ [delete]
func (s *ServiceApiRessource) Delete() echo.HandlerFunc {
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

		service, ok := c.Get("SERVICE").(*models.ServiceModel)
		if !ok {
			err := fmt.Errorf("boutique non valide")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		err = checkPermission(db, service, claims)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		result := db.Delete(service)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		type resData struct {
			Service models.ServiceModel `json:"service"`
			Deleted bool                `json:"deleted"`
		}

		resp.SetData(resData{
			Service: *service,
			Deleted: true,
		})

		return resp.Send(c)
	}
}

func checkPermission(db *gorm.DB, service *models.ServiceModel, claims jwt.MapClaims) error {
	// Récuperation de l'utilisateur
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	result := db.
		Preload("ServicePermissions").
		Where(&loginUser).First(&loginUser)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("")
		if strings.Contains(result.Error.Error(), "record not found") {
			return fmt.Errorf("utilisateur non reconnu")
		}

		return result.Error
	}

	if !loginUser.IsGrant(models.USER_MANAGER) && (loginUser.IsGrant(models.USER_MERCHANT) && !loginUser.IsServiceGrant(service.ID, models.SERVICE_MANAGER)) {
		return fmt.Errorf("permission non accordé")
	}

	return nil
}
