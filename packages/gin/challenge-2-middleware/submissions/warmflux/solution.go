package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// Article represents a blog article
type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// In-memory storage
var articles = []Article{
	{ID: 1, Title: "Getting Started with Go", Content: "Go is a programming language...", Author: "John Doe", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: 2, Title: "Web Development with Gin", Content: "Gin is a web framework...", Author: "Jane Smith", CreatedAt: time.Now(), UpdatedAt: time.Now()},
}
var nextID = 3

func main() {
	// TODO: Create Gin router without default middleware
	// Use gin.New() instead of gin.Default()
	router := gin.New()

	// TODO: Setup custom middleware in correct order
	// 1. ErrorHandlerMiddleware (first to catch panics)
	// 2. RequestIDMiddleware
	// 3. LoggingMiddleware
	// 4. CORSMiddleware
	// 5. RateLimitMiddleware
	// 6. ContentTypeMiddleware
	router.Use(
		ErrorHandlerMiddleware(),
		RequestIDMiddleware(),
		LoggingMiddleware(),
		CORSMiddleware(),
		RateLimitMiddleware(),
		ContentTypeMiddleware(),
	)

	// TODO: Setup route groups
	// Public routes (no authentication required)
	// Protected routes (require authentication)
	public := router.Group("")
	protected := router.Group("")
	protected.Use(AuthMiddleware())

	// TODO: Define routes
	// Public: GET /ping, GET /articles, GET /articles/:id
	// Protected: POST /articles, PUT /articles/:id, DELETE /articles/:id, GET /admin/stats
	{
		public.GET("/ping", ping)
		public.GET("/articles", getArticles)
		public.GET("/articles/:id", getArticle)
	}
	{
		protected.POST("/articles", createArticle)
		protected.PUT("/articles/:id", updateArticle)
		protected.DELETE("/articles/:id", deleteArticle)
		protected.GET("/admin/stats", getStats)
	}

	// TODO: Start server on port 8080
	router.Run(":8081")
}

// TODO: Implement middleware functions

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Generate UUID for request ID
		// Use github.com/google/uuid package
		// Store in context as "request_id"
		// Add to response header as "X-Request-ID"
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware logs all requests with timing information
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Capture start time
		start := time.Now()

		c.Next()

		// TODO: Calculate duration and log request
		duration := time.Since(start)
		// Format: [REQUEST_ID] METHOD PATH STATUS DURATION IP USER_AGENT
		log.Printf("[%s] %s %s %d %v %s %s",
			c.GetString("request_id"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
			c.Request.UserAgent())
	}
}

// AuthMiddleware validates API keys for protected routes
func AuthMiddleware() gin.HandlerFunc {
	// TODO: Define valid API keys and their roles
	// "admin-key-123" -> "admin"
	// "user-key-456" -> "user"
	adminValidKey := "admin"
	userValidKey := "user"

	return func(c *gin.Context) {
		// TODO: Get API key from X-API-Key header
		// TODO: Validate API key
		// TODO: Set user role in context
		// TODO: Return 401 if invalid or missing
		token := c.GetHeader("X-API-Key")
		if token == "" {
			c.JSON(401, APIResponse{
				Success: false,
			})
			c.Abort()
			return
		}
		if strings.Contains(token, adminValidKey) || strings.Contains(token, userValidKey) {
			c.Set("user_role", strings.Split(token, "-")[0])
			c.Next()
		} else {
			c.JSON(401, APIResponse{
				Success: false,
			})
			c.Abort()
			return
		}
	}
}

// CORSMiddleware handles cross-origin requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Set CORS headers
		// Allow origins: http://localhost:3000, https://myblog.com
		// Allow methods: GET, POST, PUT, DELETE, OPTIONS
		// Allow headers: Content-Type, X-API-Key, X-Request-ID
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := map[string]bool{
			"http://localhost:3000": true,
			"https://myblog.com":    true,
		}

		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Request-ID")
		// TODO: Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.Status(204)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting per IP
func RateLimitMiddleware() gin.HandlerFunc {
	// TODO: Implement rate limiting
	// Limit: 100 requests per IP per minute
	// Use golang.org/x/time/rate package
	// Set headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
	// Return 429 if rate limit exceeded
	var (
		ipLimiter = make(map[string]*rate.Limiter)
		mu        sync.RWMutex
		limit     = rate.Every(time.Minute)
		burst     = 100
	)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		mu.RLock()
		limiter, exists := ipLimiter[clientIP]
		mu.RUnlock()
		if !exists {
			mu.Lock()
			limiter = rate.NewLimiter(limit, burst)
			ipLimiter[clientIP] = limiter
			mu.Unlock()
		}
		resetTime := time.Now().Add(time.Minute)
		resetUnix := strconv.FormatInt(resetTime.Unix(), 10)
		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Reset", resetUnix)
		if !limiter.Allow() {
			c.Header("X-RateLimit-Remaining", "0")
			c.JSON(429, APIResponse{
				Success: false,
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Remaining", strconv.Itoa(int(limiter.Tokens())))
		c.Next()
	}
}

// ContentTypeMiddleware validates content type for POST/PUT requests
func ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Check content type for POST/PUT requests
		// Must be application/json
		// Return 415 if invalid content type
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				c.JSON(415, APIResponse{
					Success: false,
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// ErrorHandlerMiddleware handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		resp := APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
			Error:     "Internal server error",
		}
		switch err := recovered.(type) {
		case error:
			resp.Message = err.Error()
		default:
			resp.Message = fmt.Sprintf("%v", recovered)
		}

		c.JSON(500, resp)
	})
}

