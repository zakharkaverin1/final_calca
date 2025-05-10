package application

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

var DB *sql.DB

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OutExpression struct {
	ExpressionID string
	Expression   string
	Result       sql.NullString
	StatusID     int
}
type FullExpression struct {
	ExpressionID string         `json:"id"`
	Expression   string         `json:"expression"`
	Result       sql.NullString `json:"result"`
	StatusID     int            `json:"status_id"`
	UserID       string         `json:"user_id"`
}

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return fmt.Errorf("ошибка открытия БД: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	if err = createTables(DB); err != nil {
		return fmt.Errorf("ошибка создания таблиц: %v", err)
	}

	return nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS statuses (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS users (
            user_id   TEXT PRIMARY KEY,
            login     TEXT NOT NULL UNIQUE,
            password  TEXT NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS expressions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			expression TEXT NOT NULL,
			result TEXT,
			status_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(status_id) REFERENCES statuses(id)
		)`,
		`INSERT OR IGNORE INTO statuses (id, name) VALUES 
			(1, 'cooking'),
			(2, 'in_progress'),
			(3, 'completed'),
			(4, 'error')`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("ошибка создания таблицы: %v, запрос: %s", err, q)
		}
	}
	return nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func UserVerification(DB *sql.DB, login, password string, task int) (string, error) {
	hashedPassword := hashPassword(password)

	var Password string
	var userID string
	err := DB.QueryRow(`SELECT userId, password FROM users WHERE login = ?`, login).Scan(&userID, &Password)
	if err != nil {
		if err == sql.ErrNoRows {
			if task == 1 {
				return userID, nil
			} else {
				return "", fmt.Errorf("юзер не найден")
			}
		}
		return "", fmt.Errorf("юзер не найден")
	}
	if hashedPassword == Password {
		if task == 1 {
			return "", fmt.Errorf("юзер уже зарегестрирован")
		} else {
			return userID, nil
		}

	}

	if task == 1 {
		return "", fmt.Errorf("такой логин уже существует")
	} else {
		return "", fmt.Errorf("неверный пароль")
	}

}
func InsertUser(login, password string) error {
	userId, err := generateRandomID(10)
	if err != nil {
		return err
	}
	hashed := hashPassword(password)

	_, err = DB.Exec(`INSERT INTO users (userId, login, password) VALUES (?, ?, ?)`, userId, login, hashed)
	if err != nil {
		return err
	}
	return nil
}
func generateRandomID(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}

	return string(bytes), nil
}

func Registrate(w http.ResponseWriter, r *http.Request) (string, int, error) {

	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("invalid JSON")
	}

	if req.Login == "" || req.Password == "" {
		return "", http.StatusBadRequest, fmt.Errorf("логин и пароль отсутствуют")
	}

	userID, err := UserVerification(DB, req.Login, req.Password, 1)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	err = InsertUser(req.Login, req.Password)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("ошибка создания нового юзера: %v", err)
	}

	return userID, http.StatusOK, nil
}

func Login(w http.ResponseWriter, r *http.Request) (string, int, error) {
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("invalid JSON")
	}

	if req.Login == "" || req.Password == "" {
		return "", http.StatusBadRequest, fmt.Errorf("логин и пароль отсутствуют")
	}

	userID, err := UserVerification(DB, req.Login, req.Password, 2)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	return userID, http.StatusOK, nil
}

func InsertExpresions(exprID, userID, expression string, statusId int) error {
	_, err := DB.Exec(
		`INSERT INTO expressions (id, user_id, expression, status_id) VALUES (?, ?, ?, ?)`,
		exprID,
		userID,
		expression,
		statusId,
	)
	return err
}
func UpdateExpressionResult(expressionId, result string, statusId int) error {
	var err error

	if result == "" && statusId != 0 {
		_, err = DB.Exec(
			`UPDATE expressions 
               SET status_id = ?, result = NULL 
             WHERE id = ?`,
			statusId,
			expressionId,
		)
	} else if result != "" && statusId == 0 {
		_, err = DB.Exec(
			`UPDATE expressions 
               SET result = ? 
             WHERE id = ?`,
			result,
			expressionId,
		)
	} else if result != "" && statusId != 0 {
		_, err = DB.Exec(
			`UPDATE expressions 
               SET result = ?, status_id = ? 
             WHERE id = ?`,
			result,
			statusId,
			expressionId,
		)
	}

	return err
}
func UpdateExpression(expressionID string, statusID int, result *string) error {

	var query string
	var args []interface{}

	// Формируем запрос в зависимости от параметров
	switch {
	case statusID > 0 && result != nil:
		query = `UPDATE expressions SET status_id = ?, result = ? WHERE id = ?`
		args = []interface{}{statusID, *result, expressionID}
	case statusID > 0:
		query = `UPDATE expressions SET status_id = ? WHERE id = ?`
		args = []interface{}{statusID, expressionID}
	case result != nil:
		query = `UPDATE expressions SET result = ? WHERE id = ?`
		args = []interface{}{*result, expressionID}
	default:
		return fmt.Errorf("nothing to update")
	}

	// Выполняем запрос
	_, err := DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update expression: %v", err)
	}

	return nil
}

func updateGetExpressionStatus(expressionID string, statusID int) {
	result, err := DB.Exec(
		`UPDATE expressions 
         SET status_id = ? 
         WHERE id = ?`,
		statusID,
		expressionID,
	)
	if err != nil {
		fmt.Errorf("ошибка выполнения запроса: %v", err)
	}

	// Проверяем, что запрос затронул строки
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Errorf("ошибка проверки обновленных строк: %v", err)
	}

	if rowsAffected == 0 {
		fmt.Errorf("выражение с ID %s не найдено", expressionID)
	}

}

func GetExpression(db *sql.DB, id, userID string) (map[string]interface{}, error) {
	var (
		expression string
		statusID   int
		result     sql.NullString
		createdAt  string
	)

	err := db.QueryRow(
		`SELECT expression, status_id, result, created_at 
         FROM expressions WHERE id = ? AND user_id = ?`,
		id,
		userID,
	).Scan(&expression, &statusID, &result, &createdAt)
	if err != nil {
		return nil, err
	}

	expr := map[string]interface{}{
		"id":         id,
		"expression": expression,
		"status":     getStatusName(statusID),
		"created_at": createdAt,
	}

	if result.Valid {
		expr["result"] = result.String
	}

	return expr, nil
}

func getStatusName(statusID int) string {
	switch statusID {
	case 1:
		return "cooking"
	case 2:
		return "in_progress"
	case 3:
		return "completed"
	default:
		return "unknown"
	}
}

func FindUserByLogin(login string) (string, error) {
	var userID string
	err := DB.QueryRow(`SELECT user_id FROM users WHERE login = ?`, login).Scan(&userID)
	return userID, err
}

func InsertUserWithID(userID, login, password string) error {
	hashed := hashPassword(password)
	_, err := DB.Exec(
		`INSERT INTO users (user_id, login, password) VALUES (?, ?, ?)`,
		userID, login, hashed,
	)
	return err
}

func FindSaymExpression(expression string) (string, error) {
	var expressionId string
	err := DB.QueryRow(`SELECT expressionId FROM expressions WHERE expression  = ?`, expression).Scan(&expressionId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("выражение не найдено")
		}
		return "", err
	}
	return expressionId, nil
}

func LenExpresions() (int, error) {
	row := DB.QueryRow("SELECT COUNT(*) FROM users")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetExpressionsByUserID(userID string) ([]FullExpression, error) {
	rows, err := DB.Query("SELECT id, expression, result, status_id, user_id FROM expressions WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expressions []FullExpression
	for rows.Next() {
		var e FullExpression
		err := rows.Scan(&e.ExpressionID, &e.Expression, &e.Result, &e.StatusID, &e.UserID)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}
	return expressions, nil
}

func GetExpressionByID(exprID string) (*FullExpression, error) {
	row := DB.QueryRow("SELECT id, expression, result, status_id, user_id FROM expressions WHERE id = ?", exprID)

	var e FullExpression
	err := row.Scan(&e.ExpressionID, &e.Expression, &e.Result, &e.StatusID, &e.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("not found")
		}
		return nil, err
	}
	return &e, nil
}
