package storage

import (
	"errors"

	"github.com/prashantkumbhar2002/go_students_api/internal/types"
)

// Domain-specific errors - these are sentinel errors that can be checked using errors.Is()
var (
	ErrNotFound    = errors.New("student not found")
	ErrDuplicate   = errors.New("student already exists")
	ErrInvalidData = errors.New("invalid student data")
	ErrDatabase    = errors.New("database error")
)

type Storage interface {
	CreateStudent(name string, email string, age int) (int64, error)
	GetStudent(id int64) (types.Student, error)
	// GetStudentsList returns paginated list of students
	// offset: number of records to skip, limit: max number of records to return
	GetStudentsList(offset, limit int) ([]types.Student, error)
	// GetStudentsCount returns total count of students in database
	GetStudentsCount() (int64, error)
}
