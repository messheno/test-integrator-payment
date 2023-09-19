package shops

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

type ResShopAPICreateSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Shop models.ShopModel `json:"shop"`
		User models.UserModel `json:"user"`
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
// @Summary      	Add new shop
// @Description  	Création d'un nouvel utilisateur
// @Tags         	Shops
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData AddFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResShopAPICreateSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/ [post]
func (s *ShopApiRessource) Add() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		// Recuperation du claim de connexion
		claims, ok := c.Get("JWT_CLAIMS").(models.GrantedData)
		if !ok {
			err := fmt.Errorf("authentification obligatoire")
			// resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation des données du formulaire
		data := new(AddFormData)

		// Decodage de dataJson vers models.UserModel
		newShop := models.ShopModel{}
		err := utils.BindValidate(c, data, &newShop)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Création de la boutique
		userAdmin, err := createShopDb(&newShop, data.AuthId, claims.Claims)
		if err != nil {
			log.Error().Err(err).Msgf("")
			errMsg := err.Error()
			if strings.Contains(errMsg, "duplicated key not allowed") {
				errMsg = "La boutique existe déjà"
			}

			return resp.SendError(c, errMsg, models.TransformErr(err))
		}

		type resData struct {
			Shop models.ShopModel `json:"shop"`
			User models.UserModel `json:"user"`
		}

		resp.SetData(resData{
			Shop: newShop,
			User: *userAdmin,
		})

		return resp.Send(c)
	}
}

func createShopDb(newShop *models.ShopModel, authId string, claims jwt.MapClaims) (*models.UserModel, error) {
	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return nil, err
	}

	// Vérification des permissions
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	result := db.
		Preload("ShopPermissions").
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
	newShop.Permissions = []models.ShopPermissionModel{
		{
			UserId: userAdmin.ID,
			Role:   models.SHOP_ADMIN,
		},
	}

	result = db.Create(&newShop)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates") {
			return nil, fmt.Errorf("boutique existe déjà")
		}

		return nil, result.Error
	}

	return &userAdmin, nil
}
