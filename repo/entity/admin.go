package entity

import (
	"encoding/json"
	"time"
)

const (
	B = "pdm" // 可区分标识符
)

// Admin 管理员
type Admin struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Username  string    `json:"username"` // 用户名【唯一】
	Password  Pwd       `json:"password"` //口令加盐摘要Hex
	Salt      string    `json:"-"`        // 盐值Hex
	Role      int       `json:"role"`     // 角色类型 0 - 管理员 1 - 审计员
	Cert      string    `json:"cert"`     // 证书
}

func (c *Admin) MarshalJSON() ([]byte, error) {
	type Alias Admin
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
