package entity

import "time"

// Document 文档
type Document struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ProjectId int       `json:"projectId"` // 所属项目ID
	Title     string    `json:"title"`     // 文档名
	DocType   string    `json:"docType"`   // 文档类型 可选值有：markdown、word、txt、excel
	Priority  int       `json:"priority"`  // 优先级 默认为0，越大优先级越高，用于文档排序，非特殊情况保持0即可。
	Filename  string    `json:"filename"`  // 文件名称
}
