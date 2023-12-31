package services

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

type ResServiceAPICreateSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Service models.ServiceModel `json:"service"`
		User    models.UserModel    `json:"user"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type AddFormData struct {
	Name        string `json:"name" form:"name" xml:"name" validate:"required"`                      // Nom de la boutique
	Description string `json:"description" form:"description" xml:"description" validate:"required"` // Une courte description de la boutiuque
	SiteWeb     string `json:"site_web" form:"site_web" xml:"site_web" validate:"required"`          // Site web (ex: https://www.maboutique.ci)
	Country     string `json:"country" form:"country" xml:"country" validate:"omitempty"`            // Pays (ex: civ)
	AuthId      string `json:"auth_id" form:"auth_id" xml:"auth_id" validate:"omitempty"`            // Id d'authentification de l'admin de la boutique
}

// Add
// @Summary      	Add new service
// @Description  	Création d'un nouvel utilisateur
// @Tags         	Services
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData AddFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResServiceAPICreateSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/services/ [post]
func (s *ServiceApiRessource) Add() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		// Recuperation du claim de connexion
		claims, ok := c.Get("JWT_CLAIMS").(jwt.MapClaims)
		if !ok {
			err := fmt.Errorf("authentification obligatoire")
			// resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation des données du formulaire
		data := new(AddFormData)

		// Decodage de dataJson vers models.UserModel
		newService := models.ServiceModel{}
		err := utils.BindValidate(c, data, &newService)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Création de la boutique
		userAdmin, err := createServiceDb(&newService, data.AuthId, claims)
		if err != nil {
			log.Error().Err(err).Msgf("")
			errMsg := err.Error()
			if strings.Contains(errMsg, "duplicated key not allowed") {
				errMsg = "La boutique existe déjà"
			}

			return resp.SendError(c, errMsg, models.TransformErr(err))
		}

		type resData struct {
			Service models.ServiceModel `json:"service"`
			User    models.UserModel    `json:"user"`
		}

		resp.SetData(resData{
			Service: newService,
			User:    *userAdmin,
		})

		return resp.Send(c)
	}
}

func createServiceDb(newService *models.ServiceModel, authId string, claims jwt.MapClaims) (*models.UserModel, error) {
	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return nil, err
	}

	// Vérification des permissions
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	result := db.
		Preload("ServicePermissions").
		Where(&loginUser).First(&loginUser)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("")
		if strings.Contains(result.Error.Error(), "record not found") {
			return nil, fmt.Errorf("utilisateur non reconnu")
		}

		return nil, result.Error
	}

	userAdmin := models.UserModel{}

	// // Si authId precisé
	if loginUser.IsGrant(models.USER_MANAGER) && authId != "" {
		result := db.Model(&userAdmin).Where("auth_id = ?", authId).First(&userAdmin)
		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		userAdmin = loginUser
	}

	// Récuperation
	newService.Permissions = []models.ServicePermissionModel{
		{
			UserId: userAdmin.ID,
			Role:   models.SERVICE_ADMIN,
		},
	}

	result = db.Create(&newService)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates") {
			return nil, fmt.Errorf("boutique existe déjà")
		}

		return nil, result.Error
	}

	return &userAdmin, nil
}
