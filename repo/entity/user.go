package entity

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           int       `gorm:"autoIncrement" json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Openid       string    `json:"openid"` // 开放ID 用于关联三方系统，可以是工号
	Name         string    `json:"name"`
	NamePinyin   string    `json:"namePinyin"`
	Password     Pwd       `json:"password"`      //口令加盐摘要Hex
	Salt         string    `json:"-"`             // 盐值Hex
	Username     string    `json:"username"`      // 用户名【唯一】
	Phone        string    `json:"phone"`         // 手机号
	Email        string    `json:"email"`         // 邮箱
	Sn           string    `json:"sn"`            // 身份证号
	QQOpenid     string    `json:"qq_openid"`     // QQ Openid
	WechatOpenid string    `json:"wechat_openid"` // 微信 Openid
	Avatar       string    `json:"avatar"`        // 头像 文件名
	IsDelete     int       `json:"isDelete"`      // 是否删除 0 - 未删除（默认值） 1 - 删除
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
