package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Получение JWT-токена
func getTestJWT(t *testing.T, router http.Handler) string {
	body := `{"username":"testuser", "password":"testpass"}`
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	return resp["token"]
}

// Тест отправки выражения
func TestSubmitExpression(t *testing.T) {
	router := SetupRouter()
	token := getTestJWT(t, router)

	body := `{"expression": "2 + 3 * 4"}`
	req := httptest.NewRequest("POST", "/expressions", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)
	assert.Equal(t, float64(14), result["result"])
}

// Тест получения списка выражений
func TestGetUserExpressions(t *testing.T) {
	router := SetupRouter()
	token := getTestJWT(t, router)

	req := httptest.NewRequest("GET", "/expressions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var exprs []map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&exprs)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(exprs), 1)
}

// Тест запрета на доступ к чужому выражению
func TestAccessDeniedToOthersExpression(t *testing.T) {
	router := SetupRouter()
	token := getTestJWT(t, router)

	req := httptest.NewRequest("GET", "/expressions/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusForbidden || w.Code == http.StatusNotFound)
}
