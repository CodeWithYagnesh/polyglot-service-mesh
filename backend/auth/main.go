package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
)

func generateUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	dbURL := getEnv("DATABASE_URL", "172.3.0.100:4000")
	dbName := getEnv("DATABASE_NAME", "yagnesh")
	dbUser := getEnv("DB_USERNAME", "root")
	dbPass := getEnv("DB_PASSWORD", "")
	port := getEnv("PORT", "8080")
	fmt.Println(dbUser)
	db := mysql.New("tcp", "", dbURL, dbUser, dbPass, dbName)
	if err := db.Connect(); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.POST("/user/register", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		// Check if email exists
		query := fmt.Sprintf("SELECT id FROM users WHERE email='%s';", req.Email)
		rows, _, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if len(rows) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
			return
		}
		uid := generateUID()[:10]
		insertQuery := fmt.Sprintf("INSERT INTO users(id, username, email, password) VALUES ('%s', '%s', '%s', '%s')", uid, req.Username, req.Email, req.Password)
		_, _, err = db.Query(insertQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"userId":  uid,
			"message": "User registered successfully.",
		})
	})

	r.POST("/user/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		query := fmt.Sprintf("SELECT id FROM users WHERE email='%s' AND password='%s'", req.Email, req.Password)
		rows, _, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if len(rows) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		uid := string(rows[0][0].([]byte))
		c.JSON(http.StatusOK, gin.H{"userId": uid})
	})

	r.GET("/user/get", func(c *gin.Context) {
		userId := c.Query("userId")
		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
			return
		}
		query := fmt.Sprintf("SELECT id, email, username FROM users WHERE id='%s'", userId)
		rows, _, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if len(rows) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		row := rows[0]
		c.JSON(http.StatusOK, gin.H{
			"userId":   string(row[0].([]byte)),
			"email":    string(row[1].([]byte)),
			"username": string(row[2].([]byte)),
		})
	})

	r.PUT("/user/update", func(c *gin.Context) {
		var req struct {
			UserId   string `json:"userId"`
			Email    string `json:"email"`
			Username string `json:"username"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if req.UserId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
			return
		}
		updates := []string{}
		if req.Email != "" {
			updates = append(updates, fmt.Sprintf("email='%s'", req.Email))
		}
		if req.Username != "" {
			updates = append(updates, fmt.Sprintf("username='%s'", req.Username))
		}
		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		updateQuery := fmt.Sprintf("UPDATE users SET %s WHERE id='%s'",
								joinStrings(updates, ", "), req.UserId)
		_, _, err := db.Query(updateQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "User updated successfully.",
		})
	})

	// helper for joining strings

	r.Run(":" + port)
}
func joinStrings(arr []string, sep string) string {
	result := ""
	for i, s := range arr {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
