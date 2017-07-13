package analysis

import (
	"os"
	"io"
	"strconv"
	"strings"
	"log"
	"io/ioutil"
	"fmt"
	"encoding/json"
)

// 拷贝文件
func CopyFile(src string, dst string) {
	src_s, _ := os.Open(src)
	defer src_s.Close()
	dst_s, _ := os.Create(dst)
	defer dst_s.Close()
	io.Copy(dst_s, src_s)
}

// 读取文件内容
func Readfile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// 读取文件内容，并分割成行
func Readfileline(filename string) ([]string, error) {
	con, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(con), "\n"), err
}

type XlsList struct {
	Id     int
	Con    string
	Ispush bool
}

//dump对象
func DumpxlsList(line XlsList) string {
	return strconv.Itoa(line.Id) + "+|+" + line.Con + "+|+" + strconv.FormatBool(line.Ispush)
}

//load对象
func LoadxlsList(line string) (xls XlsList) {
	line_sp := strings.Split(line, "+|+")
	var err error
	xls.Id, err = strconv.Atoi(line_sp[0])
	if err != nil {
		log.Println("strconv.Atoi error", err)
	}
	xls.Con = line_sp[1]
	xls.Ispush, err = strconv.ParseBool(line_sp[2])
	if err != nil {
		log.Println("strconv.ParseBool error", err)
	}
	return xls
}

//判断文件是否存在
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// 获取配置键值
func GetIniVal(path, sec, key string) (string, bool) {
	if cfg, err := New(path); err != nil {
		log.Println("new config err", err)
		return "", false
	} else {
		if val, ok := cfg.Get(sec, key); ok != true {
			msg := fmt.Sprintf("get %s %s.%s error", path, sec, key)
			log.Printf(msg)
			return val, false
		} else {
			return val, true
		}
	}
}

type Watcher struct {
	WatchPath string
	Src string
	Dst string
	Usr string
}

// 获取json配置
func GetJsonVal(path string) (w Watcher){
	con, err := Readfile(path)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(con, &w)
	return w
}