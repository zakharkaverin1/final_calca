package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/joho/godotenv"
)

type Application struct {
}

func New() *Application {
	return &Application{}
}

type Request struct {
	Expression string `json:"expression"`
}

type Id struct {
	Id string `json:"id"`
}

type Orchestrator struct {
	taskList    []Task
	taskQueue   []*Task
	mu          sync.Mutex
	astStore    map[string]*ASTNode
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		taskList:    []Task{},
		taskQueue:   []*Task{},
		astStore:    make(map[string]*ASTNode),
	}
}

type Task struct {
	ID            string   `json:"id"`
	ExprID        string   `json:"-"`
	Arg1          float64  `json:"arg1"`
	Arg2          float64  `json:"arg2"`
	Operation     string   `json:"operation"`
	OperationTime int      `json:"operation_time"`
	Node          *ASTNode `json:"-"`
}

func init() {
	if err := InitDB("app.db"); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println(err, "fssfd")
		log.Fatal("Ошибка загрузки .env файла")
	}
}

func Valid(e string) bool {
	valid_chars := "1234567890+-*/()"
	// чек на посторонние символы и равное кол-во открывающихся и закрывающихся скобок
	c1 := 0
	c2 := 0
	for i := range len(e) {
		if !strings.ContainsRune(valid_chars, rune(e[i])) {
			log.Printf("Невалидные символы")
			return false
		}
		if string(e[i]) == "(" {
			c1 += 1
		} else if string(e[i]) == ")" {
			c2 += 1
		}
	}
	if c1 != c2 {
		log.Printf("Неравное кол-во скобок")
		return false
	}

	// чек на неправильную расстановку
	for i := range len(e) - 1 {
		if (string(e[i]) == "+" || string(e[i]) == "-" || string(e[i]) == "/" || string(e[i]) == "*" || string(e[i]) == "(" || string(e[i]) == ")") && (string(e[i+1]) == "+" || string(e[i+1]) == "*" || string(e[i+1]) == "/" || string(e[i+1]) == "-" || string(e[i+1]) == ")") {
			log.Printf("Невалидные знаки")
			return false
		}
	}
	//чек ласт символ
	if string(e[len(e)-1]) == "+" || string(e[len(e)-1]) == "-" || string(e[len(e)-1]) == "*" || string(e[len(e)-1]) == "/" || string(e[len(e)-1]) == "(" {
		log.Printf("Неверный последний символ")
	}
	return true
}
func (o *Orchestrator) CreateHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "неверный Authorization header", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	claims, err := ParseJWT(tokenStr)
	if err != nil {
		http.Error(w, "неверный токен", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Очистка и валидация выражения
	expr := strings.ReplaceAll(req.Expression, " ", "")
	if !Valid(expr) {
		http.Error(w, "невалидное выражение", http.StatusUnprocessableEntity)
		return
	}
	exprID, _ := generateRandomID(8)

	// Сохранение в БД
	if err := InsertExpresions(exprID, userID, expr, 1); err != nil {
		http.Error(w, "ошибка сервера", http.StatusInternalServerError)
		return
	}
	ast, _ := ParseAST(req.Expression)
	o.astStore[exprID] = ast
	o.ProcessAST(exprID, ast)


	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Id{Id: exprID})
}

func (o *Orchestrator) getTaskHandler(w http.ResponseWriter, _ *http.Request) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.taskQueue) == 0 {
		http.Error(w, `{"error":"таски закончились"}`, http.StatusNotFound)
		return
	}

	task := o.dequeueTask()
	updateGetExpressionStatus(task.ExprID, 2)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (o *Orchestrator) dequeueTask() *Task {
	task := o.taskQueue[0]
	o.taskQueue = o.taskQueue[1:]
	return task
}

