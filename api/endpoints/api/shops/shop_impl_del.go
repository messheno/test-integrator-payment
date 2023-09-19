package shops

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

type ResShopAPIDeleteSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Shop    models.ShopModel `json:"shop"`
		Deleted bool             `json:"deleted"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Delete
// @Summary      	Delete shop
// @Description  	Suppression de la boutique
// @Tags         	Shops
// @Product       	json
// @response      	200 {object} ResShopAPIDeleteSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/:id/ [delete]
func (s *ShopApiRessource) Delete() echo.HandlerFunc {
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

		shop, ok := c.Get("SHOP").(*models.ShopModel)
		if !ok {
			err := fmt.Errorf("boutique non valide")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		err = checkPermission(db, shop, claims.Claims)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		result := db.Delete(shop)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		type resData struct {
			Shop    models.ShopModel `json:"shop"`
			Deleted bool             `json:"deleted"`
		}

		resp.SetData(resData{
			Shop:    *shop,
			Deleted: true,
		})

		return resp.Send(c)
	}
}

func checkPermission(db *gorm.DB, shop *models.ShopModel, claims jwt.MapClaims) error {
	// Récuperation de l'utilisateur
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	result := db.
		Preload("ShopPermissions").
		Where(&loginUser).First(&loginUser)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("")
		if strings.Contains(result.Error.Error(), "record not found") {
			return fmt.Errorf("utilisateur non reconnu")
		}

		return result.Error
	}

	if !loginUser.IsGrant(models.USER_MANAGER) && (loginUser.IsGrant(models.USER_MERCHANT) && !loginUser.IsShopGrant(shop.ID, models.SHOP_MANAGER)) {
		return fmt.Errorf("permission non accordé")
	}

	return nil
}
