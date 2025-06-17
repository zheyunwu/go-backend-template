package response

// Response is a generic API response structure.
type Response struct {
	Status     string      `json:"status"` // "success" or "error"
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

// NewSuccessResponse creates a success response.
func NewSuccessResponse(data interface{}, message string, pagination ...Pagination) Response {
	resp := Response{
		Status:  "success",
		Message: message,
	}

	if data != nil {
		resp.Data = data
	}

	if len(pagination) > 0 {
		resp.Pagination = &pagination[0]
	}

	return resp
}

// NewErrorResponse creates an error response.
func NewErrorResponse(message string) Response {
	return Response{
		Status:  "error",
		Message: message,
	}
}
