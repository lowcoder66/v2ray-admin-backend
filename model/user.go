package model

import (
	"strconv"
)

type (
	User struct {
		Id      int    `xorm:"notnull pk autoincr INT(11)" json:"id"`
		UId     string `xorm:"notnull unique VARCHAR(36)" json:"uid"`
		Name    string `xorm:"notnull VARCHAR(255)" json:"name"`
		Email   string `xorm:"notnull unique VARCHAR(255)" json:"email"`
		Level   int    `xorm:"notnull INT(11)" json:"level"`
		AlterId int    `xorm:"notnull INT(11)" json:"alterId"`
		Passwd  string `xorm:"notnull VARCHAR(255)" json:"passwd"`
		Phone   string `xorm:"VARCHAR(255)" json:"phone"`
		Enabled bool   `xorm:"notnull Bool default(true)" json:"enabled"`
		Locked  bool   `xorm:"notnull Bool default(false)" json:"locked"`
		Admin   bool   `xorm:"notnull Bool default(false)" json:"admin"`
	}

	UserQuery struct {
		Keyword string `json:"Keyword"`
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

func GetUserById(id int) (*User, bool) {
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

func FindUser(query UserQuery, page int, size int) (*Page, error) {
	mods := make([]User, 0, size)

	sess := DB.NewSession()
	defer sess.Close()

	if &query != nil {
		if &query.Keyword != nil {
			lv := "%" + query.Keyword + "%"
			sess.Where("name like ?", lv).Or("email like ?", lv).Or("phone like ?", lv)
		}
	}

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
