package controller

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/util"
	"v2ray.com/core/infra/conf"
	v2RayJsonReader "v2ray.com/core/infra/conf/json"
)

type (
	VMessUser struct {
		Id      string `json:"id"`
		AlterId uint32 `json:"alterId"`
		Level   uint32 `json:"level"`
		Email   string `json:"email,omitempty"`
	}
)

var defPolicy = &conf.Policy{}
var defHandShake = uint32(4)
var defConnIdle = uint32(300)
var defUpLinkOnly = uint32(2)
var defDownLinkOnly = uint32(5)
var defBufferSize = int32(10240)

func init() {
	defPolicy.Handshake = &defHandShake
	defPolicy.ConnectionIdle = &defConnIdle
	defPolicy.UplinkOnly = &defUpLinkOnly
	defPolicy.DownlinkOnly = &defDownLinkOnly
	defPolicy.BufferSize = &defBufferSize

	defPolicy.StatsUserUplink = true
	defPolicy.StatsUserDownlink = true
}

func GetConf(ctx echo.Context) error {
	serverConf := readConf()

	for i, in := range serverConf.InboundConfigs {
		// 代理入站设置
		if in.Protocol == "vmess" {
			vMessConf := conf.VMessInboundConfig{}
			if err := json.Unmarshal(*in.Settings, &vMessConf); err != nil {
				log.Panicln(err)
			}

			// 加载用户
			vMessUsers := make([]json.RawMessage, 0)
			pageNum, size, query := 1, 10, model.UserQuery{Locked: false, Enabled: true}
			page, err := model.FindUser(query, pageNum, size)
			if err != nil {
				log.Panicln(err)
			}

			for {
				if page.Content != nil && len(page.Content.([]model.User)) > 0 {
					users := page.Content.([]model.User)
					for _, user := range users {
						vMessUser := VMessUser{}
						util.CopyFields(&user, &vMessUser, "Id")
						vMessUser.Id = user.UId

						userJson, err := json.Marshal(vMessUser)
						if err != nil {
							log.Panicln(err)
						}

						userJsonRaw := json.RawMessage(userJson)
						vMessUsers = append(vMessUsers, userJsonRaw)

						// policy levels
						if _, ok := serverConf.Policy.Levels[user.Level]; !ok {
							serverConf.Policy.Levels[user.Level] = defPolicy
						}
					}

					pageNum += pageNum
					page, err = model.FindUser(query, pageNum, size)
					if err != nil {
						log.Panicln(err)
					}
				} else {
					break
				}
			}

			vMessConf.Users = vMessUsers

			settingsJson, err := json.Marshal(vMessConf)
			if err != nil {
				log.Panicln(err)
			}
			settingsJsonRaw := json.RawMessage(settingsJson)

			in.Settings = &settingsJsonRaw
			serverConf.InboundConfigs[i] = in
		}
		continue
	}

	return ctx.JSON(http.StatusOK, serverConf)
}

func readConf() *conf.Config {
	configFile := "/resources/v2ray-server-config.json"
	configFilePath := getConfigFilePath(configFile)

	jsonBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Panicln("无法读取配置文件模板", err)
	}
	reader := bytes.NewReader(jsonBytes)

	jsonConfig := &conf.Config{}
	jsonContent := bytes.NewBuffer(make([]byte, 0, 10240))
	jsonReader := io.TeeReader(&v2RayJsonReader.Reader{
		Reader: reader,
	}, jsonContent)
	decoder := json.NewDecoder(jsonReader)

	if err := decoder.Decode(jsonConfig); err != nil {
		log.Panicln("读取配置文件异常", err)
	}

	return jsonConfig
}

func getConfigFilePath(configFile string) string {
	if workingDir, err := os.Getwd(); err == nil {
		configFile := filepath.Join(workingDir, configFile)
		if fileExists(configFile) {
			return configFile
		}
	}

	return configFile
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}
