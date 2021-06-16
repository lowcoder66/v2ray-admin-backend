package controller

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"sort"
	"time"
	"v2ray-admin/backend/auth"
	"v2ray-admin/backend/conf"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/response"
	"v2ray-admin/backend/service"
	"v2ray-admin/backend/util"
)

type UserTrafficRes struct {
	Limit    uint64   `json:"limit" `
	UpLink   uint64   `json:"upLink"`
	DownLink uint64   `json:"downLink"`
	History  []uint64 `json:"history"`
}

type HistoryTrafficRes struct {
	History []uint64 `json:"history"`
}
type GlobalTrafficRes struct {
	UpLink   uint64   `json:"upLink"`
	DownLink uint64   `json:"downLink"`
	History  []uint64 `json:"history"`
}

func GetGlobalTraffic(ctx echo.Context) error {
	queryAndSaveGlobalTraffic()

	up, down := model.GetGlobalTraffic()
	res := GlobalTrafficRes{up, down, userTrafficHistory(0, true, true)}

	return ctx.JSON(http.StatusOK, res)
}
func GetGlobalUpTraffic(ctx echo.Context) error {
	queryAndSaveGlobalTraffic()

	res := HistoryTrafficRes{userTrafficHistory(0, true, false)}
	return ctx.JSON(http.StatusOK, res)
}
func GetGlobalDownTraffic(ctx echo.Context) error {
	queryAndSaveGlobalTraffic()

	res := HistoryTrafficRes{userTrafficHistory(0, false, true)}
	return ctx.JSON(http.StatusOK, res)
}

func UserTraffic(ctx echo.Context) error {
	principal := ctx.Get("principal").(*auth.Principal)
	if &principal == nil {
		return ctx.JSON(http.StatusUnauthorized, response.ErrRes("未获取到用户信息", nil))
	}

	queryAndSaveUserTraffic(principal.Id, principal.Email)

	// current month
	up, down := model.CountUserTrafficOfCurrentMonth(principal.Id)
	res := UserTrafficRes{principal.Limit, up, down, userTrafficHistory(principal.Id, true, true)}

	return ctx.JSON(http.StatusOK, res)
}

func userTrafficHistory(userId uint32, countUp bool, countDown bool) []uint64 {
	history := make(map[string]uint64)

	// history, 12 month ago
	startTime := util.TruncMonth(time.Now().AddDate(0, -12, 0))

	pageNum, size := 1, 10
	page, err := model.FindUserTraffic(userId, startTime, pageNum, size)
	if err != nil {
		log.Panicln(err)
	}

	for {
		if page.Content != nil && len(page.Content.([]model.Traffic)) > 0 {
			records := page.Content.([]model.Traffic)

			for _, r := range records {
				monthKey := r.RecordTime.Format("2006-01")
				up, down := r.UpLink, r.DownLink
				if !countUp {
					up = uint64(0)
				}
				if !countDown {
					down = uint64(0)
				}
				if val, ok := history[monthKey]; ok {
					history[monthKey] = val + up + down
				} else {
					history[monthKey] = up + down
				}
			}

			pageNum += 1
			page, err = model.FindUserTraffic(userId, startTime, pageNum, size)
			if err != nil {
				log.Panicln(err)
			}
		} else {
			break
		}
	}

	// sort history
	monthKeys := make([]string, 0)
	for k := range history {
		monthKeys = append(monthKeys, k)
	}
	historyArr := make([]uint64, 0)
	sort.Strings(monthKeys)
	for _, k := range monthKeys {
		historyArr = append(historyArr, history[k])
	}

	return historyArr
}

func queryAndSaveUserTraffic(userId uint32, email string) {
	// 实时查询
	currUp, currDown := service.QueryUserTraffic(email, true)
	// 保存查询
	if currUp+currDown > 0 {
		traffic, exist := model.GetUserTrafficOfCurrentMonth(userId)
		if exist {
			model.AddTrafficUpAndDown(traffic, currUp, currDown)
		} else { // add
			traffic := model.Traffic{UserId: userId, RecordTime: time.Now(), UpLink: currUp, DownLink: currDown}
			model.AddTraffic(&traffic)
		}
	}
}

func queryAndSaveGlobalTraffic() (uint64, uint64) {
	proxyTag := "proxy"
	if conf.App.V2ray.Tag != "" {
		proxyTag = conf.App.V2ray.Tag
	}
	currUp, currDown := service.QueryGlobalTraffic(true, proxyTag)

	// 保存查询
	if currUp+currDown > 0 {
		traffic, exist := model.GetUserTrafficOfCurrentMonth(0)
		if exist {
			model.AddTrafficUpAndDown(traffic, currUp, currDown)
		} else { // add
			traffic := model.Traffic{RecordTime: time.Now(), UpLink: currUp, DownLink: currDown}
			model.AddTraffic(&traffic)
		}
	}

	return currUp, currDown
}
