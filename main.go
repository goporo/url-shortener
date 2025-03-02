package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"url-shortener/config"
	"url-shortener/db"
	_ "url-shortener/docs" // Import docs for Swagger
	"url-shortener/middleware"
	"url-shortener/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title URL Shortener API
// @version 1.0
// @description API for shortening URLs, managing redirects, and tracking statistics
// @host localhost:8080
// @BasePath /
// @schemes http

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Use the database struct from our db package
var database *db.Database

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
	timestamp := time.Now().UnixNano()
	return base62Encode(int(timestamp % 100000000))
}

// @Summary Create a new short URL
// @Description Creates a new shortened URL for the provided original URL
// @ID createShortURL
// @Accept json
// @Produce json
// @Param body body object true "URL to be shortened"
// @Success 201 {object} models.URL
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /urls [post]
// @Tags urls
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

	url := models.URL{
		Original:    request.URL,
		ShortCode:   shortCode,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		AccessCount: 0,
	}

	// Use the database package to create the short URL
	err := database.CreateShortURL(url.Original, url.ShortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store URL"})
		return
	}

	// We no longer need to call cleanUpOldURLs as that would be handled by the db package

	c.JSON(http.StatusCreated, url)
}

// @Summary Redirect to original URL
// @Description Redirects to the original URL associated with the short code
// @ID getOriginalURL
// @Param shortCode path string true "Short code of the URL"
// @Success 302 "Redirect to original URL"
// @Failure 404 "Short URL not found"
// @Router /urls/{shortCode} [get]
// @Tags urls
func getOriginalURL(c *gin.Context) {
	fmt.Println("getOriginalURL")

	shortCode := c.Param("shortCode")

	// Use the database package to get the URL by short code
	url, err := database.GetURLByShortCode(shortCode)
	if err != nil {
		c.HTML(http.StatusNotFound, "notfound.html", gin.H{
			"message": "Short URL not found",
		})
		return
	}

	// Increment access count
	if err := database.IncrementClickCount(shortCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update access count"})
		return
	}

	c.Redirect(http.StatusFound, url.OriginalURL)
}

// @Summary Update a short URL
// @Description Updates the original URL for an existing short code
// @ID updateShortURL
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code of the URL to update"
// @Param body body object true "New URL"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /urls/{shortCode} [put]
// @Tags urls
func updateShortURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	var request struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Use the database package to update the URL
	if err := database.UpdateURL(shortCode, request.URL); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL updated successfully"})
}

// @Summary Delete a short URL
// @Description Deletes a shortened URL by its short code
// @ID deleteShortURL
// @Produce json
// @Param shortCode path string true "Short code of the URL to delete"
// @Success 200 {object} object{message=string}
// @Failure 404 {object} object{error=string}
// @Router /urls/{shortCode} [delete]
// @Tags urls
func deleteShortURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	// Use the database package to delete the URL
	if err := database.DeleteURL(shortCode); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}

// @Summary Get URL statistics
// @Description Returns statistics for a shortened URL
// @ID getURLStats
// @Produce json
// @Param shortCode path string true "Short code of the URL"
// @Success 200 {object} models.URL
// @Failure 404 {object} object{error=string}
// @Router /urls/{shortCode}/stats [get]
// @Tags urls
func getURLStats(c *gin.Context) {
	shortCode := c.Param("shortCode")

	// Use the database package to get the URL stats
	url, err := database.GetURLByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	// Convert db.URL to models.URL if needed
	urlStats := models.URL{
		ID:          url.ID,
		Original:    url.OriginalURL,
		ShortCode:   url.ShortCode,
		CreatedAt:   parseTime(url.CreatedAt),
		UpdatedAt:   parseTime(url.UpdatedAt),
		AccessCount: url.Clicks,
	}

	c.JSON(http.StatusOK, urlStats)
}

// Helper function to parse time strings
func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// @Summary Get all shortened URLs
// @Description Returns a list of all shortened URLs created within the last 7 days
// @ID getAllShortURLs
// @Produce json
// @Success 200 {array} models.URL
// @Failure 500 {object} object{error=string}
// @Router /urls [get]
// @Tags urls
func getAllShortURLs(c *gin.Context) {
	urlRecords, err := database.GetAllURLs(7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	urls := []models.URL{}

	for _, record := range urlRecords {
		url := models.URL{
			ID:          record.ID,
			Original:    record.OriginalURL,
			ShortCode:   record.ShortCode,
			CreatedAt:   parseTime(record.CreatedAt),
			UpdatedAt:   parseTime(record.UpdatedAt),
			AccessCount: record.Clicks,
		}
		urls = append(urls, url)
	}

	c.JSON(http.StatusOK, urls)
}

func main() {
	var err error

	if err = godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database connection using our db package
	database, err = db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	log.Println("Successfully connected to PostgreSQL database")

	cfg := config.GetDefaultConfig()

	r := gin.Default()

	if cfg.RateLimit.Enabled {
		rateLimiter := middleware.NewRateLimitMiddleware(cfg.RateLimit.RequestsPerMinute)
		r.Use(rateLimiter.Limit)
	}

	r.Use(cors.Default())

	r.LoadHTMLGlob("templates/*")

	// Add Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/urls", getAllShortURLs)
	r.POST("/urls", createShortURL)
	r.GET("/urls/:shortCode", getOriginalURL)
	r.PUT("/urls/:shortCode", updateShortURL)
	r.DELETE("/urls/:shortCode", deleteShortURL)
	r.GET("/urls/:shortCode/stats", getURLStats)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "URL Shortener API", "docs": "/swagger/index.html"})
	})

	log.Println("Server is running on port", port)
	log.Println("Swagger documentation available at: http://localhost:" + port + "/swagger/index.html")
	r.Run(":" + port)
}
