package model

import (
	"strconv"
	"time"
	"v2ray-admin/backend/util"
)

type Traffic struct {
	Id         uint32    `xorm:"notnull pk autoincr INT(11)" json:"id"`
	UserId     uint32    `xorm:"INT(11)" json:"user_id"`
	RecordTime time.Time `xorm:"notnull DateTime" json:"record_time"`
	UpLink     uint64    `xorm:"notnull BIGINT(32)" json:"up_link"`
	DownLink   uint64    `xorm:"notnull BIGINT(32)" json:"down_link"`
}

func AddTraffic(traffic *Traffic) bool {
	sess := DB.NewSession()
	defer sess.Close()
	_ = sess.Begin()

	if _, err := sess.InsertOne(traffic); err == nil {
		_ = sess.Commit()
		return true
	}

	_ = sess.Rollback()
	return false
}
func AddTrafficUpAndDown(mod *Traffic, upLink uint64, downLink uint64) {
	sess := DB.NewSession()
	defer sess.Close()
	_ = sess.Begin()

	mod.UpLink = mod.UpLink + upLink
	mod.DownLink = mod.DownLink + downLink

	if _, err := sess.ID(mod.Id).Cols("up_link", "down_link").Update(mod); err == nil {
		_ = sess.Commit()
	}
}
func GetUserTrafficOfCurrentMonth(userId uint32) (*Traffic, bool) {
	now := time.Now()

	mod := &Traffic{}
	exist, _ := DB.Where("user_id = ?", userId).And("record_time >= ?", util.TruncMonth(now)).And("record_time <= ?", util.CeilMonth(now)).Limit(1, 0).Get(mod)
	return mod, exist
}
func CountUserTrafficOfCurrentMonth(userId uint32) (uint64, uint64) {
	now := time.Now()

	upLink, downLink := uint64(0), uint64(0)
	results, _ := DB.Query("select sum(up_link) as up_link, sum(down_link) as down_link  from traffic where user_id = ? and record_time >= ? and record_time <= ?",
		userId, util.TruncMonth(now), util.CeilMonth(now))

	if len(results) > 0 {
		row := results[0]
		ul, ok := row["up_link"]
		if ok {
			uli, _ := strconv.Atoi(string(ul))
			upLink = uint64(uli)
		}

		dl, ok := row["down_link"]
		if ok {
			dli, _ := strconv.Atoi(string(dl))
			downLink = uint64(dli)
		}

		return upLink, downLink
	}

	return upLink, downLink
}

func GetGlobalTraffic() (uint64, uint64) {
	upLink, downLink := uint64(0), uint64(0)
	results, _ := DB.Query("select sum(up_link) as up_link, sum(down_link) as down_link  from traffic where user_id = 0")

	if len(results) > 0 {
		row := results[0]
		ul, ok := row["up_link"]
		if ok {
			uli, _ := strconv.Atoi(string(ul))
			upLink = uint64(uli)
		}

		dl, ok := row["down_link"]
		if ok {
			dli, _ := strconv.Atoi(string(dl))
			downLink = uint64(dli)
		}

		return upLink, downLink
	}

	return upLink, downLink
}

func FindUserTraffic(userId uint32, startTime time.Time, page int, size int) (*Page, error) {
	mods := make([]Traffic, 0)

	sess := DB.NewSession()
	defer sess.Close()

	sess.Where("user_id = ?", userId).And("record_time >= ?", startTime)
	_ = sess.Desc("record_time").Limit(size, (page-1)*size).Find(&mods)
	count, err := sess.Count(&Traffic{})

	return &Page{mods, page, size, count}, err
}
