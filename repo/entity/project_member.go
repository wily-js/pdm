package entity

import (
	"encoding/json"
	"time"
)

const (
	RoleGuest      = 0 // 访客
	RoleTester     = 1 // 测试
	RoleDeveloper  = 2 // 开发
	RoleMaintainer = 3 // 维护
	RoleCreator    = 4 // 创建者
	RoleManger     = 5 // 管理员
)

// ProjectMember 项目成员
type ProjectMember struct {
	ID          int       `gorm:"autoIncrement" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Role        int       `json:"role"`        // 角色 角色类型包括：0 - 访客，1 - 测试，2 - 开发，3 - 维护，4-负责人
	ProjectId   int       `json:"projectId"`   // 项目ID
	UserId      int       `json:"userId"`      // 用户ID
	Description string    `json:"description"` // 任务描述
}

func (c *ProjectMember) MarshalJson() ([]byte, error) {
	type Alias ProjectMember
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
