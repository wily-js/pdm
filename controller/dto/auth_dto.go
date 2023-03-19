package dto

import "pdm/reuint/jwt"

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
