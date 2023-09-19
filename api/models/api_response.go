package models

import (
	"fmt"
	"net/http"
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
