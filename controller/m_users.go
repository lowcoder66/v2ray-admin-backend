package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
	"v2ray-admin/backend/auth"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/response"
	"v2ray-admin/backend/service"
	"v2ray-admin/backend/util"
)

type (
	SaveUserReq struct {
		Email   string `json:"email" form:"email" validate:"required"`
		UId     string `json:"uid" form:"uid" validate:"required"`
		Name    string `form:"name" json:"name" validate:"required"`
		Level   uint32 `form:"level" json:"level" validate:"required"`
		AlterId uint32 `form:"alterId" json:"alterId" validate:"required"`
		Phone   string `form:"phone" json:"phone"`
		Enabled bool   `form:"enabled" json:"enabled"`
		Locked  bool   `form:"locked" json:"locked"`
		Admin   bool   `form:"admin" json:"admin"`
	}

	UserResp struct {
		Id      uint32 `json:"id"`
		Email   string `json:"email"`
		UId     string `json:"uid"`
		Name    string `json:"name"`
		Level   uint32 `json:"level"`
		AlterId uint32 `json:"alterId"`
		Phone   string `json:"phone"`
		Enabled bool   `json:"enabled"`
		Locked  bool   `json:"locked"`
		Admin   bool   `json:"admin"`
	}
)

// GET /users
func ListUsers(ctx echo.Context) error {
	// page params
	pageParams := GetParaParams(ctx)
	// search
	page, err := model.FindUserByKeyword(ctx.QueryParam("keyword"), pageParams.Page, pageParams.Size)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes("查询用户列表错误", err))
	}

	users := page.Content.([]model.User)
	respContent := make([]UserResp, len(users))
	for i, user := range users {
		util.CopyFields(&user, &respContent[i])
	}

	return ctx.JSON(http.StatusOK, page)
}

// POST /users
func AddUser(ctx echo.Context) error {
	req := &SaveUserReq{}

	// 绑定
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes(err.Error(), nil))
	}

	// 验证
	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请求参数不正确", nil))
	}

	// 查询用户
	_, exist := model.GetUserByEmail(req.Email)
	if exist {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("邮箱已注册", nil))
	}

	user := &model.User{}
	util.CopyFields(req, user)

	// 远程调用
	if err := service.AddUser(user); err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes("添加用户失败", err))
	}

	// 持久化
	model.AddUser(user)

	return ctx.JSON(http.StatusOK, &response.IDRes{Id: user.Id})
}

// DELETE /users/:id
func DelUser(ctx echo.Context) error {
	userId := ctx.Param("id")

	intId, err := strconv.Atoi(userId)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请求参数不正确", err))
	}

	user, exist := model.GetUserById(intId)
	if !exist {
		return ctx.JSON(http.StatusNotFound, response.ErrRes("用户不存在", nil))
	}

	// 不能删除自己
	principal := ctx.Get("principal").(*auth.Principal)
	if principal.Id == user.Id {
		return ctx.JSON(http.StatusNotFound, response.ErrRes("无法删除自己", nil))
	}

	// 远程调用
	if err := service.RemoveUser(user); err != nil {
		if !strings.Contains(err.Error(), fmt.Sprintf("User %s not found", user.Email)) {
			return ctx.JSON(http.StatusInternalServerError, response.ErrRes("删除用户失败", err))
		}
	}

	model.RemoveUser(intId)

	return ctx.JSON(http.StatusOK, response.MessageRes("操作成功"))
}

// PUT /users/:id
func EditUser(ctx echo.Context) error {
	userId := ctx.Param("id")

	req := &SaveUserReq{}
	// 绑定
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrRes(err.Error(), nil))
	}

	// 验证
	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请求参数不正确", nil))
	}

	intId, err := strconv.Atoi(userId)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请求参数不正确", err))
	}

	user, exist := model.GetUserById(intId)
	if !exist {
		return ctx.JSON(http.StatusNotFound, response.ErrRes("用户不存在", nil))
	}

	util.CopyFields(req, user)

	// 远程调用修改用户
	if err := service.RemoveUser(user); err != nil {
		if !strings.Contains(err.Error(), fmt.Sprintf("User %s not found", user.Email)) {
			return ctx.JSON(http.StatusInternalServerError, response.ErrRes("修改用户失败", err))
		}
	}
	if user.Enabled {
		if err := service.AddUser(user); err != nil {
			return ctx.JSON(http.StatusInternalServerError, response.ErrRes("修改用户失败", err))
		}
	}

	// 不允许修改 email admin
	model.ModifyUser(user, "name", "level", "alter_id", "phone", "enabled", "locked")

	return ctx.JSON(http.StatusOK, response.MessageRes("操作成功"))
}

// GET /users/:id
func GetUser(ctx echo.Context) error {
	userId := ctx.Param("id")

	intId, err := strconv.Atoi(userId)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, response.ErrRes("请求参数不正确", err))
	}

	user, exist := model.GetUserById(intId)
	if !exist {
		return ctx.JSON(http.StatusNotFound, response.ErrRes("用户不存在", nil))
	}

	userResp := &UserResp{}
	util.CopyFields(user, userResp)

	return ctx.JSON(http.StatusOK, userResp)
}
