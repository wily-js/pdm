package reuint

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetParams(t *testing.T) {
	//str := "[]"
	//fmt.Println(len(str))
	str := "{\"key\":\"id\",\"value\":\"1\",\"description\":\"ID\"}"
	list := getList(str)
	fmt.Println(list)
	var dataList []map[string]interface{}
	for _, param := range list {
		var temp interface{}
		if err := json.Unmarshal([]byte(param), &temp); err != nil {
			t.Failed()
		}
		tempMap := temp.(map[string]interface{})
		dataList = append(dataList, tempMap)
		//fmt.Println(tempMap)
	}
	fmt.Println(dataList)
}
