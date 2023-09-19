package services

import (
	"fmt"
	"net/http"
	"spay/models"
	"spay/utils"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ResServiceAPIFetchSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Services   []models.ServiceModel  `json:"services"`
		Pagination models.PaginationModel `json:"pagination"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Fetch
// @Summary      	Fetch all service paginate
// @Description  	Récuperation des boutiques paginer
// @Tags         	Service
// @Product       	json
// @response      	200 {object} ResServiceAPIFetchSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/services/ [get]
func (s *ServiceApiRessource) Fetch() echo.HandlerFunc {
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

		// Connexion à la base de donnée
		db, err := models.GetDB()
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation de l'utilisateur
		loginUser := models.UserModel{}
		loginUser.AuthId = claims["sub"].(string)

		result := db.
			Preload("ServicePermissions").
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

		services := []models.ServiceModel{}

		limit, _ := c.Get("LIMIT").(int)
		offset, _ := c.Get("OFFSET").(int)
		query, _ := c.Get("QUERY").(string)
		orders, _ := c.Get("ORDERS").([]string)

		reqDb := db.Model(&models.ServiceModel{})

		var count int64
		err = fetchExec(
			reqDb,
			db,
			loginUser,
			&services,
			fetchParams{
				FilterService: c.QueryParam("filter-service"),
				Orders:        orders,
				Query:         query,
				Limit:         limit,
				Offset:        offset,
			},
			&count,
		)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type dataResponse struct {
			Services   []models.ServiceModel  `json:"services"`
			Pagination models.PaginationModel `json:"pagination"`
		}

		resp.SetData(dataResponse{
			Services: services,
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
		serviceIds := []string{}

		for _, perm := range loginUser.ServicePermissions {
			if loginUser.IsServiceGrant(perm.ServiceId, models.SERVICE_DEV) {
				serviceIds = append(serviceIds, perm.ServiceId)
			}
		}

		reqDb = reqDb.Where("id IN (?)", serviceIds)
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
			Preload("ServicePermissions").
			Where("id = ?", user.ID).
			First(&user)

		if resultUser.Error != nil {
			return nil, resultUser.Error
		}

		// Récuperation de la liste des id
		idsService := []string{}
		for _, permission := range user.ServicePermissions {
			idsService = append(idsService, permission.ServiceId)
		}

		idsService = utils.ArrayUnique(idsService)

		reqDb = reqDb.Where("id IN (?)", idsService)
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
	services *[]models.ServiceModel,
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
	reqDb, err = fetchFilterUser(reqDb, db, params.FilterService)
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
		Find(services)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

type fetchParams struct {
	FilterService string
	Orders        []string
	Query         string
	Limit         int
	Offset        int
}
