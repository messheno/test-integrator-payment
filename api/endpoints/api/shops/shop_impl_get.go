package shops

import (
	"fmt"
	"spay/models"
	"time"

	"github.com/labstack/echo/v4"
)

type ResShopAPIGetSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Shop models.ShopModel `json:"shop"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Get
// @Summary      	Get shop data
// @Description  	Récuperation des informations de la boutique
// @Tags         	Shops
// @Product       	json
// @response      	200 {object} ResShopAPIGetSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/:id/ [get]
func (s *ShopApiRessource) GetInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		shop, ok := c.Get("SHOP").(*models.ShopModel)
		if !ok {
			err := fmt.Errorf("boutique non valide")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type resData struct {
			Shop models.ShopModel `json:"shop"`
		}

		resp.SetData(resData{
			Shop: *shop,
		})

		return resp.Send(c)
	}
}
