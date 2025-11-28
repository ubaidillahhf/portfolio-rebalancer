package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		message      string
		expectedCode int
		expectedBody ErrorResponse
	}{
		{
			name:         "bad request error",
			code:         http.StatusBadRequest,
			message:      "Invalid input",
			expectedCode: http.StatusBadRequest,
			expectedBody: ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid input",
				Code:    400,
			},
		},
		{
			name:         "not found error",
			code:         http.StatusNotFound,
			message:      "Resource not found",
			expectedCode: http.StatusNotFound,
			expectedBody: ErrorResponse{
				Error:   "Not Found",
				Message: "Resource not found",
				Code:    404,
			},
		},
		{
			name:         "internal server error",
			code:         http.StatusInternalServerError,
			message:      "Something went wrong",
			expectedCode: http.StatusInternalServerError,
			expectedBody: ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Something went wrong",
				Code:    500,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			RespondWithError(w, tt.code, tt.message)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			var response ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Error != tt.expectedBody.Error {
				t.Errorf("expected error '%s', got '%s'", tt.expectedBody.Error, response.Error)
			}
			if response.Message != tt.expectedBody.Message {
				t.Errorf("expected message '%s', got '%s'", tt.expectedBody.Message, response.Message)
			}
			if response.Code != tt.expectedBody.Code {
				t.Errorf("expected code %d, got %d", tt.expectedBody.Code, response.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
			}
		})
	}
}

func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		payload      interface{}
		expectedCode int
	}{
		{
			name: "success response with map",
			code: http.StatusOK,
			payload: map[string]interface{}{
				"message": "success",
				"data":    "test",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "created response with struct",
			code: http.StatusCreated,
			payload: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   "123",
				Name: "Test",
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			RespondWithJSON(w, tt.code, tt.payload)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Verify it's valid JSON
			var result map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
		})
	}
}
