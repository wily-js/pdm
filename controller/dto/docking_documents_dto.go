package dto

import (
	"encoding/json"
	"pdm/repo/entity"
	"time"
)

// UriDto URI结构体
type UriDto struct {
	DocType string `json:"docType"` // 资源类型:image,file
	Uri     string `json:"uri"`     // 资源访问地址
}

type DocDto struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
	ProjectId int       `json:"projectId"` // 所属项目ID
	Title     string    `json:"title"`     // 文档名
	DocType   string    `json:"docType"`   // 文档类型 可选值有：markdown、word、txt、excel
	Priority  int       `json:"priority"`  // 优先级 默认为0，越大优先级越高，用于文档排序，非特殊情况保持0即可。
	Filename  string    `json:"filename"`  // 文件名称
}

func (c *DocDto) MarshalJSON() ([]byte, error) {
	type Alias DocDto
	return json.Marshal(&struct {
		*Alias
		UpdatedAt entity.DateTime `json:"updatedAt"`
	}{
		(*Alias)(c),
		entity.DateTime(c.UpdatedAt),
	})
}
