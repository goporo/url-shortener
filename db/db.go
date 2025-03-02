package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// Database struct holds the database connection
type Database struct {
	conn *sql.DB
}

// InitDB establishes a connection to the PostgreSQL database
func InitDB() (*Database, error) {
	// Get connection details from environment variables
	// Map from existing environment variables to our expected variables
	dbHost := os.Getenv("DATABASE_HOST")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")
	dbPort := os.Getenv("DATABASE_PORT")
	sslMode := os.Getenv("DATABASE_SSLMODE")

	if dbPort == "" {
		dbPort = "5432" // Default PostgreSQL port
	}

	if sslMode == "" {
		sslMode = "require" // Default to require SSL
	}

	// PostgreSQL connection string format
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)

	// Connect to the database
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Create the database structure
	if err := createSchema(conn); err != nil {
		return nil, fmt.Errorf("error creating schema: %w", err)
	}

	return &Database{conn: conn}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.conn.Close()
}

// createSchema creates the necessary tables if they don't exist
func createSchema(db *sql.DB) error {
	// PostgreSQL uses SERIAL type for auto-incrementing IDs
	// and uses single quotes for string literals in CREATE statements
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		original TEXT NOT NULL,
		short_code VARCHAR(64) NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		access_count INTEGER DEFAULT 0
	)` // Changed "clicks" to "access_count"

	_, err := db.Exec(query)
	return err
}

// GetURLByShortCode retrieves a URL by its short code
func (db *Database) GetURLByShortCode(shortCode string) (*URL, error) {
	var url URL
	query := `SELECT id, original, short_code, created_at, updated_at, access_count 
			  FROM urls WHERE short_code = $1` // Changed "clicks" to "access_count"

	err := db.conn.QueryRow(query, shortCode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.Clicks, // Maps "access_count" to Clicks
	)

	if err != nil {
		return nil, err
	}

	return &url, nil
}

// IncrementClickCount increments the click count for a short URL
func (db *Database) IncrementClickCount(shortCode string) error {
	query := `UPDATE urls SET access_count = access_count + 1 WHERE short_code = $1` // Changed "clicks" to "access_count"
	_, err := db.conn.Exec(query, shortCode)
	return err
}

// URL represents a shortened URL in the database
type URL struct {
	ID          int    `json:"id"`
	OriginalURL string `json:"original"`
	ShortCode   string `json:"shortCode"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Clicks      int    `json:"clicks"` // This maps to access_count in the database
}

// CreateShortURL stores a new short URL
func (db *Database) CreateShortURL(originalURL, shortCode string) error {
	query := `INSERT INTO urls (original, short_code, created_at, updated_at, access_count)
			  VALUES ($1, $2, NOW(), NOW(), 0)` // Changed "clicks" to "access_count" to match existing schema
	_, err := db.conn.Exec(query, originalURL, shortCode)
	return err
}

// GetAllURLs retrieves all stored URLs, ordered by update time
func (db *Database) GetAllURLs(limit int) ([]URL, error) {
	query := `SELECT id, original, short_code, created_at, updated_at, access_count
              FROM urls ORDER BY updated_at DESC LIMIT $1` // Changed "clicks" to "access_count"

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

// UpdateURL updates an existing URL
func (db *Database) UpdateURL(shortCode, newOriginalURL string) error {
	query := `UPDATE urls SET original = $1, updated_at = NOW() WHERE short_code = $2` // PostgreSQL uses $1, $2 for parameters
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

// DeleteURL deletes a URL by its short code
func (db *Database) DeleteURL(shortCode string) error {
	query := `DELETE FROM urls WHERE short_code = $1` // PostgreSQL uses $1 for parameters
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
