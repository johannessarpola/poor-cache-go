package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/johannessarpola/poor-cache-go/internal/common"
	"github.com/stretchr/testify/assert"
)

func errJson(err error) string {
	m := newErr(err)
	b, _ := json.Marshal(m)
	return string(b)
}

func asJsonStr(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func asJson(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func TestSetHandler(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		ttl            string
		body           string
		expectedStatus int
		expectedBody   string
		setFunc        func(key string, value any, ttl time.Duration) error
	}{
		{
			name:           "Successful set",
			key:            "testKey",
			ttl:            "10s",
			body:           `{"value": "testValue"}`,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message": "success"}`,
			setFunc: func(key string, value any, ttl time.Duration) error {
				if key == "testKey" && ttl == 10*time.Second {
					return nil
				}
				return assert.AnError
			},
		},
		{
			name:           "Invalid json",
			key:            "testKey",
			ttl:            "",
			body:           `{1: 2}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errJson(errBadJson),
			setFunc: func(key string, value any, ttl time.Duration) error {
				return nil
			},
		},
		{
			name:           "Missing TTL parameter",
			key:            "testKey",
			ttl:            "",
			body:           `{"value": "testValue"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errJson(errBadQuery),
			setFunc: func(key string, value any, ttl time.Duration) error {
				return nil
			},
		},
		{
			name:           "Store set error",
			key:            "testKey",
			ttl:            "10s",
			body:           `{"value": "testValue"}`,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errJson(errInternal),
			setFunc: func(key string, value any, ttl time.Duration) error {
				return assert.AnError
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStore{
				SetFunc: tt.setFunc,
			}

			service := New(mockStore)
			router := gin.Default()
			router.POST("/set/:key", service.SetHandler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/set/"+tt.key+"?ttl="+tt.ttl, bytes.NewBuffer([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestGetHandler(t *testing.T) {
	start := time.Now()
	data1 := map[string]any{"field": "value"}
	meta1 := common.Meta{
		CreatedAt:  start.Add(-1 * time.Minute),
		ModifiedAt: start,
	}

	tests := []struct {
		name           string
		key            string
		expectedStatus int
		expectedBody   string
		getFunc        func(key string, dest any) (*common.Meta, error)
	}{
		{
			name:           "Successful get",
			key:            "testKey",
			expectedStatus: http.StatusOK,
			expectedBody:   asJsonStr(map[string]any{"meta": meta1, "value": data1}),
			getFunc: func(key string, dest any) (*common.Meta, error) {
				if key == "testKey" {
					json.Unmarshal(asJson(data1), &dest)
					return &meta1, nil
				}
				return nil, assert.AnError
			},
		},
		{
			name:           "Key not found",
			key:            "testKey",
			expectedStatus: http.StatusNotFound,
			expectedBody:   errJson(errNotFound),
			getFunc: func(key string, dest any) (*common.Meta, error) {
				return nil, nil
			},
		},
		{
			name:           "Store get error",
			key:            "testKey",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errJson(errInternal),
			getFunc: func(key string, dest any) (*common.Meta, error) {
				return nil, assert.AnError
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStore{
				GetFunc: tt.getFunc,
			}

			service := New(mockStore)
			router := gin.Default()
			router.GET("/get/:key", service.GetHandler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/get/"+tt.key, nil)
			router.ServeHTTP(w, req)

			body := w.Body.String()
			fmt.Println(body)
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, body)
		})
	}
}

func TestDeleteHandler(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		expectedStatus int
		expectedBody   string
		deleteFunc     func(key string) error
	}{
		{
			name:           "Successful delete",
			key:            "testKey",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message": "success"}`,
			deleteFunc: func(key string) error {
				if key == "testKey" {
					return nil
				}
				return assert.AnError
			},
		},
		{
			name:           "Store delete error",
			key:            "testKey",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errJson(errInternal),
			deleteFunc: func(key string) error {
				return assert.AnError
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStore{
				DeleteFunc: tt.deleteFunc,
			}

			service := New(mockStore)
			router := gin.Default()
			router.DELETE("/delete/:key", service.DeleteHandler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/delete/"+tt.key, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestHasHandler(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		expectedStatus int
		expectedBody   string
		hasFunc        func(key string) bool
	}{
		{
			name:           "Key exists",
			key:            "testKey",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"exists": true}`,
			hasFunc: func(key string) bool {
				return key == "testKey"
			},
		},
		{
			name:           "Key does not exist",
			key:            "testKey",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"exists": false}`,
			hasFunc: func(key string) bool {
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStore{
				HasFunc: tt.hasFunc,
			}

			service := New(mockStore)
			router := gin.Default()
			router.GET("/has/:key", service.HasHandler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/has/"+tt.key, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
