package types


type Student struct {
	ID    int64  `json:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=18,max=100"`
}