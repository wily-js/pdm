package dto

import (
	"pdm/repo/entity"
)

type CategorizeDto struct {
	ID        int             `gorm:"autoIncrement" json:"id"`
	CreatedAt entity.DateTime `json:"createdAt"`
	UpdatedAt entity.DateTime `json:"updatedAt"`
	ParentId  int             `json:"ParentId"` // 父分类ID
	Name      string          `json:"name"`     // 分类名称 或用例名称
	Type      string          `json:"type"`     // "categorize" 或 "case"
}

type CategorizeListDto struct {
	ID        int             `gorm:"autoIncrement" json:"id"`
	CreatedAt entity.DateTime `json:"createdAt"`
	UpdatedAt entity.DateTime `json:"updatedAt"`
	Name      string          `json:"name"`   // 分类名称 或用例名称
	Type      string          `json:"type"`   // "categorize" 或 "case"
	Method    int             `json:"method"` // 请求方法
}

// Transform 将实体数据赋值给dto返回给前端
func (c *CategorizeListDto) Transform(u *entity.ApiCategorize, s *entity.ApiCase, typ string) *CategorizeListDto {
	if typ == "categorize" {
		c.ID = u.ID
		c.CreatedAt = entity.DateTime(u.CreatedAt)
		c.UpdatedAt = entity.DateTime(u.UpdatedAt)
		c.Name = u.Name
		c.Method = 255
		c.Type = "categorize"
	} else if typ == "case" {
		c.ID = s.ID
		c.CreatedAt = entity.DateTime(s.CreatedAt)
		c.UpdatedAt = entity.DateTime(s.UpdatedAt)
		c.Name = s.Name
		c.Method = s.Method
		c.Type = "case"
	}
	return c
}
