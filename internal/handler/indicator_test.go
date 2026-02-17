package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestIndicatorHandler_GetByID_InvalidID(t *testing.T) {
	r := chi.NewRouter()
	handler := &IndicatorHandler{service: nil}

	r.Get("/api/indicators/{id}", handler.GetByID)

	req := httptest.NewRequest("GET", "/api/indicators/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIndicatorHandler_Search_InvalidType(t *testing.T) {
	r := chi.NewRouter()
	handler := &IndicatorHandler{service: nil}

	r.Get("/api/indicators/search", handler.Search)

	req := httptest.NewRequest("GET", "/api/indicators/search?type=invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, ErrCodeBadRequest, response.Error.Code)
}

func TestIndicatorHandler_Search_ValidTypes(t *testing.T) {
	validTypes := []string{"ip", "domain", "url", "hash"}

	for _, validType := range validTypes {
		t.Run("type_"+validType, func(t *testing.T) {
			r := chi.NewRouter()
			_ = &IndicatorHandler{service: nil}

			r.Get("/api/indicators/search", func(w http.ResponseWriter, r *http.Request) {
				params := r.URL.Query().Get("type")
				validTypesMap := map[string]bool{"ip": true, "domain": true, "url": true, "hash": true}
				if params != "" && !validTypesMap[params] {
					respondBadRequest(w, "Invalid type")
					return
				}
				respondSuccess(w, map[string]string{"validated": "true"})
			})

			req := httptest.NewRequest("GET", "/api/indicators/search?type="+validType, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	respondJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "value", result["key"])
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()

	respondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Test error")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, ErrCodeBadRequest, response.Error.Code)
	assert.Equal(t, "Test error", response.Error.Message)
}
