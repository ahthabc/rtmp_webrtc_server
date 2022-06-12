package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PrintJSON(obj interface{}) {
	txt, _ := json.MarshalIndent(obj, "", "  ")
	fmt.Println(string(txt))
}

func ExpandUser(p string) string {
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[1:])
	}

	return p
}
func PrintJson(jsonStr string) {

	//解析json字符串
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		fmt.Println(err)
		return
	}

	for k, v := range result {
		printFmt("", k, v)
	}
}

func printObj(preKey string, datas interface{}) {

	resultMap, ok := datas.(map[string]interface{})
	if ok {
		for k, v := range resultMap {
			printFmt(preKey, k, v)
		}
	} else {
		resultInterface, ok := datas.([]interface{})
		if ok {
			for _, v := range resultInterface {
				printFmt(preKey, "", v)
			}
		}
	}

}

func printFmt(preKey, currentKey string, value interface{}) {

	//设置key的组合
	var newKey string
	if !(preKey != "" && currentKey != "") {
		newKey = fmt.Sprintf("%v", preKey+currentKey)
	} else {
		newKey = fmt.Sprintf("%v.%v", preKey, currentKey)
	}

	//判断是否是基础类型，是就直接打印，不是进入下次循环
	switch value.(type) {
	case int, int32, int64, float32, float64, string:
		fmt.Println(fmt.Sprintf("%v:%v", newKey, value))
	case interface{}:
		printObj(newKey, value)
	}
}
