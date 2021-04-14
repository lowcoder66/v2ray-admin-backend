package model

import (
	"bytes"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"reflect"
	"v2ray-admin/backend/conf"
	"xorm.io/xorm"
	"xorm.io/xorm/caches"
)

var DB *xorm.Engine

func configuredPgParams() string {
	pg := conf.App.Postgres
	t := reflect.TypeOf(pg)
	v := reflect.ValueOf(pg)

	var connectParams bytes.Buffer
	eq, tagKey, sep := "=", "toml", " "
	for i := 0; i < t.NumField(); i++ {
		if !v.Field(i).IsZero() {
			connectParams.WriteString(t.Field(i).Tag.Get(tagKey))
			connectParams.WriteString(eq)
			connectParams.WriteString(fmt.Sprintf("%v", v.Field(i).Interface()))
			connectParams.WriteString(sep)
		}
	}

	return connectParams.String()
}

func init() {
	log.Println("初始化数据库...")

	db, err := xorm.NewEngine("postgres", configuredPgParams())
	if err != nil {
		log.Fatalln("database:", err.Error())
	}

	if err = db.Ping(); err != nil {
		log.Fatalln("database:", err.Error())
	}

	db.SetMaxIdleConns(conf.App.XOrm.MaxIdle)
	db.SetMaxOpenConns(conf.App.XOrm.MaxOpen)
	db.ShowSQL(conf.App.XOrm.ShowSql)
	if conf.App.XOrm.CacheEnable {
		cache := caches.NewLRUCacher(caches.NewMemoryStore(), conf.App.XOrm.CacheCount)
		db.SetDefaultCacher(cache)
	}
	if conf.App.XOrm.Sync {
		err := db.Sync2(new(User), new(Token))
		if err != nil {
			log.Fatalln("database:", err.Error())
		}
	}

	DB = db
	log.Println("数据库初始化完成")
}
