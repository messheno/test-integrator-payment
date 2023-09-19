package providers

import (
	"fmt"
	"net/http"
	"spay/models"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ResProviderAPIFetchSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Providers  []models.ProviderModel `json:"providers"`
		Pagination models.PaginationModel `json:"pagination"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

// Fetch
// @Summary      	Fetch all providers paginate
// @Description  	Récuperation des providers paginer
// @Tags         	Providers
// @Product       	json
// @response      	200 {object} ResProviderAPIFetchSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/providers/ [get]
func (s *ProviderApiRessource) Fetch() echo.HandlerFunc {
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

		providers := []models.ProviderModel{}

		limit, _ := c.Get("LIMIT").(int)
		offset, _ := c.Get("OFFSET").(int)
		query, _ := c.Get("QUERY").(string)
		orders, _ := c.Get("ORDERS").([]string)

		reqDb := db.Model(&models.ProviderModel{})

		var count int64
		err = fetchExec(
			reqDb,
			db,
			loginUser,
			&providers,
			fetchParams{
				FilterProvider: c.QueryParam("filter-provider"),
				Orders:         orders,
				Query:          query,
				Limit:          limit,
				Offset:         offset,
			},
			&count,
		)
		if err != nil {
			log.Error().Err(err).Msgf("")
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		type dataResponse struct {
			Providers  []models.ProviderModel `json:"providers"`
			Pagination models.PaginationModel `json:"pagination"`
		}

		resp.SetData(dataResponse{
			Providers: providers,
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

func fetchQuery(reqDb *gorm.DB, query string) *gorm.DB {
	if len(query) > 0 {
		reqDb = reqDb.
			Where("lower(reference_id) LIKE ?", strings.ToLower("%"+query+"%"), strings.ToLower("%"+query+"%"))
	}

	return reqDb
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
	providers *[]models.ProviderModel,
	params fetchParams,
	count *int64,
) error {
	// Query
	reqDb = fetchQuery(reqDb, params.Query)

	result := reqDb.Count(count)
	if result.Error != nil {
		return result.Error
	}

	// Orders
	reqDb = fetchOrder(reqDb, params.Orders)

	result = reqDb.
		// Preload("Transactions").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(providers)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

type fetchParams struct {
	FilterProvider string
	Orders         []string
	Query          string
	Limit          int
	Offset         int
}
