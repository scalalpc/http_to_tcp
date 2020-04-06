package utils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"github.com/astaxie/beego"
)

func ConvertToLogLevel(levelName string) (level int) {
	switch strings.ToLower(levelName) {
	case "emergency":
		level = beego.LevelEmergency
	case "alert":
		level = beego.LevelAlert
	case "critical":
		level = beego.LevelCritical
	case "error":
		level = beego.LevelError
	case "warning":
		level = beego.LevelWarning
	case "notice":
		level = beego.LevelNotice
	case "informational":
		level = beego.LevelInformational
	case "debug":
		level = beego.LevelDebug
	default:
		level = beego.LevelWarning
	}
	return
}

func ParseKeyValueWithMark(src, mark string, includeOnlyKey bool) (ok bool, key string, value string) {
	fields := strings.Split(src, mark)
	if !includeOnlyKey && len(fields) < 2 {
		return
	}
	key = strings.TrimSpace(fields[0])
	if len(fields) > 1 {
		value = strings.TrimSpace(fields[1])
	} else if includeOnlyKey {
		value = ""
	}
	ok = true
	return
}

func GetAppDirPath() string {
	fileInfo, _ := exec.LookPath(os.Args[0])
	absPath, _ := filepath.Abs(fileInfo)
	return filepath.Dir(absPath)
}

func IsLittleEndian() bool {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return (b == 0x04)
}

func BcdToString(bcd []byte) (str string) {
	var tmp string
	for i := 0; i < len(bcd); i++ {
		tmp += strconv.FormatUint(uint64(bcd[i]), 16)
		if len(tmp) == 1 {
			tmp = "0" + tmp
		}
		str += tmp
	}
	return
}

func BcdStrToBytes(bcdStr string) (byteArr []byte) {
	if len(bcdStr)%2 == 1 {
		bcdStr = "0" + bcdStr
	}

	byteArr = make([]byte, len(bcdStr)/2)
	var hi, lo uint64
	var err error
	for i := 0; i < len(bcdStr); i += 2 {
		hi, err = strconv.ParseUint(bcdStr[i:i+1], 16, 8)
		if err != nil {
			panic(err)
		}
		lo, err = strconv.ParseUint(bcdStr[i:i+2], 16, 8)
		if err != nil {
			panic(err)
		}
		byteArr[i] = byte((hi << 4) + lo)
	}
	return
}

func SwapBytesOrderForNew(src []byte) (dest []byte) {
	dest = make([]byte, len(src))
	for i := 0; i < len(src); i++ {
		dest[i] = src[len(src)-1-i]
	}
	return
}

func ReadFile(relativePath string) (data []byte, err error) {
	var dirPath string
	dirPath, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return
	}

	filePath := filepath.Join(dirPath, relativePath)
	_, err = os.Stat(filePath)
	if err != nil {
		return
	}
	data, err = ioutil.ReadFile(filePath)
	return
}
