package analysis

import (
	"log"
	"github.com/tealeg/xlsx"
	"os"
	"bufio"
	"fmt"
	"strings"
)


const Comma string = "<c,>"					//逗号
const Semicolon string = "<s;>"			//分号
const filename_local string = "./local.log"		//缓存数据文件名

// 在未添加分行号（Semicolon）之前，清理空白格
func ClearNULL(line string) string {
	length := len(line)
	// 当NULL>=12个时候，该行为空白航，抛弃掉
	if length < 15 || strings.Count(line, "NULL")>=12 {
//		log.Printf("length less: %d, discard...%s \n", length, line)
		// 判断型号那格为空或者为目录名“型号”的话，才抛弃掉
		line_sp := strings.Split(line, Comma);
		if(len(line_sp) > 9){
			if line_sp[9] == "NULL" || line_sp[9] == "型号"{
				return ""
			}
		}else{
			return ""
		}
	}
	if line[length-8:] == "NULL<c,>"{
		return ClearNULL(line[:length-8])
	}
	return line
}


/**
 读取xlsx文件，每次读取数据都是标记为未推送，只有之后再与本地缓存文件进行对比分析的时候会出现第一次修改推送标记的真假值。
 如果在缓存文件里数据id已存在，则检查数据是否相等，如果相等则跳过，标记为已推送，不相等则替换掉，并标记为未推送状态。
  */
func ReadXlsx(excelFileName string) (arr []XlsList/* 存储数据蒲*/) {
	idx_now_num := -1 // 存储当前id索引值
	line_now_string := "" // 存储当前索引对象文本
	line_push_string := "" // 做一个预备值，在最后一个的时候跳过不操作
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		log.Fatalln("ReadXlsx xlsx.OpenFile:", err)
	}
	var text string // 缓存cell的文本内容
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			line_now_string = ""
			if row == nil {
				break
			}
			for x, cell := range row.Cells {
//				if cell.Type() == xlsx.CellTypeDate {
//					t, _ := cell.GetTime(false)
//					text = fmt.Sprintf("%d-%s-%d", t.Year(), t.Month().String(), t.Day())
//				}else{
//					text, _ = cell.String()
//				}
				if x == 3 {
					t, _ := cell.GetTime(false)
					text = t.UTC().String()
					if text != "0001-01-01 00:00:00 +0000 UTC" {
						text = text[:19]
					}else{
						// 因为有可能存在汉字表达的日期，所以当无法翻译时间的时候，可以直接使用原文日期
						text, _ = cell.String()
					}
				}else{
					text, _ = cell.String()
				}
				if text=="" {
					text = "NULL"
				}
				line_now_string += strings.Replace(text, "\n", "", -1) + Comma
				line_push_string = line_now_string
			}
			if line_push_string = ClearNULL(line_push_string); line_push_string != ""{
				// 当内容不为空的时候才添加到数组
				line_push_string += Semicolon
				//			log.Println(line_push_string)
				idx_now_num += 1
				arr = append(arr, XlsList{idx_now_num, line_push_string, false})
			}
		}
	}
	return arr
}

// 遍历列表，检查id是否存在， 并且返回实例化的元素
func ElementInList(xls_s []XlsList, line string) (XlsList, bool) {
	line_xls := LoadxlsList(line)
	for _, xls := range xls_s{
		if xls.Id == line_xls.Id {
			return line_xls, true
		}
	}
	return line_xls, false
}

// 判断内容是否一致
func ElementConEqual(xls_a XlsList, xls_b XlsList) bool {
	return xls_a.Con == xls_b.Con
}

/**
 将数据保存到文本（图片以二进制存储）
 如果在缓存文件里数据id已存在，则检查数据是否相等，如果相等则跳过，标记为已推送，不相等则替换掉，并标记为未推送状态。
 现在有一个比较大的问题是如果其中某条数据被删除了，那么后面的数据索引id就全部都会改变，必须要找到一个唯一值来确认坐标才行
 */
func Savedata2File(arr []XlsList){
	// 如果缓存文件存在，则检查文件数据与当前数据的状态是否相等
	if Exists(filename_local) {
		lines, err := Readfileline(filename_local)
		if err != nil {
			log.Fatal(err.Error())
		}
		for _, line := range lines{
			// 当line不等于空行的时候进行对比
			if line != "" {
				// 如果元素存在缓存数据内，则进行数据判断是否相等?
				if line_xls, isIn := ElementInList(arr, line); isIn {
					// 这里面处理的是满足之前有存在过的数据，判断是否有修改更新需要推送！
					if ElementConEqual(arr[line_xls.Id], line_xls) != true {
						//如果不相等，则将Ispush状态修改为假(未推送)，到时候将已修改内容推送给服务器
						arr[line_xls.Id].Ispush = false
					}else{
						//如果相等，则将之前ReadXlsx时候赋值false修改为true，视为已推送数据。
						arr[line_xls.Id].Ispush = true
					}
				}else{
					// 如果元素不存在缓存数据内，则不做任何处理，直接跳过，因为这个数据是一个新数据
				}
			}
		}
	}
	fp, _ := os.Create(filename_local)
	writer := bufio.NewWriter(fp)
	defer fp.Close()
	for _, line := range arr{
		fmt.Fprintln(writer, DumpxlsList(line))
	}
	writer.Flush()
}


// 清理文件
/**
我打算对文件内容按每个表格的行数做一个索引，每次修改，都会触发程序遍历文件所有行进行比对，如果指定行被修改过，则，程序将旧行替换成新行
把文档理解成一行一行的数据块
 */
