package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Database struct {
	conn *sql.DB
}

func InitDB() (*Database, error) {
	dbHost := os.Getenv("DATABASE_HOST")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")
	dbPort := os.Getenv("DATABASE_PORT")
	sslMode := os.Getenv("DATABASE_SSLMODE")

	if dbPort == "" {
		dbPort = "5432"
	}

	if sslMode == "" {
		sslMode = "require"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err := createSchema(conn); err != nil {
		return nil, fmt.Errorf("error creating schema: %w", err)
	}

	return &Database{conn: conn}, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		original TEXT NOT NULL,
		short_code VARCHAR(64) NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		access_count INTEGER DEFAULT 0
	)`

	_, err := db.Exec(query)
	return err
}

func (db *Database) GetURLByShortCode(shortCode string) (*URL, error) {
	var url URL
	query := `SELECT id, original, short_code, created_at, updated_at, access_count 
			  FROM urls WHERE short_code = $1`

	err := db.conn.QueryRow(query, shortCode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.Clicks,
	)

	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (db *Database) IncrementClickCount(shortCode string) error {
	query := `UPDATE urls SET access_count = access_count + 1 WHERE short_code = $1`
	_, err := db.conn.Exec(query, shortCode)
	return err
}

type URL struct {
	ID          int    `json:"id"`
	OriginalURL string `json:"original"`
	ShortCode   string `json:"shortCode"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Clicks      int    `json:"clicks"`
}

func (db *Database) CreateShortURL(originalURL, shortCode string) error {
	query := `INSERT INTO urls (original, short_code, created_at, updated_at, access_count)
			  VALUES ($1, $2, NOW(), NOW(), 0)`
	_, err := db.conn.Exec(query, originalURL, shortCode)
	return err
}

func (db *Database) GetAllURLs(limit int) ([]URL, error) {
	query := `SELECT id, original, short_code, created_at, updated_at, access_count
              FROM urls ORDER BY updated_at DESC LIMIT $1`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]URL, 0)
	for rows.Next() {
		var url URL
		if err := rows.Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt, &url.Clicks); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (db *Database) UpdateURL(shortCode, newOriginalURL string) error {
	query := `UPDATE urls SET original = $1, updated_at = NOW() WHERE short_code = $2`
	result, err := db.conn.Exec(query, newOriginalURL, shortCode)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no URL found with short code: %s", shortCode)
	}

	return nil
}

func (db *Database) DeleteURL(shortCode string) error {
	query := `DELETE FROM urls WHERE short_code = $1`
	result, err := db.conn.Exec(query, shortCode)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no URL found with short code: %s", shortCode)
	}

	return nil
}
