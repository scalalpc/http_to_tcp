package globals

import (
	"fmt"
	"http_to_tcp/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Unknwon/goconfig"
)

var MyConfig *Config

type Config struct {
	SocketHost             string
	SocketPort             int
	HttpHost               string
	HttpPort               int
	CallbackUrl            string
	BufferSize             int
	DeviceMaxConcurrency   int
	OnlineOvertimeSeconds  int
	ReceiveOvertimeSeconds int
	PushOvertimeSeconds    int
	LogLevel               string
	LogMaxLines            int
	LogMaxDays             int
	WebsiteIndex           string
	LogDirPath             string
	AllowOrigin            string
}

func init() {
	MyConfig = &Config{
		SocketHost:             "0.0.0.0",
		SocketPort:             1357,
		HttpHost:               "0.0.0.0",
		HttpPort:               1358,
		CallbackUrl:            "",
		BufferSize:             1024,
		DeviceMaxConcurrency:   1000,
		OnlineOvertimeSeconds:  120,
		ReceiveOvertimeSeconds: 30,
		PushOvertimeSeconds:    10,
		LogLevel:               "informational",
		LogMaxLines:            1000,
		LogMaxDays:             15,
		WebsiteIndex:           "index.html",
		LogDirPath:             strings.Replace(filepath.Join(utils.GetAppDirPath(), "logs", "log.txt"), "\\", "/", -1),
		AllowOrigin:            "*",
	}

	configFilePath := fmt.Sprintf("%s%s", filepath.Base(os.Args[0]), ".ini")
	gocfg, err := goconfig.LoadConfigFile(configFilePath)
	if err == nil {
		socketAddress, err := gocfg.GetValue("app", "socketBindAddress")
		if err != nil {
			fmt.Println(err.Error())
		} else {
			socketAddressFields := strings.Split(socketAddress, ":")
			MyConfig.SocketHost = socketAddressFields[0]
			MyConfig.SocketPort, err = strconv.Atoi(socketAddressFields[1])
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
				return
			}
		}
		httpAddress, err := gocfg.GetValue("app", "httpBindAddress")
		if err != nil {
			fmt.Println(err.Error())
		} else {
			httpAddressFields := strings.Split(httpAddress, ":")
			MyConfig.HttpHost = httpAddressFields[0]
			MyConfig.HttpPort, err = strconv.Atoi(httpAddressFields[1])
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
				return
			}
		}
		callbackUrl, err := gocfg.GetValue("app", "callbackUrl")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.CallbackUrl = strings.TrimSpace(callbackUrl)
		}
		bufferSize, err := gocfg.GetValue("app", "bufferSize")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.BufferSize, err = strconv.Atoi(bufferSize)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
				return
			}
		}
		deviceMaxConcurrency, err := gocfg.GetValue("app", "deviceMaxConcurrency")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.DeviceMaxConcurrency, err = strconv.Atoi(deviceMaxConcurrency)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
				return
			}
		}
		onlineOvertimeSeconds, err := gocfg.GetValue("app", "onlineOvertimeSeconds")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.OnlineOvertimeSeconds, err = strconv.Atoi(onlineOvertimeSeconds)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
				return
			}
		}
		receiveOvertimeSeconds, err := gocfg.GetValue("app", "receiveOvertimeSeconds")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.ReceiveOvertimeSeconds, err = strconv.Atoi(receiveOvertimeSeconds)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
				return
			}
		}
		pushOvertimeSeconds, err := gocfg.GetValue("app", "pushOvertimeSeconds")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.PushOvertimeSeconds, err = strconv.Atoi(pushOvertimeSeconds)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
				return
			}
		}
		logLevel, err := gocfg.GetValue("app", "logLevel")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.LogLevel = logLevel
		}
		logMaxLines, err := gocfg.GetValue("app", "logMaxLines")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.LogMaxLines, err = strconv.Atoi(logMaxLines)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		logMaxDays, err := gocfg.GetValue("app", "logMaxDays")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.LogMaxDays, err = strconv.Atoi(logMaxDays)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
				return
			}
		}
		allowOrigin, err := gocfg.GetValue("app", "allowOrigin")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.AllowOrigin = allowOrigin
		}
		websiteIndex, err := gocfg.GetValue("app", "websiteIndex")
		if err != nil {
			// fmt.Println(err.Error())
		} else {
			MyConfig.WebsiteIndex = websiteIndex
		}
	}
}
