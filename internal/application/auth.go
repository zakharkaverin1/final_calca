package application

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func (o *Orchestrator) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct{ Login, Password string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}
	// 1) Проверяем, что логин свободен
	var exists string
	err := DB.QueryRow(
		`SELECT user_id FROM users WHERE login = ?`, req.Login,
	).Scan(&exists)
	if err != sql.ErrNoRows {
		// либо нашли строку (err==nil) — Conflict,
		// либо другая ошибка — 500
		if err == nil {
			http.Error(w, "login already exists", http.StatusConflict)
		} else {
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	// 2) Генерируем user_id и хешируем пароль
	userID, err := generateRandomID(10)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	hashed := hashPassword(req.Password)

	// 3) Вставляем нового пользователя
	if _, err := DB.Exec(
		`INSERT INTO users (user_id, login, password) VALUES (?, ?, ?)`,
		userID, req.Login, hashed,
	); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// 4) Отдаём JWT
	token, err := GenerateJWT(userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})

}

func (o *Orchestrator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct{ Login, Password string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	var dbHash, userID string
    err := DB.QueryRow(
        `SELECT user_id, password FROM users WHERE login = ?`, req.Login,
    ).Scan(&userID, &dbHash)
    if err != nil {
        http.Error(w, "wrong login/password", http.StatusUnauthorized)
        return
    }

    // 2) Сравниваем хеши
    if hashPassword(req.Password) != dbHash {
        http.Error(w, "wrong login/password", http.StatusUnauthorized)
        return
    }

    // 3) Отдаём JWT
    token, err := GenerateJWT(userID)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": token})

}