// TODO: Implement route handlers

// ping handles GET /ping - health check endpoint
func ping(c *gin.Context) {
	// TODO: Return simple pong response with request ID
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   "pong",
		RequestID: c.GetString("request_id"),
	})
}

// getArticles handles GET /articles - get all articles with pagination
func getArticles(c *gin.Context) {
	// TODO: Implement pagination (optional)
	// TODO: Return articles in standard format
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit
	_ = offset

	c.JSON(200, APIResponse{
		Success:   true,
		Data:      articles,
		RequestID: c.GetString("request_id"),
	})
}

// getArticle handles GET /articles/:id - get article by ID
func getArticle(c *gin.Context) {
	// TODO: Get article ID from URL parameter
	// TODO: Find article by ID
	// TODO: Return 404 if not found
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}
	article, index := findArticleByID(id)
	if index == -1 {
		c.JSON(404, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}

	c.JSON(200, APIResponse{
		Success:   true,
		Data:      article,
		RequestID: c.GetString("request_id"),
	})
}

// createArticle handles POST /articles - create new article (protected)
func createArticle(c *gin.Context) {
	// TODO: Parse JSON request body
	// TODO: Validate required fields
	// TODO: Add article to storage
	// TODO: Return created article
	var article Article
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(400, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}
	if err := validateArticle(article); err != nil {
		c.JSON(400, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}
	article.ID = nextID
	nextID++
	articles = append(articles, article)
	c.JSON(201, APIResponse{
		Success:   true,
		Data:      articles,
		RequestID: c.GetString("request_id"),
	})
}

// updateArticle handles PUT /articles/:id - update article (protected)
func updateArticle(c *gin.Context) {
	// TODO: Get article ID from URL parameter
	// TODO: Parse JSON request body
	// TODO: Find and update article
	// TODO: Return updated article
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(404, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}
	var newArticle Article
	if err := c.ShouldBindJSON(&newArticle); err != nil {
		c.JSON(404, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}
	oldArticle, index := findArticleByID(id)
	if index == -1 {
		c.JSON(404, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}

	oldArticle.Author = newArticle.Author
	oldArticle.Content = newArticle.Content
	oldArticle.ID = newArticle.ID
	oldArticle.Title = newArticle.Title

	c.JSON(200, APIResponse{
		Success:   true,
		Data:      newArticle,
		RequestID: c.GetString("request_id"),
	})
}

// deleteArticle handles DELETE /articles/:id - delete article (protected)
func deleteArticle(c *gin.Context) {
	// TODO: Get article ID from URL parameter
	// TODO: Find and remove article
	// TODO: Return success message
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(404, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}

	_, index := findArticleByID(id)
	if index == -1 {
		c.JSON(404, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}
	articles = append(articles[:index], articles[index+1:]...)
	c.JSON(200, APIResponse{
		Success:   true,
		RequestID: c.GetString("request_id"),
	})
}

// getStats handles GET /admin/stats - get API usage statistics (admin only)
func getStats(c *gin.Context) {
	// TODO: Check if user role is "admin"
	// TODO: Return mock statistics
	role := c.GetString("user_role")
	if role != "admin" {
		c.JSON(403, APIResponse{
			Success:   false,
			RequestID: c.GetString("request_id"),
		})
		return
	}
	stats := map[string]interface{}{
		"total_articles": len(articles),
		"total_requests": 0, // Could track this in middleware
		"uptime":         "24h",
	}

	// TODO: Return stats in standard format
	c.JSON(200, APIResponse{
		Success: true,
		Data:    stats,
	})
}

// Helper functions

// findArticleByID finds an article by ID
func findArticleByID(id int) (*Article, int) {
	// TODO: Implement article lookup
	// Return article pointer and index, or nil and -1 if not found
	var index int
	for i, a := range articles {
		if a.ID == id {
			index = i
			return &articles[i], index
		}
	}
	return nil, -1
}

// validateArticle validates article data
func validateArticle(article Article) error {
	// TODO: Implement validation
	// Check required fields: Title, Content, Author
	if article.Title == "" || article.Content == "" || article.Author == "" {
		return errors.New("Error")
	}
	return nil
}