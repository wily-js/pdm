package entity

import (
	"encoding/json"
	"time"
)

// DockingDocuments 对接文档
type DockingDocuments struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   string    `json:"version"` // 版本号
	UserId    int       `json:"userId"`  // 发布者
	Content   string    `json:"content"` // 版本更新描述
	Assets    string    `json:"assets"`  // 附件列表
}

func (c *DockingDocuments) MarshalJson() ([]byte, error) {
	type Alias DockingDocuments
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
