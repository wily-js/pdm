package dto

import (
	"pdm/repo/entity"
)

// OplogSearchDto 操作日志搜索
type OplogSearchDto struct {
	Start  int64  `form:"start" json:"start"`   // 开始时间
	End    int64  `form:"end" json:"end"`       // 截止时间
	OpType int    `form:"opType" json:"opType"` // 角色类型 0 - 匿名；1 - 管理员；2 - 用户 ; 3 - 应用 ；255 - 所有
	OpId   int    `form:"opId" json:"opId"`     // 用户ID
	OpName string `form:"opName" json:"opName"` // 操作名称，支持模糊
	Page   int    `form:"page" json:"page"`     // 页码 1 起
	Limit  int    `form:"limit" json:"limit"`   // 页容量，默认20
}

// OplogDto 操作日志
type OplogDto struct {
	ID        int             `gorm:"autoIncrement" json:"id"`
	CreatedAt entity.DateTime `json:"createdAt"`
	OpType    int             `json:"opType"`  // 操作者类型 类型如下包括：0 - 匿名 1 - 管理员  2 - 用户  3 - 应用 若不知道用户或没有用户信息，则使用匿名。
	UserID    int             `json:"userId"`  // 用户id
	Name      string          `json:"name"`    // 名称
	OpName    string          `json:"opName"`  // 操作名称
	OpParam   string          `json:"opParam"` // 操作的关键参数 可选参数，例如删除用户时，删除的用户ID，复杂参数请使用JSON对象字符串，如{id: 1}
}

// Transform 将实体数据赋值给dto返回给前端
func (o *OplogDto) Transform(log *entity.Log) *OplogDto {
	o.ID = log.ID
	o.CreatedAt = entity.DateTime(log.CreatedAt)
	o.OpType = log.OpType
	o.UserID = log.OpId
	o.OpName = log.OpName
	o.OpParam = log.OpParam
	return o
}

// OplogExportDto 导出日志
type OplogExportDto struct {
	Start  int64  `form:"start" json:"start"`   // 开始时间
	End    int64  `form:"end" json:"end"`       // 截止时间
	OpType int    `form:"opType" json:"opType"` // 角色类型 0 - 匿名；1 - 管理员；2 - 用户；255 - 所有
	OpId   int    `form:"opId" json:"opId"`     // 用户ID
	OpName string `form:"opName" json:"opName"` // 操作名称，支持模糊
}
