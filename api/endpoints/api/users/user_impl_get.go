package users

import (
	"fmt"
	"spay/models"
	"time"

	"github.com/labstack/echo/v4"
)

type ResUserAPIGetSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		User models.UserModel `json:"user"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Get
// @Summary      	Get user data
// @Description  	Récuperation des informations de l'utilisateur
// @Tags         	Users
// @Product       	json
// @response      	200 {object} ResUserAPIGetSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/users/:id [get]
func (u *UserApiRessource) GetInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		user, ok := c.Get("USER").(*models.UserModel)
		if !ok {
			err := fmt.Errorf("utilisateur non valide")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type resData struct {
			User models.UserModel `json:"user"`
		}

		resp.SetData(resData{
			User: *user,
		})

		return resp.Send(c)
	}
}
