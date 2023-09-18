package endpoints

import (
	"fmt"
	"net/http"
	"strings"

	"integrator/models"

	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func AttachAPI(server *echo.Echo, db *gorm.DB) {
	apiServer := server.Group("/api")
	{
		// User => /users
		UserAttachAPI(apiServer, db)

		apiServer.POST("/init", InitAppPost(db)).Name = "init"
	}
}

// GrantMid gestion des permissions sur les routes (resource, action)
func GrantMid(roles ...models.UserRole) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
			if !ok {
				resp = models.NewResponseAPI[interface{}]()
			}

			// Récuperation du token
			authorization := c.Request().Header.Get("Authorization")

			if len(authorization) <= 0 {
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, "Champ Header:Authorization manquante", models.TransformErr(fmt.Errorf("champ Header:Authorization manquante")))
			}

			// Split
			tokenAuth := strings.Split(authorization, "Bearer")

			if len(tokenAuth) != 2 {
				resp.SetStatus(http.StatusUnauthorized)
				msg := "champ Header:Authorization manquante"

				return resp.SendError(c, msg, models.TransformErr(fmt.Errorf(msg)))
			}

			fmt.Println("TOKEN:", tokenAuth[1])

			// Controle du token
			token, err := jwt.ParseWithClaims(strings.Trim(tokenAuth[1], " "), &models.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(models.JWT_SECRET), nil
			})
			if err != nil {
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, "Token erreur", models.TransformErr(err))
			}

			if claims, ok := token.Claims.(*models.JwtCustomClaims); ok && token.Valid {
				// Traitement de la permission
				permissionOk := false

				for _, role := range roles {
					if role == models.USER_ADMIN && claims.IsAdmin {
						permissionOk = true
					} else if role == models.USER_MANAGER && claims.IsManager {
						permissionOk = true
					}
				}

				if claims.IsAdmin || claims.IsManager {
					permissionOk = true
				}

				if !permissionOk {
					err := fmt.Errorf("permission non accordé")
					resp.SetStatus(http.StatusUnauthorized)
					return resp.SendError(c, err.Error(), models.TransformErr(err))
				}

				// Sauvegarde des informations de token dans le context
				c.Set("JWT_CLAIMS", claims)

				if err := next(c); err != nil {
					// Retourne une erreur
					// c.Error(err)
					return resp.SendError(c, "Une erreur c'est produite", models.TransformErr(err))
				}

				return nil
			}

			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					resp.SetStatus(http.StatusUnauthorized)
					return resp.SendError(c, "Jeton invalide", models.TransformErr(err))
				} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					resp.SetStatus(http.StatusUnauthorized)
					return resp.SendError(c, "Jeton expiré", models.TransformErr(err))
				} else {
					resp.SetStatus(http.StatusUnauthorized)
					return resp.SendError(c, "Impossible de gérer ce jeton", models.TransformErr(err))
				}
			}

			resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, "Impossible de gérer ce jeton", models.TransformErr(fmt.Errorf("impossible de gérer ce jeton")))
		}
	}
}

// InitAppPost initialisation de l'application
// @Summary      Initialisation de l'application
// @Description  Lancer l'initialisation de l'application
// @Tags         app
// @Router       /api/init [post]
func InitAppPost(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		// Récuperation de la signature serial number
		// Vérification de la licence
		if c.FormValue("licence") != "thesecret" {
			err := fmt.Errorf("licence invalide")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Re-initialisation de la base
		err := models.ResetTable(db)
		if err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		err = models.CreateUpdateTable(db)
		if err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		u := new(models.UserModel)
		if err = c.Bind(u); err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		if err = c.Validate(u); err != nil {
			log.Error().Err(err).Msgf("")

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

			return resp.SendError(c, "Formulaire invalide", models.TransformErr(errorsApi))
		}

		if len(u.Password) <= 0 {
			err = fmt.Errorf("mot de passe obligatoire")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		u.Role = models.USER_ADMIN

		result := db.Create(&u)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		// Géneration du token
		token, err := u.GenerateJWT()
		if err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		resp.SetData(map[string]interface{}{
			"user":  u,
			"token": token,
		})

		return resp.Send(c)
	}
}
