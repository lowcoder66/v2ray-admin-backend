package controller

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"strconv"
	"time"
	"v2ray-admin/backend/auth"
	"v2ray-admin/backend/cache"
	"v2ray-admin/backend/email"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/response"
	"v2ray-admin/backend/util"
)

var errorCacheKeyPrefix = "icu.lowcoder.va.login.user#"
var resetPswEmailCacheKeyPrefix = "icu.lowcoder.va.reset-psw.user#"
var resetPswEmailSendCacheKeyPrefix = "icu.lowcoder.va.reset-psw.lock#"

type TokenReq struct {
	Username string `json:"username" form:"username" validate:"required"`
	Password string `form:"password" json:"password" validate:"required"`
}
type ResetPasswordEmailReq struct {
	Email string `json:"email" form:"email" validate:"required,email"`
}
type ResetPasswordReq struct {
	Username string `json:"username" form:"username" validate:"required"`
	Password string `form:"password" json:"password" validate:"required,gte=8"`
	Code     string `form:"code" json:"code" validate:"required"`
}

func Principal(ctx echo.Context) error {
	principal := ctx.Get("principal").(*auth.Principal)
	if &principal == nil {
		return ctx.JSON(http.StatusUnauthorized, response.ErrRes("未获取到用户信息", nil))
	}

	return ctx.JSON(http.StatusOK, principal)
}

func NewToken(ctx echo.Context) error {
	req := &TokenReq{}

	// 绑定
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes(err.Error(), nil))
	}
	// 验证
	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("用户名和密码不能为空", nil))
	}

	// 查询用户
	user, exist := model.GetUserByEmail(req.Username)
	if !exist {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("用户名或密码错误", nil))
	}

	// 用户状态
	if !user.Enabled {
		return ctx.JSON(http.StatusForbidden, response.ErrRes("用户未启用", nil))
	}
	if user.Locked {
		return ctx.JSON(http.StatusForbidden, response.ErrRes("用户已锁定，请联系管理员", nil))
	}

	// 登录锁定
	c := cache.Cache
	errCount := 0
	errCountC := c.Get(errorCacheKeyPrefix + req.Username)
	if errCountC != "" {
		i, err := strconv.Atoi(errCountC)
		if err != nil {
			i = 0
		}
		errCount = i
	}
	if errCount >= 4 {
		user.Locked = true
		model.ModifyUser(user, "Locked")
		return ctx.JSON(http.StatusForbidden, response.ErrRes("用户已锁定，请联系管理员", nil))
	}

	// 密码匹配
	if user.Passwd != "" && !util.PasswordMatch(req.Password, user.Passwd) {
		c.Put(errorCacheKeyPrefix+req.Username, strconv.Itoa(errCount+1), 5*60)
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("用户名或密码错误", nil))
	}

	// token
	token := genToken(user)
	resp := &response.AccessToken{Token: token.Value, ReqTime: token.ReqTime, ExpireIn: token.ExpireIn}

	return ctx.JSON(http.StatusOK, resp)
}

func genToken(user *model.User) *model.Token {
	token := model.GetTokenByUserId(user.Id)
	if token != nil {
		if token.ReqTime.Add(time.Second * time.Duration(token.ExpireIn)).After(time.Now()) {
			return token
		} else {
			model.RemoveToken(token.Id)
			token = nil
		}
	}

	token = &model.Token{}
	token.UserId = user.Id
	token.ExpireIn = 24 * 60 * 60
	token.ReqTime = time.Now()
	token.Value = uuid.NewString()

	model.AddToken(token)
	return token
}

func PostPassword(ctx echo.Context) error {
	// send reset password email
	op := ctx.QueryParam("op")
	if op == "send-reset-email" {
		return sendResetPasswordEmail(ctx)
	} else if op == "reset" {
		return resetPassword(ctx)
	}

	return ctx.JSON(http.StatusNotFound, nil)
}

func resetPassword(ctx echo.Context) error {
	req := &ResetPasswordReq{}

	// 绑定
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes(err.Error(), nil))
	}
	// 验证
	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("参数错误", nil))
	}

	// 查询用户
	user, exist := model.GetUserByEmail(req.Username)

	if exist {
		c := cache.Cache
		cacheKey := resetPswEmailCacheKeyPrefix + user.Email

		code := c.Get(cacheKey)
		if code == req.Code {
			user.Passwd = util.EncoderPassword(req.Password)
			model.ModifyUser(user, "Passwd")
			c.Evict(cacheKey)

			return ctx.JSON(http.StatusOK, response.MessageRes("密码已重置，使用新密码登录"))
		} else {
			return ctx.JSON(http.StatusBadRequest, response.ErrRes("随机码不正确或已过期", nil))
		}
	} else {
		return ctx.JSON(http.StatusOK, nil)
	}
}
func sendResetPasswordEmail(ctx echo.Context) error {
	req := &ResetPasswordEmailReq{}

	// 绑定
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes(err.Error(), nil))
	}
	// 验证
	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请输入正确的邮箱账户", nil))
	}

	// 查询用户
	user, exist := model.GetUserByEmail(req.Email)

	// 如果存在，发送邮件；不存在返回成功
	if exist {
		c := cache.Cache
		lockKey, cacheKey := resetPswEmailSendCacheKeyPrefix+user.Email, resetPswEmailCacheKeyPrefix+user.Email
		locked := c.Exist(lockKey)
		if locked {
			return ctx.JSON(http.StatusForbidden, response.ErrRes("请勿频繁发送", nil))
		} else {
			c.Put(lockKey, "lock", 60)

			code := c.Get(cacheKey)
			if code == "" {
				code = util.RandomStr(16)
			}

			err := email.Send("重置您的密码", buildResetPasswordEmailBody(code), "text/html", user.Email, "")
			if err != nil {
				log.Panic(err)
				//return ctx.JSON(http.StatusInternalServerError, response.ErrRes("邮件发送错误", err))
			}

			c.Put(cacheKey, code, 5*60)
			return ctx.JSON(http.StatusOK, response.MessageRes("邮件已发送"))
		}

	} else {
		return ctx.JSON(http.StatusOK, nil)
	}
}

func buildResetPasswordEmailBody(code string) string {
	var buffer bytes.Buffer

	buffer.WriteString("<p>您正在重置密码<p>")
	buffer.WriteString("<p>随机码为:<p>" + code + "</p></p>")
	buffer.WriteString("<p>随机码五分钟内有效</p>")

	return buffer.String()
}
