package auth

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"path"
	"strings"
	"time"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/response"
)

type (
	Principal struct {
		Id      int    `json:"id"`
		UId     string `json:"uid"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Level   int    `json:"level"`
		AlterId int    `json:"alterId"`
		Phone   string `json:"phone"`
		Admin   bool   `json:"admin"`
	}
)

func ManagementEndpoint() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			principal := ctx.Get("principal").(*Principal)
			if &principal == nil || !principal.Admin {
				return ctx.JSON(http.StatusForbidden, response.ErrRes("无权访问", nil))
			}

			return next(ctx)
		}
	}
}
func TokenAuth(skipPaths []string) echo.MiddlewareFunc {
	pathSkipper := func(ctx echo.Context) bool {
		for i := 0; i < len(skipPaths); i++ {
			match, _ := path.Match(skipPaths[i], ctx.Request().URL.Path)
			if match {
				return true
			}
		}
		return false
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if pathSkipper(ctx) {
				return next(ctx)
			}

			authHeader := ctx.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return ctx.JSON(http.StatusUnauthorized, response.ErrRes("未认证", nil))
			}

			token := extractToken(authHeader)
			if token == "" {
				return ctx.JSON(http.StatusUnauthorized, response.ErrRes("未认证", nil))
			}

			// 验证
			t := model.GetTokenByValue(token)
			if t == nil {
				return ctx.JSON(http.StatusUnauthorized, response.ErrRes("未认证", nil))
			}
			if t.ReqTime.Add(time.Second * time.Duration(t.ExpireIn)).Before(time.Now()) {
				model.RemoveToken(t.Id)
				return ctx.JSON(http.StatusUnauthorized, response.ErrRes("Token已过期", nil))
			}

			// 注入用户信息
			user, _ := model.GetUserById(t.UserId)
			if user == nil {
				return ctx.JSON(http.StatusForbidden, response.ErrRes("用户不存在", nil))
			}
			if user.Locked || !user.Enabled {
				return ctx.JSON(http.StatusForbidden, response.ErrRes("用户状态异常", nil))
			}

			principal := &Principal{user.Id, user.UId, user.Name,
				user.Email, user.Level, user.AlterId,
				user.Passwd, user.Admin}
			ctx.Set("principal", principal)

			return next(ctx)
		}
	}
}

func extractToken(tok string) string {
	if len(tok) > 6 && strings.ToUpper(tok[0:7]) == "BEARER " {
		return tok[7:]
	}
	return tok
}
