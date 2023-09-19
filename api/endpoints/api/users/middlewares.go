package users

import (
	"fmt"
	"net/http"
	"regexp"
	"spay/models"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func (u *UserApiRessource) GetOnMid() func(next echo.HandlerFunc) echo.HandlerFunc {
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

			user, err := getUser(c.Param("id"), claims)
			if err != nil {
				log.Error().Err(err).Msgf("")
				return resp.SendError(c, err.Error(), models.TransformErr(err))
			}

			if user == nil {
				err := fmt.Errorf("utilisateur inexistant")
				return resp.SendError(c, err.Error(), models.TransformErr(err))
			}

			c.Set("USER", user)

			if err := next(c); err != nil {
				err := fmt.Errorf("une erreur c'est produite")
				log.Error().Err(err).Msgf("")
				return resp.SendError(c, err.Error(), models.TransformErr(err))
			}

			return nil
		}
	}
}

func getUser(userId string, claims jwt.MapClaims) (*models.UserModel, error) {
	user := models.UserModel{}

	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return nil, err
	}

	// me
	if userId == "me" {
		id, err := uuid.FromString(claims["sub"].(string))
		if err != nil {
			return nil, err
		}

		db = db.Or("auth_id = ?", id.String())
	} else if id, err := uuid.FromString(userId); err == nil {
		db = db.Or("id = ? OR auth_id = ?", id.String(), id.String())
	} else if regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$").MatchString(userId) {
		db = db.Or("email = ?", userId)
	} else if len(userId) >= 8 {
		db = db.Or("phone_number = ?", userId)
	} else {
		err := fmt.Errorf("identifiant utilisateur invalide doit id ou email ou numero")
		return nil, err
	}

	result := db.
		Preload("ShopPermissions").
		Where(&user).
		First(&user)
	if result.Error != nil {
		return nil, err
	}

	return &user, nil
}
