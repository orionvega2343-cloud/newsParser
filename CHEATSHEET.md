# SQL шпаргалка для Go

## Какой метод Go использовать

| Ситуация | Метод Go |
|---|---|
| SELECT несколько строк | `db.Query` |
| SELECT одна строка | `db.QueryRow` |
| INSERT без RETURNING | `db.Exec` |
| INSERT с RETURNING id | `db.QueryRow` |
| UPDATE | `db.Exec` |
| DELETE | `db.Exec` |

---

## SELECT

### Все записи
```sql
SELECT id, title, text FROM posts
```
```go
rows, err := db.Query("SELECT id, title, text FROM posts")
defer rows.Close()
for rows.Next() {
    var p models.Post
    rows.Scan(&p.ID, &p.Title, &p.Text)
}
```

### По условию
```sql
SELECT id, title FROM posts WHERE author = $1
```
```go
rows, err := db.Query("SELECT id, title FROM posts WHERE author = $1", author)
```

### Одна запись по ID
```sql
SELECT id, title FROM posts WHERE id = $1
```
```go
err := db.QueryRow("SELECT id, title FROM posts WHERE id = $1", id).Scan(&p.ID, &p.Title)
if err == sql.ErrNoRows {
    // не найдено
}
```

### Сортировка
```sql
SELECT id, title FROM posts ORDER BY date DESC
SELECT id, title FROM posts ORDER BY title ASC
```

### Лимит
```sql
SELECT id, title FROM posts LIMIT 10
SELECT id, title FROM posts LIMIT 10 OFFSET 20  -- страница 3
```

---

## INSERT

### Без возврата ID
```sql
INSERT INTO posts (title, text, author) VALUES ($1, $2, $3)
```
```go
_, err := db.Exec("INSERT INTO posts (title, text, author) VALUES ($1, $2, $3)",
    p.Title, p.Text, p.Author)
```

### С возвратом ID
```sql
INSERT INTO posts (title, text, author) VALUES ($1, $2, $3) RETURNING id
```
```go
err := db.QueryRow("INSERT INTO posts (title, text, author) VALUES ($1, $2, $3) RETURNING id",
    p.Title, p.Text, p.Author).Scan(&p.ID)
```

---

## UPDATE

### Обновить все поля
```sql
UPDATE posts SET title=$1, text=$2, author=$3 WHERE id=$4
```
```go
result, err := db.Exec("UPDATE posts SET title=$1, text=$2 WHERE id=$3",
    p.Title, p.Text, id)
rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    // запись не найдена
}
```

### Обновить одно поле
```sql
UPDATE posts SET title=$1 WHERE id=$2
```

---

## DELETE

```sql
DELETE FROM posts WHERE id = $1
```
```go
result, err := db.Exec("DELETE FROM posts WHERE id = $1", id)
rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    // запись не найдена
}
```

---

## Типы данных PostgreSQL

| Go тип | PostgreSQL тип |
|---|---|
| `int` | `INTEGER` или `SERIAL` (автоинкремент) |
| `string` | `TEXT` или `VARCHAR(255)` |
| `float64` | `NUMERIC` или `FLOAT` |
| `bool` | `BOOLEAN` |
| `time.Time` | `TIMESTAMP` |

---

## Плейсхолдеры

PostgreSQL использует `$1, $2, $3` — не `?` как в MySQL.

```sql
WHERE id = $1 AND author = $2
```

Порядок плейсхолдеров совпадает с порядком аргументов в Go:
```go
db.Query("... WHERE id=$1 AND author=$2", id, author)
```

---

## Частые ошибки

**`sql.ErrNoRows`** — QueryRow ничего не нашёл. Всегда проверяй отдельно:
```go
if err == sql.ErrNoRows {
    return errors.New("not found")
}
```

**Забыл `rows.Close()`** — утечка соединений:
```go
rows, err := db.Query(...)
defer rows.Close() // сразу после проверки err
```

**Порядок Scan не совпадает с SELECT** — данные попадут не в те поля:
```go
// SELECT id, title, text
rows.Scan(&p.ID, &p.Title, &p.Text) // порядок должен совпадать
```