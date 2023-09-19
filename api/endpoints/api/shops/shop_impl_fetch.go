package shops

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

type ResShopAPIFetchSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Shops      []models.ShopModel     `json:"shops"`
		Pagination models.PaginationModel `json:"pagination"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Fetch
// @Summary      	Fetch all shop paginate
// @Description  	Récuperation des boutiques paginer
// @Tags         	Shops
// @Product       	json
// @response      	200 {object} ResShopAPIFetchSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/shops/ [get]
func (s *ShopApiRessource) Fetch() echo.HandlerFunc {
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

		shops := []models.ShopModel{}

		limit, _ := c.Get("LIMIT").(int)
		offset, _ := c.Get("OFFSET").(int)
		query, _ := c.Get("QUERY").(string)
		orders, _ := c.Get("ORDERS").([]string)

		reqDb := db.Model(&models.ShopModel{})

		var count int64
		err = fetchExec(
			reqDb,
			db,
			loginUser,
			&shops,
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
			Shops      []models.ShopModel     `json:"shops"`
			Pagination models.PaginationModel `json:"pagination"`
		}

		resp.SetData(dataResponse{
			Shops: shops,
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
		shopIds := []string{}

		for _, perm := range loginUser.ShopPermissions {
			if loginUser.IsShopGrant(perm.ShopId, models.SHOP_DEV) {
				shopIds = append(shopIds, perm.ShopId)
			}
		}

		reqDb = reqDb.Where("id IN (?)", shopIds)
	}

	return reqDb, nil
}

func fetchQuery(reqDb *gorm.DB, query string) *gorm.DB {
	if len(query) > 0 {
		reqDb = reqDb.
			Where("lower(name) LIKE ? OR lower(name_slug) LIKE ? OR lower(description) LIKE ? OR lower(site_web) LIKE ? OR lower(country) LIKE ?", strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"))
	}

	return reqDb
}

func fetchFilterUser(reqDb *gorm.DB, db *gorm.DB, filter string) (*gorm.DB, error) {
	if len(filter) > 0 {
		user := models.UserModel{}

		if id, err := uuid.FromString(filter); err != nil {
			return nil, err
		} else {
			user.ID = id.String()
		}

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return nil, err
		}

		resultUser := db.
			Model(&models.UserModel{}).
			Preload("ShopPermissions").
			Where("id = ?", user.ID).
			First(&user)

		if resultUser.Error != nil {
			return nil, resultUser.Error
		}

		// Récuperation de la liste des id
		idsShop := []string{}
		for _, permission := range user.ShopPermissions {
			idsShop = append(idsShop, permission.ShopId)
		}

		idsShop = utils.ArrayUnique(idsShop)

		reqDb = reqDb.Where("id IN (?)", idsShop)
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
	shops *[]models.ShopModel,
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

	// Filter by UserId
	reqDb, err = fetchFilterUser(reqDb, db, params.FilterShop)
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
		Preload("Permissions").
		Preload("Permissions.User").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(shops)
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
