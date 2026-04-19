package storage

import (
	"database/sql"
	"newsParser/models"

	_ "github.com/lib/pq"
)

func NewDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS results (
    id SERIAL PRIMARY KEY,
    header TEXT,
    link TEXT
)`)
	if err != nil {
		return err
	}
	return nil

}

func Insert(db *sql.DB, r models.Result) (models.Result, error) {
	err := db.QueryRow(
		`INSERT INTO results (header, link) VALUES ($1, $2) RETURNING id`, r.Header, r.Link).Scan(&r.ID)
	if err != nil {
		return models.Result{}, err
	}
	return r, nil
}

func GetAll(db *sql.DB) ([]models.Result, error) {
	rows, err := db.Query("SELECT id, header,link FROM results")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Result
	for rows.Next() {
		var r models.Result
		err := rows.Scan(&r.ID, &r.Header, &r.Link)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
