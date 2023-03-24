package dto

type RespDto struct {
	Header string `json:"header"` // 响应头
	Body   string `json:"body"`   // 响应体
	Cookie string `json:"cookie"` // Cookies
}

// Transform 将数据赋值给dto，返回前端
func (r *RespDto) Transform(header, body, cookie string) *RespDto {
	r.Body = body
	r.Header = header
	r.Cookie = cookie
	return r
}
