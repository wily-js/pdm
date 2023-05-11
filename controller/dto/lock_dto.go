package dto

type LockDto struct {
	UserId    int    `json:"userId"`    // 用户ID
	ProjectId int    `json:"projectId"` // 项目ID
	Id        int    `json:"id"`        // 文档ID
	DocType   string `json:"docType"`   // 文档类型
}
