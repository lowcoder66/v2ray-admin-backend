package task

import (
	"fmt"
	"github.com/robfig/cron"
	"log"
	"strings"
	"time"
	"v2ray-admin/backend/conf"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/service"
)

func RegisterTasks() {
	c := cron.New()

	_ = c.AddFunc("0 */5 * * * ?", queryAndSaveTotalTraffic)
	_ = c.AddFunc("30 */5 * * * ?", queryAndSaveUsersTraffic)

	//启动计划任务
	c.Start()
}

func queryAndSaveTotalTraffic() {
	proxyTag := "proxy"
	if conf.App.V2ray.Tag != "" {
		proxyTag = conf.App.V2ray.Tag
	}
	currUp, currDown := service.QueryGlobalTraffic(true, proxyTag)

	// 保存查询
	traffic := model.Traffic{RecordTime: time.Now(), UpLink: currUp, DownLink: currDown}
	model.AddTraffic(&traffic)
}

func queryAndSaveUsersTraffic() {
	// 启用的用户
	pageNum, size := 1, 10
	page, err := model.FindUserByEnabled(true, pageNum, size)
	if err != nil {
		log.Panicln(err)
	}

	for {
		if page.Content != nil && len(page.Content.([]model.User)) > 0 {
			users := page.Content.([]model.User)
			for _, user := range users {
				queryAndSaveUserTraffic(user)
			}

			pageNum += 1
			page, err = model.FindUserByEnabled(true, pageNum, size)
			if err != nil {
				log.Panicln(err)
			}
		} else {
			break
		}
	}
}

func queryAndSaveUserTraffic(user model.User) {
	// 实时查询
	currUp, currDown := service.QueryUserTraffic(user.Email, true)
	// 保存查询
	traffic := model.Traffic{UserId: user.Id, RecordTime: time.Now(), UpLink: currUp, DownLink: currDown}
	model.AddTraffic(&traffic)

	// 限额
	up, down := model.GetUserTrafficOfCurrentMonth(user.Id)
	if up+down >= user.Limit {
		user.Enabled = false

		// 远程调用修改用户
		if err := service.RemoveUser(&user); err != nil {
			if !strings.Contains(err.Error(), fmt.Sprintf("User %s not found", user.Email)) {
				log.Panicln(err)
			}
		}

		model.ModifyUser(&user, "enabled")
		log.Printf("disabled user [%s]: exceed the limits\n", user.Email)
	}
}
