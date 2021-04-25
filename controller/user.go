package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"v2ray-admin/backend/auth"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/response"
	"v2ray-admin/backend/util"
)

type (
	ModifyUserInfoReq struct {
		Name  string `form:"name" json:"name" validate:"required"`
		Phone string `form:"phone" json:"phone"`
	}

	ModifyPasswordReq struct {
		Current  string `json:"current" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
)

// POST /users/:id?op
func UserOperate(ctx echo.Context) error {
	userId := ctx.Param("id")
	intId, err := strconv.Atoi(userId)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请求参数不正确", err))
	}

	principal := ctx.Get("principal").(*auth.Principal)
	if &principal == nil {
		return ctx.JSON(http.StatusUnauthorized, response.ErrRes("未获取到用户信息", nil))
	}

	if uint32(intId) != principal.Id {
		return ctx.JSON(http.StatusForbidden, response.ErrRes("不允许访问", nil))
	}

	user, exist := model.GetUserById(uint32(intId))
	if !exist {
		return ctx.JSON(http.StatusNotFound, response.ErrRes("用户不存在", nil))
	}

	op := ctx.QueryParam("op")
	if op == "modify-user-info" {
		return modifyUserInfo(ctx, user)
	} else if op == "modify-password" {
		return modifyPassword(ctx, user)
	}

	return ctx.JSON(http.StatusNotFound, nil)
}

func modifyUserInfo(ctx echo.Context, user *model.User) error {
	req := &ModifyUserInfoReq{}

	// 绑定
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes(err.Error(), nil))
	}

	// 验证
	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请求参数不正确", nil))
	}

	util.CopyFields(req, user)
	model.ModifyUser(user, "name", "phone")

	return ctx.JSON(http.StatusOK, response.MessageRes("操作成功"))
}

func modifyPassword(ctx echo.Context, user *model.User) error {
	req := &ModifyPasswordReq{}

	// 绑定
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes(err.Error(), nil))
	}

	// 验证
	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请求参数不正确", nil))
	}

	if !util.PasswordMatch(req.Current, user.Passwd) {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("原密码不正确", nil))
	}

	user.Passwd = util.EncoderPassword(req.Password)
	model.ModifyUser(user, "passwd")

	return ctx.JSON(http.StatusOK, response.MessageRes("操作成功"))
}
