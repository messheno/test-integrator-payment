package transactions

import (
	"fmt"
	"net/http"
	"spay/models"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func (s *TransactionApiRessource) GetOnMid() func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
			if !ok {
				resp = models.NewResponseAPI[interface{}]()
			}

			claims, ok := c.Get("JWT_CLAIMS").(jwt.MapClaims)
			if !ok {
				err := fmt.Errorf("authentification obligatoire")
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, err.Error(), models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			transaction, err := getTransaction(c.Param("id"), claims)
			if err != nil {
				log.Error().Err(err).Msgf("")
				return resp.SendError(c, err.Error(), models.TransformErr(err))
			}

			if transaction == nil {
				err := fmt.Errorf("transaction inexistant")
				return resp.SendError(c, err.Error(), models.TransformErr(err))
			}

			c.Set("TRANSACTION", transaction)

			if err := next(c); err != nil {
				err := fmt.Errorf("une erreur c'est produite")
				log.Error().Err(err).Msgf("")
				return resp.SendError(c, err.Error(), models.TransformErr(err))
			}

			return nil
		}
	}
}

func getTransaction(transactionId string, claims jwt.MapClaims) (*models.TransactionModel, error) {
	transaction := models.TransactionModel{}

	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return nil, err
	}

	if id, err := uuid.FromString(transactionId); err == nil {
		db = db.Or("id = ?", id.String())
	} else {
		err := fmt.Errorf("identifiant boutique invalide")
		return nil, err
	}

	result := db.
		Preload("Service").
		Preload("Provider").
		Where(&transaction).
		First(&transaction)
	if result.Error != nil {
		return nil, err
	}

	return &transaction, nil
}
