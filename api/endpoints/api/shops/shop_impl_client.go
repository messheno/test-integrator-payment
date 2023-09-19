package shops

import (
	"fmt"
	"net/http"
	"spay/models"
	"time"

	"github.com/labstack/echo/v4"
)

type ResShopAPIGetClientSuccess struct {
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
// @Summary      	Get Client shop data
// @Description  	Récuperation des informations client de la boutique
// @Tags         	Shops
// @Product       	json
// @response      	200 {object} ResShopAPIGetClientSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/:id/show-client [get]
func (s *ShopApiRessource) GetClient() echo.HandlerFunc {
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

		// Controle de permission
		// Récuperation de l'utilisateur
		loginUser := models.UserModel{}
		loginUser.AuthId = claims.Claims["sub"].(string)

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		result := db.
			Preload("ShopPermissions").
			Where(&loginUser).First(&loginUser)
		if result.Error != nil {
			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		if !loginUser.IsGrant(models.USER_MANAGER) && !loginUser.IsShopGrant(shop.ID, models.SHOP_MANAGER) {
			err := fmt.Errorf("permission non accordé")
			return err
		}

		type resData struct {
			ClientId  string `json:"client_id"`
			ClientKey string `json:"client_key"`
		}

		resp.SetData(resData{
			ClientId:  shop.ClientId,
			ClientKey: shop.ClientKey,
		})

		return resp.Send(c)
	}
}

type ResShopAPIGenClientSuccess struct {
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
// @Summary      	Regeneration Client shop data
// @Description  	Régéneration du client de la boutique
// @Tags         	Shops
// @Product       	json
// @response      	200 {object} ResShopAPIGenClientSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/:id/regenerate-client [post]
func (s *ShopApiRessource) GenClient() echo.HandlerFunc {
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

		// Controle de permission
		// Récuperation de l'utilisateur
		loginUser := models.UserModel{}
		loginUser.AuthId = claims.Claims["sub"].(string)

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		result := db.
			Preload("ShopPermissions").
			Where(&loginUser).First(&loginUser)
		if result.Error != nil {
			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		if !loginUser.IsGrant(models.USER_MANAGER) && !loginUser.IsShopGrant(shop.ID, models.SHOP_MANAGER) {
			err := fmt.Errorf("permission non accordé")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		shop.GenerateClient()
		result = db.Save(shop)
		if result.Error != nil {
			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		type resData struct {
			ClientId  string `json:"client_id"`
			ClientKey string `json:"client_key"`
		}

		resp.SetData(resData{
			ClientId:  shop.ClientId,
			ClientKey: shop.ClientKey,
		})

		return resp.Send(c)
	}
}
