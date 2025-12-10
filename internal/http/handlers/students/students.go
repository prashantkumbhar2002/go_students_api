package students

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/prashantkumbhar2002/go_students_api/internal/http/response"
	"github.com/prashantkumbhar2002/go_students_api/internal/types"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){

		var student types.Student
		// Decode the request body into the student struct
		err := json.NewDecoder(r.Body).Decode(&student)
		if errors.Is(err, io.EOF) {
			slog.Error("Error decoding request body", "error", err)
			response.WriteError(w, http.StatusBadRequest, "invalid request body", "request body is empty")
			return
		}

		if err != nil {
			slog.Error("Error decoding request body", "error", err)
			response.WriteError(w, http.StatusBadRequest, "invalid request body", err.Error())
			return
		}

		// Request Body validation 
		if err := validator.New().Struct(student); err != nil {
			slog.Error("Error validating request body", "error", err)
			response.WriteValidationErrors(w, http.StatusBadRequest, err.(validator.ValidationErrors)) // type assertion to get the ValidationErrors
			return
		}

		

		slog.Info("Student created", "student", student)
		response.WriteJson(w, http.StatusCreated, map[string]string{"message": "Student resource created successfully", "id": student.ID})
	}
}