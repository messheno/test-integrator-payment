package services

import (
	"fmt"
	"net/http"
	"spay/models"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type ResServiceAPIGetClientSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		ClientId  string `json:"client_id"`
		ClientKey string `json:"client_key"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Get
// @Summary      	Get Client service data
// @Description  	Récuperation des informations client de la boutique
// @Tags         	Services
// @Product       	json
// @response      	200 {object} ResServiceAPIGetClientSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/services/:id/show-client [get]
func (s *ServiceApiRessource) GetClient() echo.HandlerFunc {
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

		// Controle de permission
		// Récuperation de l'utilisateur
		loginUser := models.UserModel{}
		loginUser.AuthId = claims["sub"].(string)

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		result := db.
			Preload("ServicePermissions").
			Where(&loginUser).First(&loginUser)
		if result.Error != nil {
			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		if !loginUser.IsGrant(models.USER_MANAGER) && !loginUser.IsServiceGrant(service.ID, models.SERVICE_MANAGER) {
			err := fmt.Errorf("permission non accordé")
			return err
		}

		type resData struct {
			ClientId  string `json:"client_id"`
			ClientKey string `json:"client_key"`
		}

		resp.SetData(resData{
			ClientId:  service.ClientId,
			ClientKey: service.ClientKey,
		})

		return resp.Send(c)
	}
}

type ResServiceAPIGenClientSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		ClientId  string `json:"client_id"`
		ClientKey string `json:"client_key"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// GenClient
// @Summary      	Regeneration Client service data
// @Description  	Régéneration du client de la boutique
// @Tags         	Services
// @Product       	json
// @response      	200 {object} ResServiceAPIGenClientSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/services/:id/regenerate-client [post]
func (s *ServiceApiRessource) GenClient() echo.HandlerFunc {
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

		// Controle de permission
		// Récuperation de l'utilisateur
		loginUser := models.UserModel{}
		loginUser.AuthId = claims["sub"].(string)

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		result := db.
			Preload("ServicePermissions").
			Where(&loginUser).First(&loginUser)
		if result.Error != nil {
			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		if !loginUser.IsGrant(models.USER_MANAGER) && !loginUser.IsServiceGrant(service.ID, models.SERVICE_MANAGER) {
			err := fmt.Errorf("permission non accordé")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		service.GenerateClient()
		result = db.Save(service)
		if result.Error != nil {
			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		type resData struct {
			ClientId  string `json:"client_id"`
			ClientKey string `json:"client_key"`
		}

		resp.SetData(resData{
			ClientId:  service.ClientId,
			ClientKey: service.ClientKey,
		})

		return resp.Send(c)
	}
}
