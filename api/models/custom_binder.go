package models

import (
	"github.com/gorilla/schema"
	"github.com/labstack/echo/v4"
)

type CustomBinder struct{}

func (cb *CustomBinder) Bind(i interface{}, c echo.Context) (err error) {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	decoder.ZeroEmpty(false)

	c.Request().ParseForm()
	postParam, _ := c.FormParams()

	return decoder.Decode(i, postParam)
}
