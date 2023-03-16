package dto

import "pdm/repo/entity"

// UserCreateDto 创建用户DTO
type UserCreateDto struct {
	ID        int             `json:"id"`
	Openid    string          `json:"openid"`    // 工号
	Name      string          `json:"name"`      // 姓名
	CreatedAt entity.DateTime `json:"createdAt"` // 创建时间
}

// Transform 将实体数据赋值给dto返回给前端
func (u *UserCreateDto) Transform(usr *entity.User) *UserCreateDto {
	u.ID = usr.ID
	u.Openid = usr.QQOpenid
	u.Name = usr.Name
	u.CreatedAt = entity.DateTime(usr.CreatedAt)
	return u
}

// PasswordDto 修改口令接口接收前端的数据
type PasswordDto struct {
	ID     int        `json:"id"`     // 用户ID
	OldPwd entity.Pwd `json:"oldPwd"` // 旧口令
	NewPwd entity.Pwd `json:"newPwd"` // 新口令
}

// NameListDto 接口将以下数据返回给前端
type NameListDto struct {
	ID     int    `json:"id"`     // 用户ID
	Openid string `json:"openid"` // 用户工号
	Name   string `json:"name"`   // 用户姓名
}

// Transform 将实体数据赋值给dto，返回前端
func (nameListDto *NameListDto) Transform(usr *entity.User) *NameListDto {
	nameListDto.ID = usr.ID
	nameListDto.Name = usr.Name
	return nameListDto
}

// UserInfoDto 接口将以下数据返回给前端
type UserInfoDto struct {
	ID       int    `json:"id"`
	Openid   string `json:"openid"`   // 工号
	Username string `json:"username"` // 用户名
	Name     string `json:"name"`     // 姓名
	Phone    string `json:"phone"`    // 手机号
	Email    string `json:"email"`    // 邮箱
	Sn       string `json:"sn"`       // 身份证号
}
