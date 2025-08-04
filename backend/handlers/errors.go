package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// Standard error response structures
type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

type ValidationErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details"`
}

// HTTP status code constants
const (
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusMethodNotAllowed    = http.StatusMethodNotAllowed
	StatusConflict            = http.StatusConflict
	StatusInternalServerError = http.StatusInternalServerError
	StatusServiceUnavailable  = http.StatusServiceUnavailable
)

// Standard error messages
const (
	ErrorMethodNotAllowed     = "HTTP method not allowed for this endpoint"
	ErrorInvalidJSON          = "Invalid JSON format in request body"
	ErrorInternalServer       = "Internal server error. Please try again later."
	ErrorServiceUnavailable   = "Service temporarily unavailable. Please try again later."
	ErrorResourceNotFound     = "Requested resource not found"
	ErrorValidationFailed     = "Request validation failed"
	ErrorInvalidID            = "Invalid ID format"
	ErrorDatabaseConnection   = "Database connection error. Please try again."
)

// sendJSONError sends a standardized JSON error response
func sendJSONError(w http.ResponseWriter, statusCode int, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := ErrorResponse{
		Error:   message,
		Details: details,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// sendValidationError sends a validation error response with multiple error details
func sendValidationError(w http.ResponseWriter, errors []ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusBadRequest)
	
	errorMessages := make([]string, len(errors))
	for i, err := range errors {
		errorMessages[i] = err.Error()
	}
	
	response := ValidationErrorResponse{
		Error:   ErrorValidationFailed,
		Details: errorMessages,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode validation error response: %v", err)
	}
}

// Standard error response helpers

// BadRequest sends a 400 Bad Request error
func BadRequest(w http.ResponseWriter, message string, details interface{}) {
	sendJSONError(w, StatusBadRequest, message, details)
}

// NotFound sends a 404 Not Found error
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = ErrorResourceNotFound
	}
	sendJSONError(w, StatusNotFound, message, nil)
}

// MethodNotAllowed sends a 405 Method Not Allowed error
func MethodNotAllowed(w http.ResponseWriter) {
	sendJSONError(w, StatusMethodNotAllowed, ErrorMethodNotAllowed, nil)
}

// InternalServerError sends a 500 Internal Server Error
func InternalServerError(w http.ResponseWriter, message string) {
	if message == "" {
		message = ErrorInternalServer
	}
	sendJSONError(w, StatusInternalServerError, message, nil)
}

// ServiceUnavailable sends a 503 Service Unavailable error
func ServiceUnavailable(w http.ResponseWriter, message string) {
	if message == "" {
		message = ErrorServiceUnavailable
	}
	sendJSONError(w, StatusServiceUnavailable, message, nil)
}

// InvalidJSON sends a 400 Bad Request error for JSON parsing failures
func InvalidJSON(w http.ResponseWriter, err error) {
	BadRequest(w, ErrorInvalidJSON, err.Error())
}

// InvalidID sends a 400 Bad Request error for invalid ID format
func InvalidID(w http.ResponseWriter, idType string) {
	message := ErrorInvalidID
	if idType != "" {
		message = "Invalid " + idType + " format"
	}
	BadRequest(w, message, nil)
}

// ValidationFailed sends a 400 Bad Request error for validation failures
func ValidationFailed(w http.ResponseWriter, errors []ValidationError) {
	sendValidationError(w, errors)
}

// DatabaseError sends a 500 Internal Server Error for database issues
func DatabaseError(w http.ResponseWriter, operation string, err error) {
	log.Printf("Database error during %s: %v", operation, err)
	InternalServerError(w, ErrorDatabaseConnection)
}

// logAndError logs an error and sends an HTTP error response
func logAndError(w http.ResponseWriter, statusCode int, logMessage string, responseMessage string) {
	log.Printf(logMessage)
	sendJSONError(w, statusCode, responseMessage, nil)
}