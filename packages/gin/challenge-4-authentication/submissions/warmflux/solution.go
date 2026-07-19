package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID             int        `json:"id"`
	Username       string     `json:"username" binding:"required,min=3,max=30"`
	Email          string     `json:"email" binding:"required,email"`
	Password       string     `json:"-"` // Never return in JSON
	PasswordHash   string     `json:"-"`
	FirstName      string     `json:"first_name" binding:"required,min=2,max=50"`
	LastName       string     `json:"last_name" binding:"required,min=2,max=50"`
	Role           string     `json:"role"`
	IsActive       bool       `json:"is_active"`
	EmailVerified  bool       `json:"email_verified"`
	LastLogin      *time.Time `json:"last_login"`
	FailedAttempts int        `json:"-"`
	LockedUntil    *time.Time `json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Username        string `json:"username" binding:"required,min=3,max=30"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	FirstName       string `json:"first_name" binding:"required,min=2,max=50"`
	LastName        string `json:"last_name" binding:"required,min=2,max=50"`
}

// TokenResponse represents JWT token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// APIResponse represents standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Global data stores (in a real app, these would be databases)
var (
	users             = []User{}
	blacklistedTokens = make(map[string]bool) // Token blacklist for logout
	refreshTokens     = make(map[string]int)  // RefreshToken -> UserID mapping
	nextUserID        = 1
)

// Configuration
var (
	jwtSecret         = []byte("your-super-secret-jwt-key")
	accessTokenTTL    = 15 * time.Minute   // 15 minutes
	refreshTokenTTL   = 7 * 24 * time.Hour // 7 days
	maxFailedAttempts = 5
	lockoutDuration   = 30 * time.Minute
)

// User roles
const (
	RoleUser      = "user"
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
)

// TODO: Implement password strength validation
func isStrongPassword(password string) bool {
	// TODO: Validate password strength:
	// - At least 8 characters
	// - Contains uppercase letter
	// - Contains lowercase letter
	// - Contains number
	// - Contains special character
	lowerReg := `[a-z]`
	upperReg := `[A-Z]`
	digitReg := `[0-9]`
	spacialReg := `[^a-zA-Z0-9]`
	length := len(password)
	containsLower, _ := regexp.MatchString(lowerReg, password)
	containsUpper, _ := regexp.MatchString(upperReg, password)
	containsDigit, _ := regexp.MatchString(digitReg, password)
	containsSpecial, _ := regexp.MatchString(spacialReg, password)

	if length >= 8 && containsDigit && containsLower && containsSpecial && containsUpper {
		return true
	}
	return false
}

// TODO: Implement password hashing
func hashPassword(password string) (string, error) {
	// TODO: Use bcrypt to hash the password with cost 12
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), nil
}

// TODO: Implement password verification
func verifyPassword(password, hash string) bool {
	// TODO: Use bcrypt to compare password with hash
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err == nil {
		return true
	}
	return false
}

func generateTokens(userID int, username, role string) (*TokenResponse, error) {
	// TODO: Generate access token with 15 minute expiry
	// TODO: Generate refresh token with 7 day expiry
	// TODO: Store refresh token in memory store
	now := time.Now()
	accessClaims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "gin-auth-api",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := generateRandomToken()
	if err != nil {
		return nil, err
	}

	refreshTokens[refreshTokenString] = userID

	return &TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
		ExpiresAt:    now.Add(accessTokenTTL),
	}, nil
}

// TODO: Implement JWT token validation
func validateToken(tokenString string) (*JWTClaims, error) {
	// TODO: Parse and validate JWT token
	// TODO: Check if token is blacklisted
	// TODO: Return claims if valid
	if blacklistedTokens[tokenString] {
		return nil, fmt.Errorf("token is blacklisted")
	}
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}
	return claims, nil
}

// TODO: Implement user lookup functions
func findUserByUsername(username string) *User {
	// TODO: Find user by username in users slice
	for i, u := range users {
		if username == u.Username {
			return &users[i]
		}
	}
	return nil
}

func findUserByEmail(email string) *User {
	// TODO: Find user by email in users slice
	for i, u := range users {
		if email == u.Email {
			return &users[i]
		}
	}
	return nil
}

func findUserByID(id int) *User {
	// TODO: Find user by ID in users slice
	for i, u := range users {
		if id == u.ID {
			return &users[i]
		}
	}
	return nil
}

// TODO: Implement account lockout check
func isAccountLocked(user *User) bool {
	// TODO: Check if account is locked based on LockedUntil field
	if user.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*user.LockedUntil) || time.Now().Equal(*user.LockedUntil)
}