func (o *Orchestrator) getAllExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "неверный Authorization header", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	claims, err := ParseJWT(tokenStr)
	if err != nil {
		http.Error(w, "неврный токен", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	expressions, err := GetExpressionsByUserID(userID)
	if err != nil {
		http.Error(w, "нет такого выражения", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expressions)
}

func (o *Orchestrator) getExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "неверный Authorization header", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	claims, err := ParseJWT(tokenStr)
	if err != nil {
		http.Error(w, "неверный токен", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "что-то пошло не так", http.StatusBadRequest)
		return
	}
	exprID := parts[4]

	expr, err := GetExpressionByID(exprID)
	if err != nil {
		http.Error(w, "выражение не существует", http.StatusNotFound)
		return
	}
	if expr.UserID != userID {
		http.Error(w, "отказано в доступе", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expr)
}

func (o *Orchestrator) postTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID string  `json:"task_id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"невалидный json"}`, http.StatusBadRequest)
		return
	}
	r.Body.Close()
	o.mu.Lock()
	task, idx := o.findTaskByID(req.TaskID)
	if task == nil {
		o.mu.Unlock()
		http.Error(w, `{"error":"таск не найден"}`, http.StatusNotFound)
		return
	}
	o.taskList = append(o.taskList[:idx], o.taskList[idx+1:]...)
	o.updateASTNode(task.Node, req.Result)

	root := o.astStore[task.ExprID]

	if root.IsLeaf {
		resultStr := fmt.Sprintf("%f", root.Value)
		if err := UpdateExpressionResult(task.ExprID, resultStr, 3); err != nil {
			o.mu.Unlock()
			http.Error(w, `{"error":"db update failed"}`, http.StatusInternalServerError)
			return
		}
	} else {
		o.ProcessAST(task.ExprID, root)
	}

	o.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}

func (o *Orchestrator) updateASTNode(node *ASTNode, result float64) {
	node.IsLeaf = true
	node.Value = result
}

func (o *Orchestrator) findTaskByID(taskID string) (*Task, int) {
	for i := range o.taskList {
		if o.taskList[i].ID == taskID {
			return &o.taskList[i], i
		}
	}
	return nil, -1
}

func (o *Orchestrator) ProcessAST(exprID string, ast *ASTNode) {
	var traverse func(*ASTNode)
	traverse = func(n *ASTNode) {
		if n == nil || n.IsLeaf {
			return
		}
		traverse(n.Right)
		traverse(n.Left)
		if n.Left != nil && n.Right != nil && n.Left.IsLeaf && n.Right.IsLeaf {
			taskID, _ := generateRandomID(8)
			task := &Task{
				ID:            taskID,
				ExprID:        exprID,
				Arg1:          n.Left.Value,
				Arg2:          n.Right.Value,
				Operation:     n.Operator,
				OperationTime: o.getOperationTime(n.Operator),
				Node:          n,
			}
			o.taskList = append(o.taskList, *task)
			o.taskQueue = append(o.taskQueue, task)
		}
	}
	traverse(ast)
}

func (o *Orchestrator) getOperationTime(operator string) int {
	var envVar string
	switch operator {
	case "+":
		envVar = "TIME_ADDITION_MS"
	case "-":
		envVar = "TIME_SUBTRACTION_MS"
	case "*":
		envVar = "TIME_MULTIPLICATIONS_MS"
	case "/":
		envVar = "TIME_DIVISIONS_MS"
	default:
		return 1000
	}
	timeStr := os.Getenv(envVar)
	if timeStr == "" {
		log.Printf("Переменная %s не сущесвтует", envVar)
		return 1000
	}

	time, err := strconv.Atoi(timeStr)
	if err != nil {
		log.Printf("Невалидные данные %s: %s", envVar, timeStr)
		return 1000
	}

	return time
}

func (o *Orchestrator) Run() error {
	http.HandleFunc("/api/v1/calculate", o.CreateHandler)
	http.HandleFunc("/api/v1/register", o.RegisterHandler)
	http.HandleFunc("/api/v1/login", o.LoginHandler)
	http.HandleFunc("/api/v1/expressions", o.getAllExpressionsHandler)
	http.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			o.getTaskHandler(w, r)
		}
		if r.Method == http.MethodPost {
			o.postTaskHandler(w, r)
		}
	})
	http.HandleFunc("/api/v1/expressions/", o.getExpressionByIDHandler)
	log.Printf("Сервер запущен")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
		return nil
	} else {
		return err
	}

}
