package students

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/prashantkumbhar2002/go_students_api/internal/http/response"
	"github.com/prashantkumbhar2002/go_students_api/internal/types"
	"github.com/prashantkumbhar2002/go_students_api/internal/storage"
)

func NewStudentHandler(storage storage.Storage) http.HandlerFunc {
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

		// Create the student in the database
		id, err := storage.CreateStudent(student.Name, student.Email, student.Age)
		if err != nil {
			slog.Error("Error creating student in the database", "error", err)
			response.WriteError(w, http.StatusInternalServerError, "error creating student", err.Error())
			return
		}

		student.ID = id

		slog.Info("Student created", "student", student)
		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": student.ID})
	}
}

func GetStudentHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// id := r.URL.Query().Get("id") // Reading the query parameters
		id := r.PathValue("id") // Reading the path parameters
		slog.Info("ID", "id", id)
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			slog.Error("Error parsing ID", "error", err)
			response.WriteError(w, http.StatusBadRequest, "invalid ID", err.Error())
			return
		}

		// Get the student from the database
		student, err := storage.GetStudent(idInt)
		if err != nil {
			if err.Error() == "student not found" {
				slog.Error("Student not found", "error", err)
				response.WriteError(w, http.StatusNotFound, "student not found", err.Error())
				return
			}
			slog.Error("Error getting student", "error", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error", err.Error())
			return
		}
		slog.Info("Student fetched by ID", "id", idInt, "student", student)
		response.WriteJson(w, http.StatusOK, student)
	}
}