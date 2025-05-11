# Параллельный калькулятор на GO

## Краткое описание писание
Данный проект представляет из себя систему, вычисляющую сложные арифметичсекие выражения. Состоит из оркестратора и агентов, выполняющих роль вычислителей. Выражения хранятся в специальной базе данных. Также пристутсвует регистрация и аутентификация с помощью JWT стандарта.

---

## Установка и запуск

### Шаг 1: Клонировать репозиторий
Вводим в консоль данную команду
```bash
git clone https://github.com/zakharkaverin1/calculator
```

### Шаг 2
```bash
cd calculator
```

### Шаг 3: Установка зависимостей 
```bash
go mod download
```

### Шаг 4: Запускаем оркестратор
```bash
go run .\cmd\orchestrator\main.go
```

### Шаг 5: Открываем вторую консоль
![image](https://github.com/user-attachments/assets/e54daca0-b395-4f3c-ae91-5da4ee645ecf)

### Шаг 6: 
```bash
cd calculator
```

### Шаг 7: Запускаем агента
```bash
go run .\cmd\agent\main.go
```

---

## Архитектура проекта

![dgrm](https://github.com/user-attachments/assets/75c2c4ff-ffaf-4214-b283-2c5ec9a5d5b5)

### Оркестратор
  - принимает выражения
  - разбивает выражения на подзадачи
  - управляет задачами
  - собирает результаты
### Агенты
  - берут задачи с помощью http-запросов
  - вычисляют
  - отправляют результаты на сервер

### Возможности 
  + регистрация и аутентификация
  + вычисление сложных арифметических выражений с использованием сложения, вычитания, умножения и деления
  + параллельное вычисление некоторых подзадач
  + никто, кроме вас, не может смотреть ваши запросы

---

# API Эндпоинты

## Аутентификация

###  Регистрация
**POST** `/register`

```json
{
  "username": "testuser",
  "password": "testpass"
}
```

**curl:**
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}'
```

---

### Логин
**POST** `/login`

**curl:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}'
```

**Ответ:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn..."
}
```

---

## Работа с выражениями

> ❗ Все ниже указанные запросы требуют заголовок `Authorization: Bearer <JWT_TOKEN>`

---

### Отправить выражение
**POST** `/expressions`

```json
{
  "expression": "2 + 3 * 4"
}
```

**curl:**
```bash
curl -X POST http://localhost:8080/expressions \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"expression": "2 + 3 * 4"}'
```

**Ответ:**
```json
{
  "id": 1,
  "expression": "2 + 3 * 4",
  "result": 14
}
```

---

### 📋 Получить все выражения
**GET** `/expressions`

**curl:**
```bash
curl -X GET http://localhost:8080/expressions \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

### Получить выражение по ID
**GET** `/expressions/{id}`

**curl:**
```bash
curl -X GET http://localhost:8080/expressions/1 \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Возможные ответы:**
- `200 OK` — если выражение найдено и принадлежит пользователю
- `403 Forbidden` — если чужое выражение
- `404 Not Found` — если не существует
```

---

---
