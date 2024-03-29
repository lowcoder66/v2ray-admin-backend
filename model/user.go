package model

import (
	"strconv"
	"strings"
)

type (
	User struct {
		Id      uint32 `xorm:"notnull pk autoincr INT(11)" json:"id"`
		UId     string `xorm:"notnull unique VARCHAR(36)" json:"uid"`
		Name    string `xorm:"notnull VARCHAR(255)" json:"name"`
		Email   string `xorm:"notnull unique VARCHAR(255)" json:"email"`
		Level   uint32 `xorm:"notnull INT(11)" json:"level"`
		AlterId uint32 `xorm:"notnull INT(11) default(0)" json:"alterId"`
		Passwd  string `xorm:"notnull VARCHAR(255)" json:"passwd"`
		Phone   string `xorm:"VARCHAR(255)" json:"phone"`
		Enabled bool   `xorm:"notnull Bool default(true)" json:"enabled"`
		Locked  bool   `xorm:"notnull Bool default(false)" json:"locked"`
		Admin   bool   `xorm:"notnull Bool default(false)" json:"admin"`
		Limit   uint64 `xorm:"notnull BIGINT(32) default(0)" json:"limit"`
	}
)

func GetUserByEmail(email string) (*User, bool) {
	mod := &User{}
	exist, _ := DB.Where("email=?", email).Get(mod)
	return mod, exist
}

func ExistUserByEmail(email string) bool {
	exist, _ := DB.Exist(&User{
		Email: email,
	})
	return exist
}

func GetUserById(id uint32) (*User, bool) {
	mod := &User{}
	exist, _ := DB.ID(id).Get(mod)
	return mod, exist
}

func ModifyUser(mod *User, cols ...string) bool {
	sess := DB.NewSession()
	defer sess.Close()
	_ = sess.Begin()

	if _, err := sess.ID(mod.Id).Cols(cols...).Update(mod); err == nil {
		_ = sess.Commit()
		return true
	}

	_ = sess.Rollback()
	return false
}

func AddUser(mod *User) bool {
	sess := DB.NewSession()
	defer sess.Close()
	_ = sess.Begin()

	if _, err := sess.InsertOne(mod); err == nil {
		_ = sess.Commit()
		return true
	}

	_ = sess.Rollback()
	return false
}

func FindUserByKeyword(keyword string, page int, size int) (*Page, error) {
	mods := make([]User, 0)

	sess := DB.NewSession()
	defer sess.Close()

	if keyword != "" {
		lv := "%" + strings.ToUpper(keyword) + "%"
		sess.Where("UPPER(name) like ?", lv).Or("UPPER(email) like ?", lv).Or("phone like ?", lv)
	}

	err := sess.Desc("id").Limit(size, (page-1)*size).Find(&mods)
	count, err := sess.Count(&User{})

	return &Page{mods, page, size, count}, err
}

func FindUserByEnabled(enabled bool, page int, size int) (*Page, error) {
	mods := make([]User, 0)

	sess := DB.NewSession()
	defer sess.Close()

	sess.Where("enabled = ?", enabled)

	err := sess.Desc("id").Limit(size, (page-1)*size).Find(&mods)
	count, err := sess.Count(&User{})

	return &Page{mods, page, size, count}, err
}

func RemoveUser(id int) bool {
	sess := DB.NewSession()
	defer sess.Close()
	_ = sess.Begin()

	if _, err := sess.ID(id).Delete(&User{}); err == nil {
		_ = DB.ClearCacheBean(&User{}, strconv.Itoa(id))
		_ = sess.Commit()
		return true
	}

	_ = sess.Rollback()
	return false
}
