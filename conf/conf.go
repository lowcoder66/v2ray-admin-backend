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
	V2ray struct {
		Host       string `toml:"host"`
		Port       int    `toml:"port"`
		Tag        string `toml:"tag"`
		LevelRange string `toml:"level_range"`
	} `toml:"v2ray"`
	ConfTpl struct {
		Address       string `toml:"address" json:"address"`
		Port          int    `toml:"port" json:"port"`
		AlterId       int    `toml:"alter_id" json:"alterId"`
		Security      string `toml:"security" json:"security"`
		Network       string `toml:"network" json:"network"`
		Type          string `toml:"type" json:"type"`
		Host          string `toml:"host" json:"host"`
		Path          string `toml:"path" json:"path"`
		Tls           string `toml:"tls" json:"tls"`
		AllowInsecure string `toml:"allow_insecure" json:"allowInsecure"`
	} `toml:"conf_tpl"`
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
		log.Fatal("configuration: ", err.Error())
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
