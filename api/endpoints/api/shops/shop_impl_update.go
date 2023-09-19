package shops

import (
	"encoding/json"
	"fmt"
	"net/http"
	"spay/models"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ResShopAPIUpdateSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Shop models.ShopModel `json:"shop"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type UpdateFormData struct {
	Name        string `json:"name,omitempty" form:"name" xml:"name" validate:"omitempty"`                      // Nom de la boutique
	Description string `json:"description,omitempty" form:"description" xml:"description" validate:"omitempty"` // Une courte description de la boutiuque
	SiteWeb     string `json:"site_web,omitempty" form:"site_web" xml:"site_web" validate:"omitempty"`          // Site web (ex: https://www.maboutique.ci)
	Country     string `json:"country,omitempty" form:"country" xml:"country" validate:"omitempty"`             // Pays (ex: civ)
}

// UpdateInfo
// @Summary      	UpdateInfo shop information
// @Description  	Mise à jour des données de la boutique
// @Tags         	Shops
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData UpdateFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResShopAPIUpdateSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/:id/ [put]
func (s *ShopApiRessource) UpdateInfo() echo.HandlerFunc {
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

		// Récuperation de l'utilisateur à modifier
		updateShop, ok := c.Get("SHOP").(*models.ShopModel)
		if !ok {
			err := fmt.Errorf("boutique non valide")
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation des informations du formulaires
		data := new(UpdateFormData)
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

		err := checkForUpdateShop(claims.Claims, updateShop, data, c)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type resData struct {
			Shop models.ShopModel `json:"shop"`
		}

		resp.SetData(resData{
			Shop: *updateShop,
		})

		return resp.Send(c)
	}
}

func checkForUpdateShop(claims jwt.MapClaims, updateShop *models.ShopModel, formData *UpdateFormData, c echo.Context) error {
	// Récuperation de l'utilisateur
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return err
	}

	result := db.
		Preload("ShopPermissions").
		Where(&loginUser).First(&loginUser)
	if result.Error != nil {
		return result.Error
	}

	if !loginUser.IsGrant(models.USER_MANAGER) && !loginUser.IsShopGrant(updateShop.ID, models.SHOP_MANAGER) {
		err := fmt.Errorf("permission non accordé")
		return err
	}

	// Convert to json
	shopFormJson, err := json.Marshal(formData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(shopFormJson, updateShop)
	if err != nil {
		return err
	}

	// Check
	// Validation du formulaire
	if err := c.Validate(updateShop); err != nil {
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

		return errorsApi
	}

	// Mise à jour des informations utilisateurs
	result = db.Save(updateShop)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates") {
			return fmt.Errorf("impossible d'effectuer la modification, donnée dupliquée detecter")
		}

		return err
	}

	return nil
}
