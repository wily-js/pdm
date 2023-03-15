package entity

import (
	"encoding/json"
	"time"
)

type User struct {
	ID         int       `gorm:"autoIncrement" json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Username   string    `json:"username"` // 用户名【唯一】
	Name       string    `json:"name"`
	NamePinyin string    `json:"namePinyin"`
	Password   Pwd       `json:"password"` //口令加盐摘要Hex
	Salt       string    `json:"-"`        // 盐值Hex
	Avatar     []byte    `json:"avatar"`   // 头像 二进制值
	OpenId     string    `json:"openId"`   // 开放ID 用于关联三方系统，可以是工号
	Phone      string    `json:"phone"`    // 手机号
	Email      string    `json:"email"`    // 邮箱
	IsDelete   int       `json:"isDelete"` // 是否删除 0 - 未删除（默认值） 1 - 删除
}

func (c *User) MarshalJSON() ([]byte, error) {
	type Alias User
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
