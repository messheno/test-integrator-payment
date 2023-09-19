package utils

import (
	"encoding/json"
	"spay/models"

	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/labstack/echo/v4"
)

func BindValidate(c echo.Context, dataForm interface{}, value interface{}) error {
	// Récuperation des données du formulaire
	if err := c.Bind(dataForm); err != nil {
		return err
	}

	// Validation du formulaire
	if err := c.Validate(dataForm); err != nil {
		errorsApi := models.ResErrorAPI{}

		// Traitement de la reponse de validator
		for _, err := range err.(validator.ValidationErrors) {
			str := stringy.New(err.Field())
			errorsApi = append(errorsApi, models.ErrorAPI{
				Code:    "400",
				Message: str.SnakeCase().ToLower(),
				Data:    err.Tag(),
			})
		}

		return errorsApi
	}

	// Encodage des donnée en json pour faciliter le traitement
	dataJson, err := json.Marshal(dataForm)
	if err != nil {
		return err
	}

	err = json.Unmarshal(dataJson, value)
	if err != nil {
		return err
	}

	return nil
}
