//Агент
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "path/to/calculator"
)

type server struct{}

func (s *server) Calculate(ctx context.Context, req *pb.Expression) (*pb.Result, error) {
	switch req.Operation {
	case "+":
		return &pb.Result{Value: req.Operand1 + req.Operand2}, nil
	case "-":
		return &pb.Result{Value: req.Operand1 - req.Operand2}, nil
	case "*":
		return &pb.Result{Value: req.Operand1 * req.Operand2}, nil
	case "/":
		if req.Operand2 == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &pb.Result{Value: req.Operand1 / req.Operand2}, nil
	}
	return nil, fmt.Errorf("invalid operation")
}

func main() {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCalculatorServer(s, &server{})
	fmt.Println("Agent is running on port 8081")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//Оркестратор
package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	pb "path/to/calculator"
	_ "github.com/mattn/go-sqlite3"
)

type Expression struct {
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Status     string  `json:"status"`
	Result     float64 `json:"result"`
}

type Task struct {
	ID      string
	Expr    string
	Result  float64
	IsReady bool
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Claims struct {
	Login string `json:"login"`
	jwt.StandardClaims
}

var (
	db           *sql.DB
	tasks        = make(map[string]*Task)
	mu           sync.RWMutex
	agents       = []string{"localhost:8081", "localhost:8082", /* other agent URLs */}
	conn         *grpc.ClientConn
	client       pb.CalculatorClient
	expressionID int
	jwtKey       = []byte("your_secret_key")
)

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./expressions.db")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS expressions (
		id INTEGER PRIMARY KEY,
		expression TEXT NOT NULL,
		status TEXT NOT NULL,
		result REAL
	);
	CREATE TABLE IF NOT EXISTS users (
		login TEXT PRIMARY KEY,
		password TEXT NOT NULL
	);
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func init() {
	initDB()

	var err error
	conn, err = grpc.Dial(":8081", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client = pb.NewCalculatorClient(conn)
}

func hashPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword := hashPassword(user.Password)

	_, err = db.Exec("INSERT INTO users (login, password) VALUES (?, ?)", user.Login, hashedPassword)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var login, password string
	err = db.QueryRow("SELECT login, password FROM users WHERE login = ?", user.Login).Scan(&login, &password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if hashPassword(user.Password) != password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	claims := &Claims{
		Login: login,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": signedToken})
}

func addExpressionHandler(w http.ResponseWriter, r *http.Request) {
	expr := r.URL.Query().Get("expression")
	if expr == "" {
		http.Error(w, "Empty expression is not allowed", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	task := &Task{
		ID:      id,
		Expr:    expr,
		Result:  0,
		IsReady: false,
	}

	mu.Lock()
	tasks[id] = task
	mu.Unlock()

	go calculateExpression(task)

	err := saveExpression(id, expr, "waiting", 0)
	if err != nil {
		http.Error(w, "Failed to save expression: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, id)
}

func calculateExpression(task *Task) {
	parts := strings.Split(task.Expr, " ")
	operation := parts[0]
	operand1, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		handleError(task, "Invalid operand1: "+err.Error())
		return
	}
	operand2, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		handleError(task, "Invalid operand2: "+err.Error())
		return
	}

	// Choose a random agent from the list
	selectedAgent := agents[rand.Intn(len(agents))]

	resp, err := client.Calculate(context.Background(), &pb.Expression{
		Operation: operation,
		Operand1:  operand1,
		Operand2:  operand2,
	})
	if err != nil {
		handleError(task, "Error sending expression to agent: "+err.Error())
		return
	}

	mu.Lock()
	task.Result = resp.Value
	task.IsReady = true
	mu.Unlock()

	err = updateExpressionResult(task.ID, resp.Value, "completed")
	if err != nil {
		log.Printf("Failed to update expression result: %v", err)
	}
}

func handleError(task *Task, errorMsg string) {
	fmt.Println(errorMsg)
	mu.Lock()
	task.Result = 0
	task.IsReady = true
	mu.Unlock()

	err := updateExpressionResult(task.ID, 0, "error")
	if err != nil {
		log.Printf("Failed to update expression result: %v", err)
	}
}

func getExpressionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	expression, err := getExpressionFromDB(id)
	if err != nil {
		http.Error(w, "Expression not found: "+err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(expression)
}

func listExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	expressions, err := getExpressionsFromDB()
	if err != nil {
		http.Error(w, "Failed to get expressions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(expressions)
}

func listOperationsHandler(w http.ResponseWriter, r *http.Request) {
	type Operation struct {
		Name string `json:"name"`
		Time int    `json:"time"`
	}

	operations := []Operation{
		{Name: "addition", Time: 5},
		{Name: "subtraction", Time: 5},
		{Name: "multiplication", Time: 10},
		{Name: "division", Time: 10},
	}

	json.NewEncoder(w).Encode(operations)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	for _, task := range tasks {
		if !task.IsReady {
			json.NewEncoder(w).Encode(task)
			return
		}
	}
	http.Error(w, "No tasks available", http.StatusNotFound)
}

func receiveResultHandler(w http.ResponseWriter, r *http.Request) {
	var result struct {
		TaskID string  `json:"task_id"`
		Result float64 `json:"result"`
	}
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		http.Error(w, "Failed to decode result", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	task, ok := tasks[result.TaskID]
	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	task.Result = result.Result
	task.IsReady = true

	err = updateExpressionResult(task.ID, result.Result, "completed")
	if err != nil {
		http.Error(w, "Failed to update expression result: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/api/v1/register", registerHandler)
	http.HandleFunc("/api/v1/login", loginHandler)
	http.HandleFunc("/add", addExpressionHandler)
	http.HandleFunc("/expression", getExpressionHandler)
	http.HandleFunc("/list", listExpressionsHandler)
	http.HandleFunc("/operations", listOperationsHandler)
	http.HandleFunc("/getTask", getTaskHandler)
	http.HandleFunc("/receiveResult", receiveResultHandler)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}