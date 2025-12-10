package types


type Student struct {
	ID    string `json:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Age   int    `json:"age" validate:"required,min=18,max=100"`
}