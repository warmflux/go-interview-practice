package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// User represents a user in our system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// In-memory storage
var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 25},
	{ID: 3, Name: "Bob Wilson", Email: "bob@example.com", Age: 35},
}
var nextID = 4

func main() {
	// TODO: Create Gin router
	r := gin.Default()

	// TODO: Setup routes
	// GET /users - Get all users
	// GET /users/:id - Get user by ID
	// POST /users - Create new user
	// PUT /users/:id - Update user
	// DELETE /users/:id - Delete user
	// GET /users/search - Search users by name
	r.GET("/user", getAllUsers)
	r.GET("/user/:id", getUserByID)
	r.POST("/users", createUser)
	r.PUT("/user/:id", updateUser)
	r.DELETE("/users/:id", deleteUser)

	// TODO: Start server on port 8080
	r.Run(":8081")
}

// TODO: Implement handler functions

// getAllUsers handles GET /users
func getAllUsers(c *gin.Context) {
	// TODO: Return all users
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "",
		Data:    users,
		Code:    200,
	})
}

// getUserByID handles GET /users/:id
func getUserByID(c *gin.Context) {
	// TODO: Get user by ID
	// Handle invalid ID format
	// Return 404 if user not found
	id := c.Param("id")
	var userBeGet User
	found := false
	userId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    400,
			Error:   "",
		})
		return
	}
	for _, i := range users {
		if i.ID == userId {
			userBeGet = i
			found = true
		}
	}
	if found {
		c.JSON(200, Response{
			Success: true,
			Message: "",
			Data:    userBeGet,
			Code:    200,
		})
	} else {
		c.JSON(404, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    404,
			Error:   "",
		})
	}
}

// createUser handles POST /users
func createUser(c *gin.Context) {
	// TODO: Parse JSON request body
	// Validate required fields
	// Add user to storage
	// Return created user
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    400,
			Error:   "",
		})
		return
	}
	user.ID = nextID

	if user.Name == "" || user.Email == "" {
		c.JSON(400, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    400,
			Error:   "",
		})
		return
	}
	for _, u := range users {
		if user.ID == u.ID {
			c.JSON(400, Response{
				Success: false,
				Message: "",
				Data:    nil,
				Code:    400,
				Error:   "",
			})
			return
		}
	}
	users = append(users, user)
	c.JSON(201, Response{
		Success: true,
		Message: "",
		Data:    user,
		Code:    201,
	})
}

// updateUser handles PUT /users/:id
func updateUser(c *gin.Context) {
	// TODO: Get user ID from path
	// Parse JSON request body
	// Find and update user
	// Return updated user
	id := c.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(404, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    404,
			Error:   "",
		})
		return
	}
	user, code := findUserByID(userId)
	if code == -1 {
		c.JSON(404, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    404,
			Error:   "",
		})
		return
	}
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(404, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    404,
			Error:   "",
		})
		return
	}
	user.ID = newUser.ID
	user.Age = newUser.Age
	user.Email = newUser.Email
	user.Name = newUser.Name
	c.JSON(200, Response{
		Success: true,
		Message: "",
		Data:    newUser,
		Code:    200,
	})
}

// deleteUser handles DELETE /users/:id
func deleteUser(c *gin.Context) {
	// TODO: Get user ID from path
	// Find and remove user
	// Return success message
	id := c.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(404, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    404,
			Error:   "",
		})
		return
	}
	_, code := findUserByID(userId)
	if code == -1 {
		c.JSON(404, Response{
			Success: false,
			Message: "",
			Data:    nil,
			Code:    404,
			Error:   "",
		})
		return
	}
	var index int
	for i, u := range users {
		if u.ID == userId {
			index = i
		}
	}
	users = append(users[:index], users[index+1:]...)
	c.JSON(200, Response{
		Success: true,
		Message: "",
		Data:    nil,
		Code:    200,
	})
}

// searchUsers handles GET /users/search?name=value
func searchUsers(c *gin.Context) {
	// TODO: Get name query parameter
	// Filter users by name (case-insensitive)
	// Return matching users
	userName := c.Query("name")
	if userName == "" {
		c.JSON(400, Response{
			Success: false,
			Error:   "",
			Code:    400,
			Data:    nil,
		})
		return
	}

	matchedUsers := make([]User, 0)
	for _, u := range users {
		if strings.Contains(strings.ToLower(u.Name), strings.ToLower(userName)) {
			matchedUsers = append(matchedUsers, u)
		}
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    matchedUsers,
		Message: "Users retrieved successfully!",
	})
}

// Helper function to find user by ID
func findUserByID(id int) (*User, int) {
	// TODO: Implement user lookup
	// Return user pointer and index, or nil and -1 if not found

	var user User
	for i, u := range users {
		if u.ID == id {
			user = u
			return &user, i
		}
	}
	return nil, -1
}

// Helper function to validate user data
func validateUser(user User) error {
	// TODO: Implement validation
	// Check required fields: Name, Email
	// Validate email format (basic check)
	if user.Name == "" {
		return errors.New("")
	}
	if user.Email == "" || !strings.Contains(user.Email, "@") {
		return errors.New("")
	}
	return nil
}