package entity

import (
	"encoding/json"
	"time"
)

// Project 项目
type Project struct {
	ID          int       `gorm:"autoIncrement" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `json:"name"`        // 项目名称
	NamePinyin  string    `json:"namePinyin"`  // 项目名称拼音缩写
	Description string    `json:"description"` // 简介
	Manager     int       `json:"manager"`     // 项目负责人ID
	Version     string    `json:"version"`     // 版本号
	IsDelete    int       `json:"isDelete"`    // 是否删除 0 - 未删除（默认值） 1 - 删除
}

func (c *Project) MarshalJson() ([]byte, error) {
	type Alias Project
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
