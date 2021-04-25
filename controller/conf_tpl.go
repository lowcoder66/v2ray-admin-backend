package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"v2ray-admin/backend/conf"
)

// Get /conf-tpl
func GetConfTpl(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, conf.App.ConfTpl)
}
