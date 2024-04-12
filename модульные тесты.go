package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	reqBody := strings.NewReader(`{"login": "testuser", "password": "testpassword"}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/register", reqBody)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	registerHandler(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestLoginHandler(t *testing.T) {
	reqBody := strings.NewReader(`{"login": "testuser", "password": "testpassword"}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/login", reqBody)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	loginHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var tokenResponse struct {
		Token string `json:"token"`
	}
	err := json.NewDecoder(rr.Body).Decode(&tokenResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenResponse.Token)
}

func TestAddExpressionHandler(t *testing.T) {
	// Create a mock client for gRPC CalculatorClient
	type mockClient struct {
		pb.UnimplementedCalculatorServer
	}

	func (m *mockClient) Calculate(ctx context.Context, req *pb.Expression) (*pb.Result, error) {
		return &pb.Result{Value: 3}, nil
	}

	// Replace the global client with the mock client
	client = &mockClient{}

	req, _ := http.NewRequest(http.MethodGet, "/add?expression=1+2", nil)
	rr := httptest.NewRecorder()
	addExpressionHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "1", rr.Body.String())

	// Restore the original client
	conn, err := grpc.Dial(":8081", grpc.WithInsecure())
	assert.NoError(t, err)
	client = pb.NewCalculatorClient(conn)
}

func TestGetExpressionHandler(t *testing.T) {
	// Add an expression to the database
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	err := saveExpression(id, "1+2", "waiting", 0)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/expression?id="+id, nil)
	rr := httptest.NewRecorder()
	getExpressionHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var expression Expression
	err = json.NewDecoder(rr.Body).Decode(&expression)
	assert.NoError(t, err)
	assert.Equal(t, id, expression.ID)
	assert.Equal(t, "1+2", expression.Expression)
	assert.Equal(t, "waiting", expression.Status)
	assert.Equal(t, float64(0), expression.Result)
}

func TestListExpressionsHandler(t *testing.T) {
	// Add an expression to the database
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	err := saveExpression(id, "1+2", "waiting", 0)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/list", nil)
	rr := httptest.NewRecorder()
	listExpressionsHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var expressions []Expression
	err = json.NewDecoder(rr.Body).Decode(&expressions)
	assert.NoError(t, err)
	assert.Len(t, expressions, 1)
	assert.Equal(t, id, expressions[0].ID)
	assert.Equal(t, "1+2", expressions[0].Expression)
	assert.Equal(t, "waiting", expressions[0].Status)
	assert.Equal(t, float64(0), expressions[0].Result)
}

func TestListOperationsHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/operations", nil)
	rr := httptest.NewRecorder()
	listOperationsHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var operations []map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&operations)
	assert.NoError(t, err)
	assert.Len(t, operations, 4)
}

func TestGetTaskHandler(t *testing.T) {
	// Add a task to the tasks map
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	task := &Task{
		ID:      id,
		Expr:    "1+2",
		Result:  0,
		IsReady: false,
	}

	mu.Lock()
	tasks[id] = task
	mu.Unlock()

	req, _ := http.NewRequest(http.MethodGet, "/getTask", nil)
	rr := httptest.NewRecorder()
	getTaskHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var taskResponse Task
	err := json.NewDecoder(rr.Body).Decode(&taskResponse)
	assert.NoError(t, err)
	assert.Equal(t, id, taskResponse.ID)
	assert.Equal(t, "1+2", taskResponse.Expr)
	assert.Equal(t, float64(0), taskResponse.Result)
	assert.False(t, taskResponse.IsReady)
}

func TestReceiveResultHandler(t *testing.T) {
	// Add a task to the tasks map
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	task := &Task{
		ID:      id,
		Expr:    "1+2",
		Result:  0,
		IsReady: false,
	}

	mu.Lock()
	tasks[id] = task
	mu.Unlock()

	reqBody := strings.NewReader(fmt.Sprintf(`{"task_id": "%s", "result": 3}`, id))
	req, _ := http.NewRequest(http.MethodPost, "/receiveResult", reqBody)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	receiveResultHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check if the task is updated in the tasks map
	mu.RLock()
	updatedTask, ok := tasks[id]
	mu.RUnlock()
	assert.True(t, ok)
	assert.Equal(t, float64(3), updatedTask.Result)
	assert.True(t, updatedTask.IsReady)

	// Check if the expression is updated in the database
	expression, err := getExpressionFromDB(id)
	assert.NoError(t, err)
	assert.Equal(t, "1+2", expression.Expression)
	assert.Equal(t, "completed", expression.Status)
	assert.Equal(t, float64(3), expression.Result)
}

func TestSaveExpression(t *testing.T) {
	// Test saving an expression to the database
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	err := saveExpression(id, "1+2", "waiting", 0)
	assert.NoError(t, err)

	// Check if the expression is saved in the database
	expression, err := getExpressionFromDB(id)
	assert.NoError(t, err)
	assert.Equal(t, id, expression.ID)
	assert.Equal(t, "1+2", expression.Expression)
	assert.Equal(t, "waiting", expression.Status)
	assert.Equal(t, float64(0), expression.Result)
}

func TestUpdateExpressionResult(t *testing.T) {
	// Add an expression to the database
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	err := saveExpression(id, "1+2", "waiting", 0)
	assert.NoError(t, err)

	// Update the expression result in the database
	err = updateExpressionResult(id, 3, "completed")
	assert.NoError(t, err)

	// Check if the expression is updated in the database
	expression, err := getExpressionFromDB(id)
	assert.NoError(t, err)
	assert.Equal(t, id, expression.ID)
	assert.Equal(t, "1+2", expression.Expression)
	assert.Equal(t, "completed", expression.Status)
	assert.Equal(t, float64(3), expression.Result)
}

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hashedPassword := hashPassword(password)
	assert.NotEmpty(t, hashedPassword)

	// Hash the password again and check if it's the same
	hashedPassword2 := hashPassword(password)
	assert.Equal(t, hashedPassword, hashedPassword2)
}

func TestGenerateJWT(t *testing.T) {
	claims := &Claims{
		Login: "testuser",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, signedToken)

	// Parse the token and check if the claims are correct
	parsedToken, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, parsedToken)

	parsedClaims, ok := parsedToken.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, "testuser", parsedClaims.Login)
}
