package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"v2ray-admin/backend/auth"
	"v2ray-admin/backend/conf"
	"v2ray-admin/backend/controller"
	"v2ray-admin/backend/task"
)

func init() {

}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	// echo engine
	engine := echo.New()

	// validator
	engine.Validator = &CustomValidator{validator: validator.New()}

	// middleware
	engine.Use(middleware.Secure())
	//skipPaths := []string{"/token", "/password", "/configuration"}
	//engine.Use(auth.TokenAuth(skipPaths))

	log.Println("注册路由...")
	addRouters(engine)

	// 启动定时任务
	log.Println("注册定时任务...")
	task.RegisterTasks()

	// 输出配置文件
	log.Println("输出配置文件：" + conf.App.V2ray.ConfigFile)
	controller.WriteConfJson()

	log.Println("启动Echo引擎...")
	err := engine.Start(fmt.Sprintf(":%d", conf.App.Server.Port))
	if err != nil {
		log.Println("echo engine:", err)
	}

}

func addRouters(e *echo.Echo) {
	// public
	e.POST(`/token`, controller.NewToken)
	e.POST(`/password`, controller.PostPassword)
	e.GET(`/configuration`, controller.GetConf)
	e.GET(`/configuration.json`, controller.GetConf)
	e.GET(`/configuration/level-range`, controller.GetConfLevelRange)

	// auth
	ag := e.Group("", auth.TokenAuth(nil))
	ag.GET("/principal", controller.Principal)
	ag.GET("/user-traffic", controller.UserTraffic)
	ag.POST("/users/:id", controller.UserOperate)
	ag.GET("/conf-tpl", controller.GetConfTpl)
	ag.DELETE(`/token`, controller.RevokeCurrentToken)

	// management
	mg := e.Group("/management", auth.TokenAuth(nil), auth.ManagementEndpoint())
	mg.GET("/users", controller.ListUsers)
	mg.POST("/users", controller.AddUser)
	mg.PUT("/users/:id", controller.EditUser)
	mg.DELETE("/users/:id", controller.DelUser)
	mg.GET("/users/:id", controller.GetUser)
	mg.GET("/traffic", controller.GetGlobalTraffic)
	mg.GET("/traffic/up", controller.GetGlobalUpTraffic)
	mg.GET("/traffic/down", controller.GetGlobalDownTraffic)
}
