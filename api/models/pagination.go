package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

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

			var limit int = 10
			var offset int = 1
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

func PaginationJson(pageCount int64, limit int64, offset int64, query string) map[string]interface{} {
	return map[string]interface{}{
		"page_count": pageCount,
		"limit":      limit,
		"page":       offset/limit + 1,
		"query":      query,
	}
}

func orderString(sortQuery string) []string {
	sorts := strings.Split(sortQuery, ",")
	orders := []string{}
	for _, sort := range sorts {
		srt := strings.Split(strings.Trim(sort, " "), " ")

		if len(srt) == 2 && (strings.ToLower(srt[1]) == "asc" || strings.ToLower(srt[1]) == "desc") {
			orders = append(orders, fmt.Sprintf("%v %v", string(srt[0]), strings.ToLower(string(srt[1]))))
		}
	}

	return orders
}
