package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_Interview_MethodNotAllowed(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("PUT", "/api/interviews", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", w.Code)
	}
}

func TestRouter_InterviewID_BadRequest(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/interviews/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}

func TestRouter_EvaluationID_BadRequest(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/evaluation/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", w.Code)
	}
}

func TestRouter_Evaluation_MethodNotAllowed(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("PUT", "/api/evaluation", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", w.Code)
	}
}
