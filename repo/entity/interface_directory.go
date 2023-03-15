package entity

import "time"

// InterfaceDirectory 接口目录
type InterfaceDirectory struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ProjectId int       `json:"projectId"` // 所属项目ID
	Title     string    `json:"title"`     // 文档名
	Filename  string    `json:"filename"`  // 文件名称
}
