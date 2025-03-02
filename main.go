package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URL struct {
	ID          int       `json:"id"`
	Original    string    `json:"original"`
	ShortCode   string    `json:"shortCode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessCount"`
}

var db *sql.DB

func base62Encode(num int) string {
	encoded := ""
	for num > 0 {
		remainder := num % 62
		encoded = string(chars[remainder]) + encoded
		num /= 62
	}
	return encoded
}

func generateShortCode() string {
	var lastID int
	db.QueryRow("SELECT MAX(id) FROM urls").Scan(&lastID)
	return base62Encode(lastID + 1)
}

func createShortURL(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	shortCode := generateShortCode()
	timestamp := time.Now()

	stmt, err := db.Prepare("INSERT INTO urls (original, short_code, created_at, updated_at, access_count) VALUES (?, ?, ?, ?, 0)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(request.URL, shortCode, timestamp, timestamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store URL"})
		return
	}

	cleanUpOldURLs()

	c.JSON(http.StatusCreated, gin.H{
		"original":  request.URL,
		"shortCode": shortCode,
		"createdAt": timestamp,
		"updatedAt": timestamp,
	})
}

func cleanUpOldURLs() {
	limit := 100

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)
	if err != nil {
		log.Println("Failed to count URLs:", err)
		return
	}

	if count > limit {
		_, err := db.Exec("DELETE FROM urls WHERE id IN (SELECT id FROM urls ORDER BY created_at ASC LIMIT ?)", count-100)
		if err != nil {
			log.Println("Failed to clean up old URLs:", err)
		}
	}
}

func getOriginalURL(c *gin.Context) {
	fmt.Println("getOriginalURL")

	shortCode := c.Param("shortCode")
	var url URL

	err := db.QueryRow("SELECT id, original, short_code, created_at, updated_at, access_count FROM urls WHERE short_code = ?", shortCode).
		Scan(&url.ID, &url.Original, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount)

	if err != nil {
		c.HTML(http.StatusNotFound, "notfound.html", gin.H{
			"message": "Short URL not found",
		})
		return
	}

	// Increment access count
	_, err = db.Exec("UPDATE urls SET access_count = access_count + 1 WHERE short_code = ?", shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update access count"})
		return
	}

	c.Redirect(http.StatusFound, url.Original)
}

func updateShortURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	var request struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	stmt, err := db.Prepare("UPDATE urls SET original = ?, updated_at = ? WHERE short_code = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(request.URL, time.Now(), shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update URL"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL updated successfully"})
}

func deleteShortURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	stmt, err := db.Prepare("DELETE FROM urls WHERE short_code = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete URL"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}

func getURLStats(c *gin.Context) {
	shortCode := c.Param("shortCode")
	var url URL

	err := db.QueryRow("SELECT id, original, short_code, created_at, updated_at, access_count FROM urls WHERE short_code = ?", shortCode).
		Scan(&url.ID, &url.Original, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"original":    url.Original,
		"shortCode":   url.ShortCode,
		"createdAt":   url.CreatedAt,
		"updatedAt":   url.UpdatedAt,
		"accessCount": url.AccessCount,
	})
}

func getAllShortURLs(c *gin.Context) {
	rows, err := db.Query("SELECT id, original, short_code, created_at, updated_at, access_count FROM urls ORDER BY updated_at DESC LIMIT 7")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var url URL
		if err := rows.Scan(&url.ID, &url.Original, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan URL"})
			return
		}
		urls = append(urls, url)
	}

	c.JSON(http.StatusOK, urls)
}

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")

	db, err = sql.Open("sqlite3", "urls.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        original TEXT NOT NULL,
        short_code TEXT NOT NULL UNIQUE,
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL,
        access_count INTEGER DEFAULT 0
    )`)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.Use(cors.Default())

	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/urls", getAllShortURLs)              // Get all shortened URLs
	r.POST("/urls", createShortURL)              // Create a new shortened URL
	r.GET("/urls/:shortCode", getOriginalURL)    // Redirect to the original URL
	r.PUT("/urls/:shortCode", updateShortURL)    // Update a shortened URL
	r.DELETE("/urls/:shortCode", deleteShortURL) // Delete a shortened URL
	r.GET("/urls/:shortCode/stats", getURLStats) // Get stats for a shortened URL

	log.Println("Server is running on port", port)
	r.Run(":" + port)
}