// TODO: Implement failed attempt tracking
func recordFailedAttempt(user *User) {
	// TODO: Increment failed attempts counter
	// TODO: Lock account if max attempts reached
	user.FailedAttempts++
	if user.FailedAttempts >= maxFailedAttempts {
		user.IsActive = false
	}
}

func resetFailedAttempts(user *User) {
	// TODO: Reset failed attempts counter and unlock account
	user.FailedAttempts = 0
	user.IsActive = true
}

// TODO: Generate secure random token
func generateRandomToken() (string, error) {
	// TODO: Generate cryptographically secure random token
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// POST /auth/register - User registration
func register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// TODO: Validate password confirmation
	if req.Password != req.ConfirmPassword {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Passwords do not match",
		})
		return
	}

	// TODO: Validate password strength
	if !isStrongPassword(req.Password) {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Password does not meet strength requirements",
		})
		return
	}

	// TODO: Check if username already exists
	// TODO: Check if email already exists
	// TODO: Hash password
	// TODO: Create user and add to users slice
	for _, u := range users {
		if u.Username == req.Username {
			c.JSON(409, APIResponse{
				Success: false,
				Error:   "Username already exists",
			})
			return
		}
	}
	for _, u := range users {
		if u.Email == req.Email {
			c.JSON(409, APIResponse{
				Success: false,
				Error:   "Email already exists",
			})
			return
		}
	}

	hash, _ := hashPassword(req.Password)
	newUser := User{
		ID:           nextUserID,
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		PasswordHash: hash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}
	nextUserID++
	users = append(users, newUser)

	c.JSON(201, APIResponse{
		Success: true,
		Message: "User registered successfully",
	})
}

// POST /auth/login - User login
func login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Invalid credentials format",
		})
		return
	}

	// TODO: Find user by username
	user := findUserByUsername(req.Username)
	if user == nil {
		c.JSON(401, APIResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	// TODO: Check if account is locked
	if isAccountLocked(user) {
		c.JSON(423, APIResponse{
			Success: false,
			Error:   "Account is temporarily locked",
		})
		return
	}

	// TODO: Verify password
	if !verifyPassword(req.Password, user.PasswordHash) {
		recordFailedAttempt(user)
		c.JSON(401, APIResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	// TODO: Reset failed attempts on successful login
	resetFailedAttempts(user)

	// TODO: Update last login time
	now := time.Now()
	user.LastLogin = &now

	// TODO: Generate tokens
	tokens, err := generateTokens(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(500, APIResponse{
			Success: false,
			Error:   "Failed to generate tokens",
		})
		return
	}

	c.JSON(200, APIResponse{
		Success: true,
		Data:    tokens,
		Message: "Login successful",
	})
}

// POST /auth/logout - User logout
func logout(c *gin.Context) {
	// TODO: Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(401, APIResponse{
			Success: false,
			Error:   "Authorization header required",
		})
		return
	}

	// TODO: Extract token from "Bearer <token>" format
	// TODO: Add token to blacklist
	// TODO: Remove refresh token from store
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
	if token == "" {
		c.JSON(401, APIResponse{
			Success: false,
			Error:   "Bearer header required",
		})
		return
	}

	blacklistedTokens[token] = true
	delete(refreshTokens, token)

	c.JSON(200, APIResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// POST /auth/refresh - Refresh access token
func refreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Refresh token required",
		})
		return
	}

	// TODO: Validate refresh token
	// TODO: Get user ID from refresh token store
	// TODO: Find user by ID
	// TODO: Generate new access token
	// TODO: Optionally rotate refresh token
	userID, ok := refreshTokens[req.RefreshToken]
	if !ok {
		c.JSON(401, APIResponse{
			Success: false,
			Error:   "Invalid refresh token",
		})
		return
	}

	user := findUserByID(userID)
	if user == nil {
		c.JSON(401, APIResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	tokens, err := generateTokens(userID, user.Username, user.Role)
	if err != nil {
		c.JSON(500, APIResponse{
			Success: false,
			Error:   "Failed to generate tokens",
		})
		return
	}

	c.JSON(200, APIResponse{
		Success: true,
		Data:    tokens,
		Message: "Token refreshed successfully",
	})
}

// Middleware: JWT Authentication
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, APIResponse{
				Success: false,
				Error:   "Authorization header required",
			})
			c.Abort()
			return
		}

		// TODO: Extract token from "Bearer <token>" format
		// TODO: Validate token using validateToken function
		// TODO: Set user info in context for route handlers
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
		claims, err := validateToken(token)
		if err != nil {
			c.JSON(401, APIResponse{
				Success: false,
				Error:   "Invalid token",
			})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("userName", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// Middleware: Role-based authorization
func requireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Get user role from context (set by authMiddleware)
		// TODO: Check if user role is in allowed roles
		// TODO: Return 403 if not authorized
		role := c.GetString("role")
		if role == "" {
			c.JSON(401, APIResponse{
				Success: false,
				Error:   "Unauthorized",
			})
			c.Abort()
			return
		}
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		c.JSON(403, APIResponse{
			Success: false,
			Error:   "Forbidden",
		})
		c.Abort()
	}
}

