package response

// Response 通用API响应结构
type Response struct {
	Status     string      `json:"status"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	TotalCount  int `json:"total_count"`
	PageSize    int `json:"page_size"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}, messgae string, pagination ...Pagination) Response {
	resp := Response{
		Status:  "success",
		Message: messgae,
	}

	if data != nil {
		resp.Data = data
	}

	if len(pagination) > 0 {
		resp.Pagination = &pagination[0]
	}

	return resp
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string) Response {
	return Response{
		Status:  "error",
		Message: message,
	}
}
