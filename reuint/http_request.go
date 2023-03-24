package reuint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"pdm/repo/entity"
	"strings"
)

// ParsingData 将json格式的字符串解析为Map格式数据
func ParsingData(data string) ([]map[string]interface{}, error) {
	list := getList(data)
	var dataList []map[string]interface{}
	for _, param := range list {
		var temp interface{}
		if err := json.Unmarshal([]byte(param), &temp); err != nil {
			return nil, err
		}
		tempMap := temp.(map[string]interface{})
		dataList = append(dataList, tempMap)
	}
	return dataList, nil
}

// GenRequestUrl 生成请求地址
func GenRequestUrl(base string, requestParams string) (string, error) {
	params, err := ParsingData(requestParams)
	if err != nil || params == nil {
		return base, err
	}
	var url string
	for i, param := range params {
		if i == 0 {
			url = fmt.Sprintf("%s?%s=%s", base, param["key"], param["value"])
		} else {
			url = fmt.Sprintf("%s&%s=%s", url, param["key"], param["value"])
		}
	}
	return url, nil
}

// GenRequestHeader 生成请求头
func GenRequestHeader(request *http.Request, requestHeaders string) (*http.Request, error) {
	headers, err := ParsingData(requestHeaders)
	if err != nil {
		return request, err
	}
	if headers != nil {
		for _, header := range headers {
			request.Header.Add(header["key"].(string), header["value"].(string))
		}
	}
	return request, nil
}

// GenRequestBody return json格式数据，form-data格式数据,err
// GenRequestBody 生成请求体
func GenRequestBody(bodyType int, requestBody string) (*strings.Reader, url.Values, error) {
	data, err := ParsingData(requestBody)
	if err != nil {
		return nil, nil, err
	}
	if data == nil {
		return nil, nil, nil
	}
	var marshal []byte
	form := make(url.Values)
	switch bodyType {
	case entity.BodyTypeNone:
		return nil, nil, nil
	case entity.BodyTypeJson:
		marshal, err = json.Marshal(data[0])
		if err != nil {
			return nil, nil, err
		}
	case entity.BodyTypeForm:
		for _, body := range data {
			value := make([]string, 0)
			value = append(value, body["value"].(string))
			form[body["key"].(string)] = value
		}
	case entity.BodyTypeBinary:
		//TODO 二进制文件
	default:
		return nil, nil, errors.New("无效参数")
	}
	return strings.NewReader(string(marshal)), form, nil
}

// ParsingResponseBody 生成响应体
func ParsingResponseBody(method int, resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 解析响应体数据
	var respBody []byte
	if method == entity.MethodGet || method == entity.MethodDelete {
		var v interface{}
		err := json.Unmarshal(body, &v)
		if err != nil {
			return "", err
		}
		respBody, err = json.Marshal(v)
		if err != nil {
			return "", err
		}
	} else {
		respBody = body
	}
	return string(respBody), nil
}

// ParsingResponseHeader 生成响应头
func ParsingResponseHeader(resp *http.Response) (string, error) {
	var respHeaders []map[string]string
	for key, value := range resp.Header {
		mp := make(map[string]string)
		mp["key"] = key
		var temp string
		for i, val := range value {
			if i == 0 {
				temp = val
			} else {
				temp = fmt.Sprintf("%s%s", temp, val)
			}
		}
		mp["value"] = temp
		respHeaders = append(respHeaders, mp)
	}
	respHeader, err := json.Marshal(respHeaders)
	if err != nil {
		return "", err
	}
	return string(respHeader), nil
}

// 例：	data := "[{\"key\":\"id\",\"value\":\"1\",\"description\":\"ID\"}]"
// return [{"key":"id","value":"1","description":"ID"}]
// getList 将含有一组或多组json格式数据的字符串转化为切片
func getList(data string) []string {

	if !strings.HasSuffix(data, "[") && !strings.HasSuffix(data, "]") {
		if !strings.HasSuffix(data, "{") && !strings.HasSuffix(data, "}") {
			return nil
		}
	}
	if strings.HasPrefix(data, "[") && strings.HasSuffix(data, "]") {
		data = data[1 : len(data)-1]
	}
	off := 0
	list := make([]string, 0)
	for i := 0; i < len(data); i++ {
		if data[i] == '{' {
			off = i
		}
		if data[i] == '}' {
			param := data[off : i+1]
			list = append(list, param)
		}
	}
	return list
}
