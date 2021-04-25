package auth

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"path"
	"strings"
	"time"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/response"
	"v2ray-admin/backend/util"
)

type (
	Principal struct {
		Id      uint32 `json:"id"`
		UId     string `json:"uid"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Level   uint32 `json:"level"`
		AlterId uint32 `json:"alterId"`
		Phone   string `json:"phone"`
		Admin   bool   `json:"admin"`
		Limit   uint64 `json:"limit"`
	}
)

func ManagementEndpoint() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			principalInterface := ctx.Get("principal")
			if principalInterface == nil {
				return ctx.JSON(http.StatusUnauthorized, response.ErrRes("未认证", nil))
			}
			principal := principalInterface.(*Principal)
			if !principal.Admin {
				return ctx.JSON(http.StatusForbidden, response.ErrRes("无权访问", nil))
			}

			return next(ctx)
		}
	}
}
func TokenAuth(skipPaths []string) echo.MiddlewareFunc {
	pathSkipper := func(ctx echo.Context) bool {
		if skipPaths == nil {
			return false
		}

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
			// 未启用的也可以登录使用
			//if user.Locked || !user.Enabled {
			if user.Locked {
				return ctx.JSON(http.StatusForbidden, response.ErrRes("用户已被锁定", nil))
			}

			principal := &Principal{}
			util.CopyFields(user, principal)

			ctx.Set("principal", principal)
			ctx.Set("tokenId", t.Id)

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
