package entity

import (
	"encoding/json"
	"time"
)

const (
	ProjectSetUp       = 0 // 立项
	ProjectDesign      = 1 // 设计
	ProjectDevelopment = 2 // 开发
	ProjectTest        = 3 // 测试
	ProjectPublish     = 4 // 发布
	ProjectIteration   = 5 // 迭代
)

// Project 项目
type Project struct {
	ID           int       `gorm:"autoIncrement" json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Name         string    `json:"name"`         // 项目名称
	NamePinyin   string    `json:"namePinyin"`   // 项目名称拼音缩写
	Description  string    `json:"description"`  // 简介
	Manager      int       `json:"manager"`      // 项目负责人ID
	Version      string    `json:"version"`      // 版本号
	Stage        int       `json:"stage"`        // 阶段类型 枚举值：0 - 立项（默认值）1 - 设计	2 - 开发	3 - 测试	4 - 发布	5 - 迭代
	IssueCount   int       `json:"issueCount"`   //产品问题数量
	ProblemCount int       `json:"problemCount"` //研发问题数量
	IsDelete     int       `json:"isDelete"`     // 是否删除 0 - 未删除（默认值） 1 - 删除
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
