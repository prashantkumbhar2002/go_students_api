package helpers

import (
	"net/http"
	"strconv"

	"github.com/prashantkumbhar2002/go_students_api/internal/types"
)

// parsePaginationParams extracts and validates pagination parameters from request
func ParsePaginationParams(r *http.Request) types.PaginationParams {
	// Get query parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Set defaults
	page := types.DefaultPage
	limit := types.DefaultLimit

	// Parse page
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			// Enforce maximum limit to prevent abuse
			if limit > types.MaxLimit {
				limit = types.MaxLimit
			}
			if limit < types.MinLimit {
				limit = types.MinLimit
			}
		}
	}

	return types.PaginationParams{
		Page:  page,
		Limit: limit,
	}
}