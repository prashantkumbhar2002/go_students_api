package storage

import "github.com/prashantkumbhar2002/go_students_api/internal/types"

type Storage interface {
	CreateStudent(name string, email string, age int) (int64, error)
	GetStudent(id int64) (types.Student, error)
}
