package entity

import "time"

// TechnicalProposal 技术方案
type TechnicalProposal struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"` // 技术方案文件夹名称
	ProjectId int       `json:"projectId"`
}
