package controller

import (
	"github.com/labstack/echo/v4"
	"strconv"
)

type PageParams struct {
	Page int
	Size int
}

func GetParaParams(ctx echo.Context) *PageParams {
	params := &PageParams{1, 10}

	page := ctx.QueryParam("page")
	if page != "" {
		pageNum, err := strconv.Atoi(page)
		if err != nil || pageNum < 0 {
			pageNum = 1
		}
		params.Page = pageNum
	}

	size := ctx.QueryParam("size")
	if size != "" {
		sizeNum, err := strconv.Atoi(size)
		if err != nil || sizeNum < 0 {
			sizeNum = 10
		}
		params.Size = sizeNum
	}

	return params
}
