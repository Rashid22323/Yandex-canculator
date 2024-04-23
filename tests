package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegisterUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(registerHandler))
	defer ts.Close()

	user := User{
		Login:    "testuser",
		Password: "testpassword",
	}

	data, err := json.Marshal(user)
	assert.NoError(t, err)

	resp, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(data))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestLoginUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(loginHandler))
	defer ts.Close()

	user := User{
		Login:    "testuser",
		Password: "testpassword",
	}

	data, err := json.Marshal(user)
	assert.NoError(t, err)

	resp, err := http.Post(ts.URL+"/api/v1/register", "application/json", bytes.NewBuffer(data))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = http.Post(ts.URL, "application/json", bytes.NewBuffer(data))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResponse struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenResponse.Token)
}

func TestAddExpression(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(addExpressionHandler))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "?expression=2+2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	expressionID := string(body)
	assert.NotEmpty(t, expressionID)

	time.Sleep(1 * time.Second)

	resp, err = http.Get(ts.URL + "/expression?id=" + expressionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var expression Expression
	err = json.NewDecoder(resp.Body).Decode(&expression)
	assert.NoError(t, err)
	assert.Equal(t, "2+2", expression.Expression)
	assert.Equal(t, float64(4), expression.Result)
	assert.Equal(t, "completed", expression.Status)
}

func TestListExpressions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(listExpressionsHandler))
	defer ts.Close()

	for i := 0; i < 5; i++ {
		resp, err := http.Get(ts.URL + "/add?expression=2+2")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)

		expressionID := string(body)
		assert.NotEmpty(t, expressionID)

		time.Sleep(1 * time.Second)
	}

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var expressions []Expression
	err = json.NewDecoder(resp.Body).Decode(&expressions)
	assert.NoError(t, err)
	assert.Len(t, expressions, 5)

	for _, expression := range expressions {
		assert.Equal(t, "completed", expression.Status)
		assert.Equal(t, float64(4), expression.Result)
	}
}
