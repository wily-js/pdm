package dto

import (
	"encoding/json"
	"pdm/repo/entity"
	"time"
)

// DockingDocDto 搜索发布版本
type DockingDocDto struct {
	ID        int               `json:"id"` // 版本ID
	CreatedAt entity.DateTime   `json:"createdAt"`
	UpdatedAt entity.DateTime   `json:"updatedAt"`
	Name      string            `json:"name"`      // 对接文档名
	Asserts   map[string]string `json:"asserts"`   // 附件列表
	Content   string            `json:"content"`   // 版本更新描述
	Publisher member            `json:"publisher"` // 项目负责人ID
}

// Transform 传入参数
func (p *DockingDocDto) Transform(c *entity.DockingDocuments, u *entity.User) *DockingDocDto {
	p.ID = c.ID
	p.CreatedAt = entity.DateTime(c.CreatedAt)
	p.UpdatedAt = entity.DateTime(c.UpdatedAt)
	p.Name = c.Name
	err := json.Unmarshal([]byte(c.Assets), &p.Asserts)
	if err != nil {
		p.Asserts = make(map[string]string)
	}
	p.Content = c.Content
	p.Publisher = member{
		ID:   c.UserId,
		Name: u.Name,
	}
	return p
}

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
