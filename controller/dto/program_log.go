package dto

// FileItemDto 搜索文件返回值
type FileItemDto struct {
	Name      string `json:"name"`      // 文件名
	Type      string `json:"type"`      // 文件类型：dir、file
	Path      string `json:"path"`      // 文件路径
	Size      int64  `json:"size"`      // 文件大小，单位B
	UpdatedAt string `json:"updatedAt"` // 最后更新时间格式 YYYY-MM-DD HH:mm:ss
}

// FileTransferDto 文件转移DTO
type FileTransferDto struct {
	From string `json:"from"`
	To   string `json:"to"`
}
