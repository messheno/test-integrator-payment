package transactions

import (
	"fmt"
	"spay/models"
	"time"

	"github.com/labstack/echo/v4"
)

type ResServiceAPIGetSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Service models.ServiceModel `json:"service"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Get
// @Summary      	Get service data
// @Description  	Récuperation des informations de la boutique
// @Tags         	Services
// @Product       	json
// @response      	200 {object} ResServiceAPIGetSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/services/:id/ [get]
func (s *TransactionApiRessource) GetInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		service, ok := c.Get("SERVICE").(*models.ServiceModel)
		if !ok {
			err := fmt.Errorf("boutique non valide")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type resData struct {
			Service models.ServiceModel `json:"service"`
		}

		resp.SetData(resData{
			Service: *service,
		})

		return resp.Send(c)
	}
}
