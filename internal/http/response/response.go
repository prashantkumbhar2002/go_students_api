package response

import (
	"encoding/json"
	// "log/slog"
	"net/http"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	StatusOK = "ok"
	StatusError = "Error"
)
type ErrResponse struct {
	Error   string 	`json:"error"`
	Status  string  `json:"status"`
	Message string  `json:"message,omitempty"`
}

func WriteJson(w http.ResponseWriter, status int, data any) error {
	// slog.Info("Writing JSON response", "status", status, "data", data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, err string, message string) error {
	return WriteJson(w, status, ErrResponse{
		Error:   err,
		Status:  StatusError,
		Message: message,
	})
}

func WriteValidationErrors(w http.ResponseWriter, status int, errors validator.ValidationErrors) error {
	var errMsgs []string
	for _, err := range errors {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("%s is required", err.Field()))
		case "uuid":
			errMsgs = append(errMsgs, fmt.Sprintf("%s is not a valid UUID", err.Field()))
		case "min":
			errMsgs = append(errMsgs, fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param()))
		case "max":
			errMsgs = append(errMsgs, fmt.Sprintf("%s must be less than %s", err.Field(), err.Param()))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("%s is not a valid email", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("%s is not valid for tag %s", err.Field(), err.ActualTag()))
		}
	}
	return WriteJson(w, status, ErrResponse{
		Error:   "validation errors",
		Status:  StatusError,
		Message: strings.Join(errMsgs, "; "),
	})
}