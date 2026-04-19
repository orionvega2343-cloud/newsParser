# PostgreSQL в Go — полное руководство

## 1. Подключение к базе данных

### Установка драйвера

```bash
go get github.com/lib/pq
```

### Подключение

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func main() {
    connStr := "host=localhost port=5432 user=postgres password=yourpassword dbname=blog sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Проверка соединения
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Подключено к базе данных")
}
```

`sql.Open` не открывает соединение — только валидирует строку подключения.  
`db.Ping()` — вот что реально проверяет соединение.

---

## 2. Создание таблицы

```go
_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS posts (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        text TEXT NOT NULL,
        author TEXT NOT NULL,
        date TEXT NOT NULL
    )
`)
if err != nil {
    log.Fatal(err)
}
```

`SERIAL` — автоинкремент, PostgreSQL сам назначает ID.  
`IF NOT EXISTS` — не падает с ошибкой если таблица уже есть.

---

## 3. Основные операции (CRUD)

### INSERT — создать запись

```go
func CreatePost(db *sql.DB, p models.Post) (models.Post, error) {
    err := db.QueryRow(
        "INSERT INTO posts (title, text, author, date) VALUES ($1, $2, $3, $4) RETURNING id",
        p.Title, p.Text, p.Author, p.Date,
    ).Scan(&p.ID)

    if err != nil {
        return models.Post{}, err
    }
    return p, nil
}
```

`$1, $2, $3` — плейсхолдеры в PostgreSQL (в MySQL это `?`).  
`RETURNING id` — возвращает ID только что созданной записи.  
`Scan(&p.ID)` — записывает результат в переменную.

---

### SELECT все записи

```go
func GetAll(db *sql.DB) ([]models.Post, error) {
    rows, err := db.Query("SELECT id, title, text, author, date FROM posts")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var posts []models.Post
    for rows.Next() {
        var p models.Post
        err := rows.Scan(&p.ID, &p.Title, &p.Text, &p.Author, &p.Date)
        if err != nil {
            return nil, err
        }
        posts = append(posts, p)
    }
    return posts, nil
}
```

`rows.Close()` — всегда закрывай через defer, иначе утечка соединений.  
`rows.Next()` — итерируется по строкам результата.  
`rows.Scan()` — считывает поля строки в переменные. Порядок должен совпадать с SELECT.

---

### SELECT одна запись по ID

```go
func GetByID(db *sql.DB, id int) (models.Post, error) {
    var p models.Post
    err := db.QueryRow(
        "SELECT id, title, text, author, date FROM posts WHERE id = $1", id,
    ).Scan(&p.ID, &p.Title, &p.Text, &p.Author, &p.Date)

    if err == sql.ErrNoRows {
        return models.Post{}, errors.New("not found")
    }
    if err != nil {
        return models.Post{}, err
    }
    return p, nil
}
```

`QueryRow` — для одной строки, не нужен цикл.  
`sql.ErrNoRows` — специальная ошибка когда запись не найдена. Всегда проверяй её отдельно.

---

### UPDATE — обновить запись

```go
func Update(db *sql.DB, id int, p models.Post) (models.Post, error) {
    result, err := db.Exec(
        "UPDATE posts SET title=$1, text=$2, author=$3, date=$4 WHERE id=$5",
        p.Title, p.Text, p.Author, p.Date, id,
    )
    if err != nil {
        return models.Post{}, err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return models.Post{}, err
    }
    if rowsAffected == 0 {
        return models.Post{}, errors.New("not found")
    }

    p.ID = id
    return p, nil
}
```

`RowsAffected()` — сколько строк затронул запрос. Если 0 — запись с таким ID не существует.

---

### DELETE — удалить запись

```go
func Delete(db *sql.DB, id int) error {
    result, err := db.Exec("DELETE FROM posts WHERE id = $1", id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return errors.New("not found")
    }
    return nil
}
```

---

## 4. Передача db в handlers

Создай структуру для хранения зависимостей:

```go
// handlers/handlers.go
type Handler struct {
    DB *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
    return &Handler{DB: db}
}

func (h *Handler) HandlePosts(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        posts, err := storage.GetAll(h.DB)
        // ...
    }
}
```

В `main.go`:

```go
db, _ := sql.Open("postgres", connStr)

h := handlers.NewHandler(db)

http.HandleFunc("/posts", h.HandlePosts)
```

Так `db` не глобальная переменная, а зависимость — это правильная архитектура.

---

## 5. Разница между Query, QueryRow, Exec

| Метод | Когда использовать | Возвращает |
|---|---|---|
| `db.Exec` | INSERT, UPDATE, DELETE | `sql.Result` |
| `db.Query` | SELECT несколько строк | `*sql.Rows` |
| `db.QueryRow` | SELECT одна строка | `*sql.Row` |

---

## 6. Переменные окружения для подключения

Хранить пароль в коде — плохая практика. Используй переменные окружения:

```go
import "os"

connStr := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    os.Getenv("DB_HOST"),
    os.Getenv("DB_PORT"),
    os.Getenv("DB_USER"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_NAME"),
)
```

Запуск:
```bash
DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=secret DB_NAME=blog go run .
```

---

## 7. Структура проекта с БД

```
blog/
├── go.mod
├── main.go
├── models/
│   └── models.go
├── storage/
│   └── storage.go    ← функции принимают *sql.DB
├── handlers/
│   └── handlers.go   ← Handler struct с полем DB
```

---

## 8. Частые ошибки

**Забыл `rows.Close()`** — утечка соединений, БД перестанет отвечать.

**Не проверил `sql.ErrNoRows`** — программа упадёт с непонятной ошибкой вместо 404.

**Порядок в Scan не совпадает с SELECT** — данные попадут в неправильные поля.

**`sql.Open` без `db.Ping()`** — не узнаешь что БД недоступна до первого запроса.