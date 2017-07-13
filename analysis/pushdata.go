package analysis

import (
	"log"
	"net/url"
	"net/http"
	"os"
	"bufio"
	"fmt"
	"crypto/tls"
	"strings"
)
/**
到这边过来处理的都是字符串
 */

const (
	GET  = "GET"
	POST = "POST"
)

type PushVector struct{
	url string
	param url.Values
	method string
}

func HttpPush(_url, method string, param url.Values) (*http.Response, error) {
	var respone = &http.Response{}
	var err error
	switch method {
	case GET:
		respone, err = http.Get(_url)
		if err != nil {
			log.Println("respone err...", err)
		}
	case POST:
		respone, err = http.PostForm(_url, param)
		if err != nil {
			log.Println("respone err...", err)
		}
	}
	return respone, err
}

func HttpSPush(_url, method string, param url.Values) (*http.Response, error){
	var respone = &http.Response{}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	var err error
	switch method {
	case GET:
		respone, err = client.Get(_url)
		if err != nil {
			log.Println("respone err...", err)
		}
	case POST:
		respone, err = client.PostForm(_url, param)
		if err != nil {
			log.Println("respone err...", err)
		}
	}
	return respone, err
}

/**
推送服务
 */
func PushData(_url, method string, param url.Values) (*http.Response, error) {
	if strings.Index(_url, "https")>-1 {
		return HttpSPush(_url, method, param)
	}else{
		return HttpPush(_url, method, param)
	}
}

// 推送订单数据  -  读取本地缓存数据,挑出未推送的数据，进行推送，
// 因为（Savedata2File）第二次保存的时候会把与上一次相同的数据翻牌成已推送，所以，在这边不做任何的数据修改。
// 反正每次保存都会触发一次推送，而（Savedata2File）第二次保存就会自动翻牌，就没有必要在这做二次修改。
func PushOrderData(w Watcher){
	posturl := "https://192.168.1.109:8888/pushorder"
	// 如果缓存文件存在，则检查文件数据与当前数据的状态是否相等
	if Exists(filename_local) {
		lines, err := Readfileline(filename_local)
		if err != nil {
			log.Fatal(err.Error())
		}
		var xls XlsList
		for idx, line := range lines {
			// 当line不等于空行的时候进行对比
			if line != "" {
				xls = LoadxlsList(line)
				if xls.Ispush == false {
					param := url.Values{}
					param.Add("usr", w.Usr)
					param.Add("data", line)
					// 当网络有响应的时候再进行数据传输
					if _, err := PushData(posturl, POST, param); err != nil {
						break
					}
					// 修改状态为已推送
					xls.Ispush = true
					lines[idx] = DumpxlsList(xls)
					//log.Println(xls, line)
				}
			}
		}
		// 将数据保存到本地
		fp, _ := os.Create(filename_local)
		writer := bufio.NewWriter(fp)
		defer fp.Close()
		for _, line := range lines{
			fmt.Fprintln(writer, line)
		}
		writer.Flush()
	}
}

/**
检测是否推送
传递过来的数据，只有已推送和未推送两种
已推送： 跳过
未推送： 与服务器沟通是否存在该条数据
	已存在：跳过
	未存在：推送
 */
func CheckUnPush(line string) bool {
	return LoadxlsList(line).Ispush != true
}