package providers

import (
	"fmt"
	"spay/models"
	"spay/utils"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ResProviderAPICreateSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Provider models.ProviderModel `json:"provider"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type AddFormData struct {
	Name        string `json:"name" form:"name" xml:"name" validate:"required"`                       // Nom du provider
	Description string `json:"description" form:"description" xml:"description" validate:"omitempty"` // Description du provider
}

// Add
// @Summary      	Add new provider
// @Description  	Création d'une nouvelle provider
// @Tags         	Providers
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData AddFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResProviderAPICreateSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/providers/ [post]
func (s *ProviderApiRessource) Add() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		// Recuperation du claim de connexion
		_, ok = c.Get("JWT_CLAIMS").(jwt.MapClaims)
		if !ok {
			err := fmt.Errorf("authentification obligatoire")
			// resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation des données du formulaire
		data := new(AddFormData)

		// Decodage de dataJson vers models.UserModel
		newProvider := models.ProviderModel{}
		err := utils.BindValidate(c, data, &newProvider)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Création de la boutique
		err = createProviderDb(&newProvider)
		if err != nil {
			log.Error().Err(err).Msgf("")
			errMsg := err.Error()
			if strings.Contains(errMsg, "duplicated key not allowed") {
				errMsg = "La boutique existe déjà"
			}

			return resp.SendError(c, errMsg, models.TransformErr(err))
		}

		type resData struct {
			Provider models.ProviderModel `json:"provider"`
		}

		resp.SetData(resData{
			Provider: newProvider,
		})

		return resp.Send(c)
	}
}

func createProviderDb(newProvider *models.ProviderModel) error {
	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return err
	}

	result := db.Create(&newProvider)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates") {
			return fmt.Errorf("boutique existe déjà")
		}

		return result.Error
	}

	return nil
}
