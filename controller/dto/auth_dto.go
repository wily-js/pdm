package dto

import (
	"pdm/repo/entity"
	"pdm/reuint/jwt"
)

// LoginToDto login接口后端传数据给前端
type LoginToDto struct {
	UserType string `json:"type"`   // 用户类型
	ID       int    `json:"id"`     // 用户Id
	Openid   string `json:"openid"` // 工号
	Name     string `json:"name"`   // 用户姓名
	Exp      int64  `json:"exp"`    // 会话过期时间，单位Unix时间戳毫秒（ms）
}

// Transform 将数据赋值给dto，返回前端
func (loginToDto *LoginToDto) Transform(claims *jwt.Claims) *LoginToDto {
	loginToDto.UserType = claims.Type
	loginToDto.ID = claims.Sub
	loginToDto.Exp = claims.Exp
	return loginToDto
}

type AdminLoginDto struct {
	UserType string `json:"type"`     // 用户类型
	ID       int    `json:"id"`       // 用户Id
	Username string `json:"username"` // 用户名
	Avatar   []byte `json:"avatar"`   // 头像
	Exp      int64  `json:"exp"`      // 会话过期时间，单位Unix时间戳毫秒（ms）
}

// Transform 将数据赋值给dto，返回前端
func (a *AdminLoginDto) Transform(claims *jwt.Claims, admin *entity.Admin) *AdminLoginDto {
	a.UserType = claims.Type
	a.ID = claims.Sub
	a.Username = admin.Username
	a.Exp = claims.Exp
	return a
}
