package shops

import (
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
	"gorm.io/gorm"
)

type ResShopPermissionAPIFetchSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		ShopPermissions []models.ShopPermissionModel `json:"shop_permissions"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// FetchPermission
// @Summary      	Fetch shop permissions
// @Description  	Récuperation des permissions de la boutique
// @Tags         	Shop Permissions
// @Product       	json
// @response      	200 {object} ResShopPermissionAPIFetchSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/:id/permissions/ [get]
func (s *ShopApiRessource) FetchPermission() echo.HandlerFunc {
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

		shop, ok := c.Get("SHOP").(*models.ShopModel)
		if !ok {
			err := fmt.Errorf("boutique non valide")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation de l'utilisateur
		err = checkUserPerm(claims.Claims, db, *shop, models.SHOP_DEV)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		permissions := []models.ShopPermissionModel{}

		limit, _ := c.Get("LIMIT").(int)
		offset, _ := c.Get("OFFSET").(int)
		query, _ := c.Get("QUERY").(string)
		orders, _ := c.Get("ORDERS").([]string)

		reqDb := db.Model(&models.ShopPermissionModel{})

		var count int64
		err = fetchPermExec(
			reqDb,
			*shop,
			&permissions,
			fetchParams{
				Orders: orders,
				Query:  query,
				Limit:  limit,
				Offset: offset,
			},
			&count,
		)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type dataResponse struct {
			Permissions []models.ShopPermissionModel `json:"permissions"`
			Pagination  models.PaginationModel       `json:"pagination"`
		}

		resp.SetData(dataResponse{
			Permissions: permissions,
			Pagination: models.PaginationModel{
				Count:  count,
				Limit:  limit,
				Offset: offset,
				Query:  query,
			},
		})

		return resp.Send(c)
	}
}

func checkUserPerm(claims jwt.MapClaims, db *gorm.DB, shop models.ShopModel, role models.ShopRole) error {
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	result := db.
		Preload("ShopPermissions").
		Where(&loginUser).First(&loginUser)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "record not found") {
			return fmt.Errorf("utilisateur inexistant")
		}

		return result.Error
	}

	if !loginUser.IsGrant(models.USER_MANAGER) && !(loginUser.IsGrant(models.USER_MERCHANT) && loginUser.IsShopGrant(shop.ID, role)) {
		err := fmt.Errorf("permission non accordé perms")
		return err
	}

	return nil
}

func fetchPermOrder(reqDb *gorm.DB, orders []string) *gorm.DB {
	if len(orders) > 0 {
		for _, order := range orders {
			reqDb = reqDb.Order(order)
		}
	}

	return reqDb
}

func fetchPermExec(
	reqDb *gorm.DB,
	shop models.ShopModel,
	permissions *[]models.ShopPermissionModel,
	params fetchParams,
	count *int64,
) error {
	result := reqDb.Count(count)
	if result.Error != nil {
		return result.Error
	}

	// Orders
	reqDb = fetchPermOrder(reqDb, params.Orders)

	result = reqDb.
		Preload("Shop").
		Where("shop_id = ?", shop.ID).
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(permissions)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

type ResShopPermissionAPIAddSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		ShopPermission models.ShopPermissionModel `json:"shop_permission"`
		Shop           models.ShopModel           `json:"shop"`
		User           models.UserModel           `json:"user"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type AddPermissionFormData struct {
	UserId string `json:"user_id" form:"user_id" xml:"user_id" validate:"required"` // Identifiant de l'utilisateur
	Role   int    `json:"role" form:"role" xml:"role" validate:"min=0,max=2"`       // Role: 0: DEV, 1: MANAGER, 2: ADMIN
}

// AddUserToShop
// @Summary      	Add user to shop
// @Description  	Récuperation des permissions de la boutique
// @Tags         	Ajout d'un utilisateurs à la boutique
// @Product       	json
// @response      	200 {object} ResShopPermissionAPIAddSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/:id/permissions/add [post]
func (s *ShopApiRessource) AddUserToShop() echo.HandlerFunc {
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

		// Récuperation des données du formulaire
		data := new(AddPermissionFormData)
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

		shop, ok := c.Get("SHOP").(*models.ShopModel)
		if !ok {
			err := fmt.Errorf("boutique non valide")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		permRole := models.SHOP_MANAGER

		if models.ShopRole(data.Role) == models.SHOP_ADMIN {
			permRole = models.SHOP_ADMIN
		}

		// Récuperation de l'utilisateur
		err = checkUserPerm(claims.Claims, db, *shop, permRole)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Get User to add
		userToAdd := models.UserModel{}
		result := db.Model(&userToAdd).Where("id = ? OR auth_id = ?", data.UserId, data.UserId).First(&userToAdd)
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "record not found") {
				err := fmt.Errorf("utilisateur inexistant")
				return resp.SendError(c, err.Error(), models.TransformErr(err))
			}

			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		// Création de la permission
		permission := models.ShopPermissionModel{
			ShopId: shop.ID,
			UserId: userToAdd.ID,
			Role:   models.ShopRole(data.Role),
		}

		result = db.Create(&permission)
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "duplicated key not allowed") {
				err := fmt.Errorf("utilisateur déjà ajouté a cette boutique")
				return resp.SendError(c, err.Error(), models.TransformErr(err))
			}

			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		type resData struct {
			ShopPermission models.ShopPermissionModel `json:"shop_permission"`
			Shop           models.ShopModel           `json:"shop"`
			User           models.UserModel           `json:"user"`
		}

		resp.SetData(resData{
			ShopPermission: permission,
			Shop:           *shop,
			User:           userToAdd,
		})

		return resp.Send(c)
	}
}
