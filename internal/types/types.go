package types

type Student struct {
	ID    int64  `json:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=18,max=100"`
}

// PaginationParams holds pagination query parameters
type PaginationParams struct {
	Page  int `json:"page"`  // Current page number (1-indexed)
	Limit int `json:"limit"` // Number of items per page
}

// PaginatedResponse wraps paginated results with metadata
type PaginatedResponse struct {
	Data       interface{} `json:"data"`        // The actual data (students)
	Page       int         `json:"page"`        // Current page
	Limit      int         `json:"limit"`       // Items per page
	TotalItems int64       `json:"total_items"` // Total number of items
	TotalPages int         `json:"total_pages"` // Total number of pages
	HasNext    bool        `json:"has_next"`    // Whether there's a next page
	HasPrev    bool        `json:"has_prev"`    // Whether there's a previous page
}

// Default pagination values
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100 // Prevent clients from requesting too many records
	MinLimit     = 1
)
