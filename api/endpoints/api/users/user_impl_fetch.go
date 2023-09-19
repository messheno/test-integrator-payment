package users

import (
	"fmt"
	"net/http"
	"spay/models"
	"spay/utils"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ResUserAPIFetchSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Users      []models.UserModel     `json:"users"`
		Pagination models.PaginationModel `json:"pagination"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Fetch
// @Summary      	Fetch all shop paginate
// @Description  	Récuperation des boutiques paginer
// @Tags         	Users
// @Product       	json
// @response      	200 {object} ResUserAPIFetchSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/ [get]
func (u *UserApiRessource) Fetch() echo.HandlerFunc {
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

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation de l'utilisateur
		loginUser := models.UserModel{}
		loginUser.AuthId = claims.Claims["sub"].(string)

		result := db.
			Preload("ShopPermissions").
			Where(&loginUser).First(&loginUser)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")
			if strings.Contains(result.Error.Error(), "record not found") {
				return resp.SendError(c, "Utilisateur non reconnu", models.TransformErr(result.Error))
			}

			return resp.SendError(c, result.Error.Error(), models.TransformErr(result.Error))
		}

		// if !loginUser.IsGrant(models.USER_MANAGER) {
		// 	err := fmt.Errorf("permission non accordé")
		// 	log.Error().Err(err).Msgf("")
		// 	return resp.SendError(c, err.Error(), models.TransformErr(err))
		// }

		users := []models.UserModel{}

		limit, _ := c.Get("LIMIT").(int)
		offset, _ := c.Get("OFFSET").(int)
		query, _ := c.Get("QUERY").(string)
		orders, _ := c.Get("ORDERS").([]string)

		reqDb := db.Model(&models.UserModel{})

		var count int64
		err = fetchExec(
			reqDb,
			db,
			loginUser,
			&users,
			fetchParams{
				FilterShop: c.QueryParam("filter-shop"),
				Orders:     orders,
				Query:      query,
				Limit:      limit,
				Offset:     offset,
			},
			&count,
		)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type dataResponse struct {
			Users      []models.UserModel     `json:"users"`
			Pagination models.PaginationModel `json:"pagination"`
		}

		resp.SetData(dataResponse{
			Users: users,
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

func fetchRestricted(reqDb *gorm.DB, db *gorm.DB, loginUser models.UserModel) (*gorm.DB, error) {
	if loginUser.Role == models.USER_MERCHANT {
		ids := []string{}

		for _, perm := range loginUser.ShopPermissions {
			if loginUser.IsShopGrant(perm.ShopId, models.SHOP_MANAGER) {
				ids = append(ids, perm.ShopId)
			}
		}

		perms := []models.ShopPermissionModel{}

		// Récuperation de toute les permission
		result := db.Model(&models.ShopPermissionModel{}).Where("shop_id IN (?)", ids).Find(&perms)
		if result.Error != nil {
			return nil, result.Error
		}

		usersIds := []string{}
		for _, perm := range perms {
			usersIds = append(usersIds, perm.UserId)
		}

		reqDb = reqDb.Where("id IN (?)", usersIds)
	}

	return reqDb, nil
}

func fetchQuery(reqDb *gorm.DB, query string) *gorm.DB {
	if len(query) > 0 {
		reqDb = reqDb.
			Where("lower(last_name) LIKE ? OR lower(first_name) LIKE ? OR lower(email) LIKE ? OR phone_number LIKE ?", strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"))
	}

	return reqDb
}

func fetchFilterShop(reqDb *gorm.DB, db *gorm.DB, filter string) (*gorm.DB, error) {
	if len(filter) > 0 {
		shop := models.ShopModel{}

		if id, err := uuid.FromString(filter); err != nil {
			return nil, err
		} else {
			shop.ID = id.String()
		}

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return nil, err
		}

		resultShop := db.
			Model(&models.ShopModel{}).
			Preload("Permissions").
			Where("id = ?", shop.ID).
			First(&shop)

		if resultShop.Error != nil {
			log.Error().Err(resultShop.Error).Msgf("")
			return nil, resultShop.Error
		}

		// Récuperation de la liste des id
		idsUser := []string{}
		for _, permission := range shop.Permissions {
			idsUser = append(idsUser, permission.UserId)
		}

		idsUser = utils.ArrayUnique(idsUser)

		reqDb = reqDb.Where("id IN (?)", idsUser)
	}

	return reqDb, nil
}

func fetchOrder(reqDb *gorm.DB, orders []string) *gorm.DB {
	if len(orders) > 0 {
		for _, order := range orders {
			reqDb = reqDb.Order(order)
		}
	}

	return reqDb
}

func fetchExec(
	reqDb *gorm.DB,
	db *gorm.DB,
	loginUser models.UserModel,
	users *[]models.UserModel,
	params fetchParams,
	count *int64,
) error {
	// Restreindre
	reqDb, err := fetchRestricted(reqDb, db, loginUser)
	if err != nil {
		return err
	}

	// Query
	reqDb = fetchQuery(reqDb, params.Query)

	// Filter by ShopId
	reqDb, err = fetchFilterShop(reqDb, db, params.FilterShop)
	if err != nil {
		return err
	}

	result := reqDb.Count(count)
	if result.Error != nil {
		return result.Error
	}

	// Orders
	reqDb = fetchOrder(reqDb, params.Orders)

	result = reqDb.
		Preload("ShopPermissions").
		Preload("ShopPermissions.Shop").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(users)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

type fetchParams struct {
	FilterShop string
	Orders     []string
	Query      string
	Limit      int
	Offset     int
}
