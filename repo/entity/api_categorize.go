package entity

import (
	"encoding/json"
	"time"
)

type ApiCategorize struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ParentId  int       `json:"parentId"`  // 父分类ID
	Name      string    `json:"name"`      // 分类名称
	ProjectId int       `json:"projectId"` // 所属项目ID
	UserId    int       `json:"userId"`    // 创建人ID
}

func (c *ApiCategorize) MarshalJSON() ([]byte, error) {
	type Alias ApiCategorize
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
