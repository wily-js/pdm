package dto

import "pdm/repo/entity"

// MemberDTO add接口接收前端数据
type MemberDTO struct {
	Role      int   `json:"role"`      // 角色类型
	ProjectId int   `json:"projectId"` // 项目ID
	UserId    []int `json:"userId"`    // 用户ID
}

// MemberAllDTO all接口将数据返回前端
type MemberAllDTO struct {
	ID     int    `json:"id"`     // 记录ID
	UserId int    `json:"userId"` // 用户ID
	Role   int    `json:"role"`   // 角色 角色类型包括：0 - 访客，1 - 测试，2 - 开发，3 - 维护，4-负责人
	Name   string `json:"name"`
}

// Transform 将实体数据赋值给dto返回给前端
func (m *MemberAllDTO) Transform(p *entity.ProjectMember, u *entity.User) *MemberAllDTO {
	m.ID = p.ID
	m.UserId = p.UserId
	m.Role = p.Role
	m.Name = u.Name
	return m
}
