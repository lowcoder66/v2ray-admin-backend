package conf

import (
	"github.com/BurntSushi/toml"
	"log"
)

type AppConfig struct {
	title  string `toml:"title"`
	Server struct {
		Port int `toml:"port"`
	} `toml:"server"`
	Postgres struct {
		DBName                  string `toml:"dbname"`
		User                    string `toml:"user"`
		Password                string `toml:"password"`
		Host                    string `toml:"host"`
		Port                    int    `toml:"port"`
		SSLMode                 string `toml:"sslmode"`
		FallbackApplicationName string `toml:"fallback_application_name"`
		ConnectTimeout          int    `toml:"connect_timeout"`
		SSLCert                 string `toml:"sslcert"`
		SSLKey                  string `toml:"sslkey"`
		SSLRootCert             string `toml:"sslrootcert"`
	} `toml:"postgres"`
	XOrm struct {
		MaxIdle     int  `toml:"max_idle"`
		MaxOpen     int  `toml:"max_open"`
		ShowSql     bool `toml:"show_sql"`
		Sync        bool `toml:"sync"`
		CacheEnable bool `toml:"cache_enable"`
		CacheCount  int  `toml:"cache_count"`
	} `toml:"xorm"`
	Cache struct {
		Manager string `toml:"manager"`
	} `toml:"cache"`
	Redis struct {
		Enable        bool   `toml:"enable"`
		Host          string `toml:"host"`
		Port          int    `toml:"port"`
		PoolMaxIdle   int    `toml:"pool_max_idle"`
		PollMaxActive int    `toml:"poll_max_active"`
	} `toml:"redis"`
	Smtp struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		Username string `toml:"username"`
		Password string `toml:"password"`
		From     string `toml:"from"`
	} `toml:"smtp"`
}

var (
	App           *AppConfig
	defaultConfig = "./conf/conf.toml"
)

func init() {
	log.Println("初始化配置...")
	var err error
	App, err = initConfig()
	if err != nil {
		log.Println("configuration: ", err.Error())
	}
	log.Println("配置初始化完成")
}

func initConfig() (*AppConfig, error) {
	app := &AppConfig{}
	_, err := toml.DecodeFile(defaultConfig, &app)
	if err != nil {
		return nil, err
	}
	return app, nil
}
