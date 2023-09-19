package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"spay/models"
	"spay/utils"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

const TokenExpMsg = "token invalide ou expiré"

// GrantMid gestion des permissions sur les routes (resource, action)
func GrantMid(roles ...models.UserRole) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
			if !ok {
				resp = models.NewResponseAPI[interface{}]()
			}

			// Chargement de la configuration
			config, err := models.LoadConfig()
			if err != nil {
				return resp.SendError(c, "Une erreur c'est produite lors du chargement du fichier de configuration", models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			// Controle du token
			tokenArr := strings.Split(c.Request().Header.Get("Authorization"), "Bearer")

			if len(tokenArr) < 2 {
				err := fmt.Errorf("champ Header:Authorization incorrect")

				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, err.Error(), models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			token := strings.Trim(tokenArr[1], " ")

			// Connexion au serveur d'authentification keycloak
			kcClient := utils.KeycloakGetClient(config)

			// Configuration du context de connexion
			ctxDecodeToken, ctxDecodeTokenCancelFunc := context.WithTimeout(context.Background(), models.ConnectTimeout)
			defer ctxDecodeTokenCancelFunc()

			kcToken, kcClaims, err := kcClient.DecodeAccessToken(ctxDecodeToken, token, config.KeyCloakClientRealm)
			if err != nil {
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, TokenExpMsg, models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			// Encode claims to json
			claimsJson, err := json.Marshal(kcClaims)
			if err != nil {
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, TokenExpMsg, models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			var claims jwt.MapClaims
			err = json.Unmarshal(claimsJson, &claims)
			if err != nil {
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, TokenExpMsg, models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			// Validité du token
			if !kcToken.Valid {
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, TokenExpMsg, models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			// Claims validation
			if err := kcClaims.Valid(); err != nil {
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, TokenExpMsg, models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			c.Set("JWT_CLAIMS", claims)

			if err := next(c); err != nil {
				return resp.SendError(c, "Une erreur c'est produite", models.ResErrorAPI{
					models.ErrorAPI{
						Code:    "400",
						Message: err.Error(),
						Data:    err,
					},
				})
			}

			return nil
		}
	}
}
