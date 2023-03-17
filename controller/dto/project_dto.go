package dto

import "pdm/repo/entity"

// ProjectDto 项目Dto
type ProjectDto struct {
	ID          int    `gorm:"autoIncrement" json:"id"`
	Name        string `json:"name"`        // 项目名称
	Manager     int    `json:"manager"`     // 项目负责人ID
	Description string `json:"description"` // 简介
	Version     string `json:"version"`     //版本
}

// Transform 传入参数
func (p *ProjectDto) Transform(c *entity.Project) *ProjectDto {
	p.ID = c.ID
	p.Name = c.Name
	p.Manager = c.Manager
	p.Description = c.Description
	p.Version = c.Version
	return p
}

// ProjectSearchDto 项目搜索Dto
type ProjectSearchDto struct {
	ID        int             `gorm:"autoIncrement" json:"id"`
	CreatedAt entity.DateTime `json:"createdAt"`
	UpdatedAt entity.DateTime `json:"updatedAt"`
	Name      string          `json:"name"`    // 项目名称
	Manage    member          `json:"manager"` // 项目负责人ID
	Version   string          `json:"version"` //版本号
}

type member struct {
	ID   int    `json:"id"`   //用户ID
	Name string `json:"name"` //姓名
}

// Transform 传入参数
func (p *ProjectSearchDto) Transform(c *entity.Project, u *entity.User) *ProjectSearchDto {
	p.ID = c.ID
	p.CreatedAt = entity.DateTime(c.CreatedAt)
	p.UpdatedAt = entity.DateTime(c.UpdatedAt)
	p.Name = c.Name
	p.Manage = member{
		ID:   c.Manager,
		Name: u.Name,
	}
	p.Version = c.Version
	return p
}

// ProjectInfoDto 项目详细信息Dto
type ProjectInfoDto struct {
	ID          int             `gorm:"autoIncrement" json:"id"`
	CreatedAt   entity.DateTime `json:"createdAt"`
	Name        string          `json:"name"`        // 项目名称
	Description string          `json:"description"` // 简介
	Manage      member          `json:"manager"`     // 项目负责人ID
	Version     string          `json:"version"`     //版本号
}

// Transform 传入参数
func (p *ProjectInfoDto) Transform(c *entity.Project, u *entity.User) *ProjectInfoDto {
	p.ID = c.ID
	p.CreatedAt = entity.DateTime(c.CreatedAt)
	p.Name = c.Name
	p.Manage = member{
		ID:   c.Manager,
		Name: u.Name,
	}
	p.Description = c.Description
	p.Version = c.Version
	return p
}
