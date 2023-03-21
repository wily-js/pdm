package entity

import (
	"encoding/json"
	"time"
)

type Cases struct {
	ID           int       `gorm:"autoIncrement" json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Name         string    `json:"name"`         // 分类名称
	UserId       int       `json:"userId"`       // 创建人ID
	CategorizeId int       `json:"categorizeId"` // 所属分类ID
	Description  string    `json:"description"`  // 接口描述
	Method       int       `json:"method"`       // 请求方法 0-GET，1-POST，2-PUT，3-DELETE
	Path         string    `json:"path"`         // 请求路径
	Params       string    `json:"params"`       // 请求参数
	Headers      string    `json:"headers"`      // 请求头
	BodyType     int       `json:"bodyType"`     // 请求体类型 0-none，1-json，2-form，3-binary
	Body         string    `json:"body"`         // 请求体
}

func (c *Cases) MarshalJSON() ([]byte, error) {
	type Alias Cases
	return json.Marshal(&struct {
		*Alias
		CreatedAt DateTime `json:"createdAt"`
		UpdatedAt DateTime `json:"updatedAt"`
	}{
		(*Alias)(c),
		DateTime(c.CreatedAt),
		DateTime(c.UpdatedAt),
	})
}
