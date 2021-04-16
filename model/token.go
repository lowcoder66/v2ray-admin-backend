package model

import (
	"strconv"
	"time"
)

type Token struct {
	Id        uint32    `xorm:"notnull pk autoincr INT(11)" json:"id"`
	UserId    uint32    `xorm:"notnull INT(11)" json:"user_id"`
	Value     string    `xorm:"notnull VARCHAR(255)" json:"value"`
	ReqTime   time.Time `xorm:"notnull DateTime" json:"req_time"`
	ExpireIn  uint32    `xorm:"notnull INT(11)" json:"expire_in"`
	DeletedAt time.Time `xorm:"deleted"`
}

func AddToken(token *Token) bool {
	sess := DB.NewSession()
	defer sess.Close()
	_ = sess.Begin()

	if _, err := sess.InsertOne(token); err == nil {
		_ = sess.Commit()
		return true
	}

	_ = sess.Rollback()
	return false
}

func GetTokenByUserId(userId uint32) *Token {
	token := &Token{}
	exist, _ := DB.Where("user_id=?", userId).Get(token)
	if exist {
		return token
	} else {
		return nil
	}
}

func RemoveToken(id uint32) bool {
	sess := DB.NewSession()
	defer sess.Close()
	_ = sess.Begin()

	if _, err := sess.ID(id).Delete(&Token{}); err == nil {
		_ = DB.ClearCacheBean(&Token{}, strconv.Itoa(int(id)))
		_ = sess.Commit()
		return true
	}

	_ = sess.Rollback()
	return false
}

func GetTokenByValue(value string) *Token {
	token := &Token{}
	exist, _ := DB.Where("value=?", value).Get(token)
	if exist {
		return token
	} else {
		return nil
	}
}