// GET /user/profile - Get current user profile
func getUserProfile(c *gin.Context) {
	// TODO: Get user ID from context (set by authMiddleware)
	// TODO: Find user by ID
	// TODO: Return user profile (without sensitive data)
	idStr, ok := c.Get("userID")
	if !ok {
		c.JSON(400, APIResponse{
			Success: false,
		})
		return
	}

	findUserByID(idStr.(int))

	c.JSON(200, APIResponse{
		Success: true,
		Data:    nil, // TODO: Return user data
		Message: "Profile retrieved successfully",
	})
}

// PUT /user/profile - Update user profile
func updateUserProfile(c *gin.Context) {
	var req struct {
		FirstName string `json:"first_name" binding:"required,min=2,max=50"`
		LastName  string `json:"last_name" binding:"required,min=2,max=50"`
		Email     string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// TODO: Get user ID from context
	// TODO: Find user by ID
	// TODO: Check if new email is already taken
	// TODO: Update user profile
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(400, APIResponse{
			Success: false,
		})
	}
	user := findUserByID(userID.(int))
	if req.Email == "" {
		c.JSON(400, APIResponse{
			Success: false,
		})
	}

	user.Email = req.Email

	c.JSON(200, APIResponse{
		Success: true,
		Message: "Profile updated successfully",
	})
}

// POST /user/change-password - Change user password
func changePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// TODO: Get user ID from context
	// TODO: Find user by ID
	// TODO: Verify current password
	// TODO: Validate new password strength
	// TODO: Hash new password and update user
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(400, APIResponse{
			Success: false,
		})
		return
	}
	user := findUserByID(userID.(int))
	if !verifyPassword(req.CurrentPassword, user.PasswordHash) {
		c.JSON(400, APIResponse{
			Success: false,
		})
		return
	}

	if !isStrongPassword(req.NewPassword) {
		c.JSON(400, APIResponse{
			Success: false,
		})
		return
	}

	hash, _ := hashPassword(req.NewPassword)
	user.Password = req.NewPassword
	user.PasswordHash = hash

	c.JSON(200, APIResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// GET /admin/users - List all users (admin only)
func listUsers(c *gin.Context) {
	// TODO: Get pagination parameters
	// TODO: Return list of users (without sensitive data)

	c.JSON(200, APIResponse{
		Success: true,
		Data:    users, // TODO: Filter sensitive data
		Message: "Users retrieved successfully",
	})
}

// PUT /admin/users/:id/role - Change user role (admin only)
func changeUserRole(c *gin.Context) {
	userID := c.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Invalid role data",
		})
		return
	}

	// TODO: Validate role value
	validRoles := []string{RoleUser, RoleAdmin, RoleModerator}
	isValid := false
	for _, role := range validRoles {
		if req.Role == role {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(400, APIResponse{
			Success: false,
			Error:   "Invalid role",
		})
		return
	}

	// TODO: Find user by ID
	// TODO: Update user role
	user := findUserByID(id)
	if user == nil {
		c.JSON(404, APIResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}
	user.Role = req.Role

	c.JSON(200, APIResponse{
		Success: true,
		Message: "User role updated successfully",
	})
}

// Setup router with authentication routes
func setupRouter() *gin.Engine {
	router := gin.Default()

	// Public routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", register)
		auth.POST("/login", login)
		auth.POST("/logout", logout)
		auth.POST("/refresh", refreshToken)
	}

	// Protected user routes
	user := router.Group("/user")
	user.Use(authMiddleware())
	{
		user.GET("/profile", getUserProfile)
		user.PUT("/profile", updateUserProfile)
		user.POST("/change-password", changePassword)
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(authMiddleware())
	admin.Use(requireRole(RoleAdmin))
	{
		admin.GET("/users", listUsers)
		admin.PUT("/users/:id/role", changeUserRole)
	}

	return router
}

func main() {
	// Initialize with a default admin user
	adminHash, _ := hashPassword("admin123")
	users = append(users, User{
		ID:            nextUserID,
		Username:      "admin",
		Email:         "admin@example.com",
		PasswordHash:  adminHash,
		FirstName:     "Admin",
		LastName:      "User",
		Role:          RoleAdmin,
		IsActive:      true,
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
	nextUserID++

	router := setupRouter()
	router.Run(":8080")
}