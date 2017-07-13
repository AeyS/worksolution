package analysis

import (
	"log"
	"github.com/vaughan0/go-ini"
)


/**
新建一个配置文件，可以指定文件路径
 */
func New(src string)(ini.File, error){
	return ini.LoadFile(src)
}

// dump文档以供后续操作
func CopyXlsx(src string){
	config, err := New(src)
	if err != nil {
		log.Println(err)
	}
	dst, ok := config.Get("watcher", "dst")
	if ok != true {
		log.Println(ok)
	}
	src, ok_ := config.Get("watcher", "src")
	if ok_ != true {
		log.Println(ok)
	}
	CopyFile(src, dst)
}

// dump文档以供后续操作
func CopyXlsxWatcher(w Watcher){
	CopyFile(w.Src, w.Dst)
}