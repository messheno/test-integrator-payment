package transactions

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

type ResTransactionAPIFetchSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Transactions []models.TransactionModel `json:"transactions"`
		Pagination   models.PaginationModel    `json:"pagination"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Fetch
// @Summary      	Fetch all transaction paginate
// @Description  	Récuperation des transactions paginer
// @Tags         	Transactions
// @Product       	json
// @response      	200 {object} ResTransactionAPIFetchSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/transactions/ [get]
func (s *TransactionApiRessource) Fetch() echo.HandlerFunc {
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

		transactions := []models.TransactionModel{}

		limit, _ := c.Get("LIMIT").(int)
		offset, _ := c.Get("OFFSET").(int)
		query, _ := c.Get("QUERY").(string)
		orders, _ := c.Get("ORDERS").([]string)

		reqDb := db.Model(&models.TransactionModel{})

		var count int64
		err = fetchExec(
			reqDb,
			db,
			loginUser,
			&transactions,
			fetchParams{
				FilterTransaction: c.QueryParam("filter-transaction"),
				Orders:            orders,
				Query:             query,
				Limit:             limit,
				Offset:            offset,
			},
			&count,
		)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type dataResponse struct {
			Transactions []models.TransactionModel `json:"transactions"`
			Pagination   models.PaginationModel    `json:"pagination"`
		}

		resp.SetData(dataResponse{
			Transactions: transactions,
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

		reqDb = reqDb.Where("service_id IN (?)", serviceIds)
	}

	return reqDb, nil
}

func fetchQuery(reqDb *gorm.DB, query string) *gorm.DB {
	if len(query) > 0 {
		reqDb = reqDb.
			Where("lower(reference_id) LIKE ?", strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"))
	}

	return reqDb
}

func fetchFilterTransaction(reqDb *gorm.DB, db *gorm.DB, filter string) (*gorm.DB, error) {
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

		reqDb = reqDb.Where("service_id IN (?)", idsService)
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
	transactions *[]models.TransactionModel,
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
	reqDb, err = fetchFilterTransaction(reqDb, db, params.FilterTransaction)
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
		Preload("Provider").
		Preload("Service").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(transactions)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

type fetchParams struct {
	FilterTransaction string
	Orders            []string
	Query             string
	Limit             int
	Offset            int
}
