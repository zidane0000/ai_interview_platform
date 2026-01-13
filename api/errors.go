package api

import "net/http"

// Centralized error messages and codes for API responses

const (
	ErrMsgMissingInterviewID  = "Bad Request: missing interview ID"
	ErrMsgMissingEvaluationID = "Bad Request: missing evaluation ID"
	ErrMsgMethodNotAllowed    = "Method Not Allowed"
)

const (
	ErrCodeBadRequest       = http.StatusBadRequest
	ErrCodeMethodNotAllowed = http.StatusMethodNotAllowed
)
