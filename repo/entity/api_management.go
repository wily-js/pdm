package entity

import (
	"encoding/json"
	"time"
)

// ApiManagement 接口管理
type ApiManagement struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ProjectId int       `json:"projectId"` // 所属项目ID
	Title     string    `json:"title"`     // 文档名
	Filename  string    `json:"filename"`  // 文件名称
}

func (c *ApiManagement) MarshalJSON() ([]byte, error) {
	type Alias ApiManagement
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
