package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/v2fly/v2ray-core/v4/infra/conf"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	v2RayJsonReader "github.com/v2fly/v2ray-core/v4/infra/conf/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	c "v2ray-admin/backend/conf"
	"v2ray-admin/backend/model"
	"v2ray-admin/backend/util"
)

type (
	VMessUser struct {
		Id      string `json:"id"`
		AlterId uint32 `json:"alterId"`
		Level   uint32 `json:"level"`
		Email   string `json:"email,omitempty"`
	}

	ServerConfig struct {
		LogConfig       *conf.LogConfig             `json:"log,omitempty"`
		RouterConfig    *conf.RouterConfig          `json:"routing,omitempty"`
		DNSConfig       *conf.DNSConfig             `json:"dns,omitempty"`
		InboundConfigs  []InboundDetourConfig       `json:"inbounds,omitempty"`
		OutboundConfigs []conf.OutboundDetourConfig `json:"outbounds,omitempty"`
		Transport       *conf.TransportConfig       `json:"transport,omitempty"`
		Policy          *conf.PolicyConfig          `json:"policy,omitempty"`
		Api             *conf.APIConfig             `json:"api,omitempty"`
		Stats           *conf.StatsConfig           `json:"stats,omitempty"`
		Reverse         *conf.ReverseConfig         `json:"reverse,omitempty"`
	}

	InboundDetourConfig struct {
		Protocol       string                              `json:"protocol,omitempty"`
		PortRange      interface{}                         `json:"port,omitempty"`
		ListenOn       interface{}                         `json:"listen,omitempty"`
		Settings       *json.RawMessage                    `json:"settings,omitempty"`
		Tag            string                              `json:"tag,omitempty"`
		Allocation     *conf.InboundDetourAllocationConfig `json:"allocate,omitempty"`
		StreamSetting  *conf.StreamConfig                  `json:"streamSettings,omitempty"`
		DomainOverride cfgcommon.StringList                `json:"domainOverride,omitempty"`
		SniffingConfig *conf.SniffingConfig                `json:"sniffing,omitempty"`
	}
)

var defPolicy = &conf.Policy{}
var defHandShake = uint32(4)
var defConnIdle = uint32(300)
var defUpLinkOnly = uint32(1)
var defDownLinkOnly = uint32(1)
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

func GetConfLevelRange(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, getLevelRange())
}

func GetConf(ctx echo.Context) error {
	serverConf := mergeConf()
	return ctx.JSON(http.StatusOK, serverConf)
}

func WriteConfJson() {
	serverConf := mergeConf()
	writeConfJsonToFile(serverConf)
}

func writeConfJsonToFile(conf *ServerConfig) {
	filePath := "/etc/v2ray/config.json"
	if c.App.V2ray.ConfigFile != "" {
		filePath = c.App.V2ray.ConfigFile
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Panicln("打开配置文件出错", err)
	}

	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		log.Panicln("清空文件出错", err)
	}

	write := bufio.NewWriter(file)

	content, err := json.Marshal(conf)
	if err != nil {
		log.Panicln("序列化json出错", err)
	}

	_, err = write.Write(content)
	if err != nil {
		log.Panicln("写入配置文件出错", err)
	}

	err = write.Flush()
	if err != nil {
		log.Panicln("写入配置文件出错", err)
	}
}

func mergeConf() *ServerConfig {
	serverConf := readConf()

	// level range
	levelRange := getLevelRange()
	for ri := levelRange[0]; ri <= levelRange[1]; ri++ {
		if _, ok := serverConf.Policy.Levels[uint32(ri)]; !ok {
			serverConf.Policy.Levels[uint32(ri)] = defPolicy
		}
	}

	for i, in := range serverConf.InboundConfigs {
		// 代理入站设置
		if in.Protocol == "vmess" {
			vMessConf := conf.VMessInboundConfig{}
			if err := json.Unmarshal(*in.Settings, &vMessConf); err != nil {
				log.Panicln(err)
			}

			// 加载用户（启用的用户）
			vMessUsers := make([]json.RawMessage, 0)
			pageNum, size := 1, 10
			page, err := model.FindUserByEnabled(true, pageNum, size)
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

						// user policy levels
						if _, ok := serverConf.Policy.Levels[user.Level]; !ok {
							serverConf.Policy.Levels[user.Level] = defPolicy
						}
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

			vMessConf.Users = vMessUsers

			settingsJson, err := json.Marshal(vMessConf)
			if err != nil {
				log.Panicln(err)
			}
			settingsJsonRaw := json.RawMessage(settingsJson)

			in.Settings = &settingsJsonRaw

			// tag
			tag := "proxy"
			if c.App.V2ray.Tag != "" {
				tag = c.App.V2ray.Tag
			}
			in.Tag = tag

			serverConf.InboundConfigs[i] = in
		}
		continue
	}

	return serverConf
}

func readConf() *ServerConfig {
	configFile := "/resources/v2ray-server-config.json"
	configFilePath := getConfigFilePath(configFile)

	jsonBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Panicln("无法读取配置文件模板", err)
	}
	reader := bytes.NewReader(jsonBytes)

	jsonConfig := &ServerConfig{}
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

func getLevelRange() []int {
	levelRange := []int{1, 10}
	rangeStr := c.App.V2ray.LevelRange
	if &rangeStr != nil {
		arr := strings.Split(rangeStr, "-")
		if &arr[0] != nil {
			i, err := strconv.Atoi(arr[0])
			if err == nil {
				levelRange[0] = i
			}
		}
		if &arr[1] != nil {
			i, err := strconv.Atoi(arr[1])
			if err == nil {
				levelRange[1] = i
			}
		}
	}

	return levelRange
}
