package models

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type ResFailure struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data ResErrorAPI `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type ResErrorAPI []ErrorAPI

func (m ResErrorAPI) Error() string {
	errMsg := ""
	for _, err := range m {
		errMsg += err.Message
	}

	return errMsg
}

type ResponseAPI[T any] struct {
	UrlPath string `json:"url_path"`
	Status  int    `json:"status"`
	Message string `json:"message"`

	IsError bool `json:"is_error"`

	Data T `json:"data"`

	status int

	RequestDate time.Time `json:"request_date"`

	startTime   time.Time
	TimeElapsed string `json:"time_elapsed"`
}

type ErrorAPI struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewResponseAPI[T any]() *ResponseAPI[T] {
	return &ResponseAPI[T]{
		startTime: time.Now(),
	}
}

func (resp *ResponseAPI[T]) SetStatus(status int) {
	resp.status = status
}

func (resp *ResponseAPI[T]) SetData(data T) {
	resp.IsError = false
	resp.Data = data
}

func (resp *ResponseAPI[T]) SetErrors(errors T) {
	// Vérification du type
	resp.IsError = true
	resp.Data = errors
}

func (resp *ResponseAPI[T]) Send(c echo.Context) error {
	resp.UrlPath = c.Request().URL.Path
	resp.Status = http.StatusOK
	resp.TimeElapsed = time.Since(resp.startTime).String()

	if resp.status != http.StatusOK && resp.status > 0 {
		resp.Status = resp.status
	}

	resp.RequestDate = time.Now()

	return c.JSON(resp.Status, resp)
}

func (resp *ResponseAPI[T]) SendError(c echo.Context, message string, errors T) error {
	resp.Message = message
	resp.SetErrors(errors)

	return resp.Send(c)
}

// Process is the middleware function.
func ResponseAPIMid(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp := NewResponseAPI[interface{}]()
		c.Set("RESP", resp)

		return next(c)
	}
}

// TransformErr transformation des erreurs en ResErrorAPI
func TransformErr(err error) ResErrorAPI {
	if fmt.Sprintf("%T", err) == "models.ResErrorAPI" {
		return err.(ResErrorAPI)
	}

	return ResErrorAPI{
		ErrorAPI{
			Code:    "400",
			Message: err.Error(),
			Data:    err,
		},
	}
}

// PaginationModel n
type PaginationModel struct {
	Count     int64    `json:"count"`
	Sorts     []string `json:"sorts,omitempty"`
	Page      int      `json:"page,omitempty"`
	PageCount int      `json:"page_count,omitempty"`
	Query     string   `json:"query,omitempty"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset,omitempty"`
	// Contents  interface{} `json:"contents"`
}

// PaginationMid middleware de gestion de pagination
func PaginationMid() func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			resp, ok := c.Get("RESP").(*ResponseAPI[interface{}])
			if !ok {
				resp = NewResponseAPI[interface{}]()
			}

			var limit = 10
			var offset = 1
			query := c.QueryParam("query")

			// Limit
			if c.QueryParam("limit") != "" {
				limit, _ = strconv.Atoi(c.QueryParam("limit"))
			}

			// Page
			if c.QueryParam("page") != "" {
				offset, _ = strconv.Atoi(c.QueryParam("page"))
			}

			// Sorts
			orders := orderString(c.QueryParam("sorts"))

			if offset <= 0 {
				offset = 1
			}

			offset = limit * (offset - 1)

			c.Set("LIMIT", limit)
			c.Set("OFFSET", offset)
			c.Set("QUERY", query)
			c.Set("ORDERS", orders)

			if err := next(c); err != nil {
				return resp.SendError(c, "Une erreur c'est produite", []ErrorAPI{
					{
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

func orderString(sortQuery string) []string {
	sorts := strings.Split(sortQuery, ",")
	var orders []string
	for _, sort := range sorts {
		srt := strings.Split(strings.Trim(sort, " "), " ")

		if len(srt) == 2 && (strings.ToLower(srt[1]) == "asc" || strings.ToLower(srt[1]) == "desc") {
			orders = append(orders, fmt.Sprintf("%v %v", srt[0], strings.ToLower(srt[1])))
		}
	}

	if len(orders) <= 0 {
		orders = []string{}
	}

	return orders
}
