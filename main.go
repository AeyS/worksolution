package main

import (
	"github.com/fsnotify/fsnotify"
	"self/worksolution/analysis"
	"log"
	"time"
	"os"
)

// 读取订单数据，清洗数据，然后保存
func ReadXlsAndClearSaveOrder(xlsPath string){
	arr := analysis.ReadXlsx(xlsPath)
	analysis.Savedata2File(arr)
}

//监听文件改变
func watchFolder(folderPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
//				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
//					log.Println("modify file:", event.Name)
					analysis.CopyXlsx("./config.ini")
					log.Println("Dump file Release: d:/test.xlsx")
					ReadXlsAndClearSaveOrder("")
					time.Sleep(6 * time.Second)
				}
			case err := <-watcher.Errors:
				log.Println("err:", err)
			}
		}
	}()

	err = watcher.Add(folderPath);
	if err != nil {
		log.Fatalln(err)
	}
	<-done
}

func WatcherFile(){
	cfg := analysis.GetJsonVal("./config.json")
	xlsPath := cfg.WatchPath
	log.Println(xlsPath)
	initialStat, err := os.Stat(xlsPath)
	if err != nil {
		log.Fatalln("xlsPath stat err", err)
	}
	for {
		stat, err := os.Stat(xlsPath)
		if err != nil {
			log.Fatalln("xlsPath stat err", err)
		}
		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			// 修改为当前值，为下次判断做准备
			initialStat = stat
			// 工作
			analysis.CopyXlsxWatcher(cfg)
			//log.Println("Dump file Release: d:/test.xlsx")
			ReadXlsAndClearSaveOrder(xlsPath)
			analysis.PushOrderData(cfg)
		}
		time.Sleep(3 * time.Second)
	}
}

/**
一旦保存文件，ws会自动拷贝一分副本，并且对副本进行操作，为减少重复操作，限定为3分钟内只做一次监听，
因为可能存在删除之前几个月份的某几行数据，而一旦删除数据，该数据后面的所有订单数据的索引排序都得变更，这将使得数据的存储变得非常昂贵。
因此只保存纯文本格式的数据，避免大面积的刷新数据是个简单的解决方案，服务器端要展示数据，或者分析数据直接通过纯文本获取。
 */

var VERSION string = "0.1"

func main() {
	if len(os.Args)>1 {
		if os.Args[1] == "-v" || os.Args[1] == "-ver" || os.Args[1] == "-version" {
			println("Author: aeys\nEmail: 1481681434@qq.com\nVersion:", VERSION)
			return
		}
	}
	f, err := os.OpenFile("./debug.log", os.O_RDWR| os.O_CREATE| os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)
	WatcherFile()
}
