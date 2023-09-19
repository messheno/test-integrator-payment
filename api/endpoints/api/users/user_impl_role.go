package users

import (
	"fmt"
	"net/http"
	"spay/models"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ResUserAPIRoleSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		User    models.UserModel `json:"user"`
		NewRole string           `json:"new_role"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type ChangeRoleFormData struct {
	Role   int    `json:"role" form:"role" validate:"numeric,min=0,max=2"` // 0: Merchant, 1: Manager, 2: Admin
	AuthId string `json:"auth_id" form:"auth_id" validate:"required,uuid"` // Identifiant de connexion de l'utilisateur
}

// ChangeRole
// @Summary      	Change user role
// @Security 		ApiKeyAuth
// @Description  	Changement de la permission de l'utilisateur
// @Tags         	Users
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData ChangeRoleFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResUserAPIRoleSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/users/change-role [post]
func (u *UserApiRessource) ChangeRole() echo.HandlerFunc {
	// Admin
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		claims, ok := c.Get("JWT_CLAIMS").(jwt.MapClaims)
		if !ok {
			err := fmt.Errorf("authentification obligatoire")
			resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation des données du formulaire
		data := new(ChangeRoleFormData)
		if err := c.Bind(data); err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Validation du formulaire
		if err := c.Validate(data); err != nil {
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

		// Récuperation de l'utilisateur actuel
		userForUpdate, err := updateUserRole(claims, *data)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type dataResp struct {
			User    models.UserModel `json:"user"`
			NewRole string           `json:"new_role"`
		}

		resp.SetData(dataResp{
			User:    *userForUpdate,
			NewRole: models.UserRole(data.Role).String(),
		})

		return resp.Send(c)
	}
}

func updateUserRole(claims jwt.MapClaims, data ChangeRoleFormData) (*models.UserModel, error) {
	logeduser := models.UserModel{}
	userForUpdate := models.UserModel{}

	db, err := models.GetDB()
	if err != nil {
		return nil, err
	}

	result := db.Model(&logeduser).Where("auth_id = ?", claims["sub"]).First(&logeduser)
	if result.Error != nil {
		return nil, err
	}

	// Verification du role
	if !logeduser.IsGrant(models.USER_ADMIN) {
		err := fmt.Errorf("permission insuffisante, vous devez être administrateur pour cette opération")
		return nil, err
	}

	// Récuperation de l'utilisateur à modifier
	result = db.Model(&userForUpdate).Where("auth_id = ?", data.AuthId).First(&userForUpdate)
	if result.Error != nil {
		return nil, err
	}

	role := models.UserRole(data.Role)
	userForUpdate.Role = role

	if userForUpdate.ID == logeduser.ID {
		err := fmt.Errorf("interdition de modifier son propre rôle")
		return nil, err
	}

	result = db.Save(&userForUpdate)
	if result.Error != nil {
		return nil, err
	}

	return &userForUpdate, nil
}
