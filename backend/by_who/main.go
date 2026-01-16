package main
// Example SQL to create a category table:
//
// CREATE TABLE category (
//     category_id VARCHAR(32) PRIMARY KEY,
//     user_id VARCHAR(32) NOT NULL,
//     name VARCHAR(255) NOT NULL,
//     description TEXT,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
// );

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
	r.POST("/bywho/add", func(c *gin.Context) {
		var req struct {
			UserId      string `json:"userId"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		byWhoId := generateUID()[:10]
		query := fmt.Sprintf(
			"INSERT INTO by_who(by_who_id, user_id, name, description) VALUES ('%s', '%s', '%s', '%s')",
			byWhoId, req.UserId, req.Name, req.Description,
		)
		_, _, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "insert by_who failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"byWhoId": byWhoId})
	})

	r.GET("/bywho/list", func(c *gin.Context) {
		userId := c.Query("userId")
		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
			return
		}
		query := fmt.Sprintf("SELECT by_who_id, name, description FROM by_who WHERE user_id='%s'", userId)
		rows, _, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		var result []gin.H
		for _, row := range rows {
			result = append(result, gin.H{
				"byWhoId":     string(row[0].([]byte)),
				"name":        string(row[1].([]byte)),
				"description": string(row[2].([]byte)),
			})
		}
		c.JSON(http.StatusOK, result)
	})

	r.GET("/bywho/:byWhoId", func(c *gin.Context) {
		byWhoId := c.Param("byWhoId")
		userId := c.Query("userId")
		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
			return
		}
		query := fmt.Sprintf(
			"SELECT by_who_id, name, description FROM by_who WHERE by_who_id='%s' AND user_id='%s'",
			byWhoId, userId,
		)
		rows, _, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if len(rows) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "by_who not found"})
			return
		}
		row := rows[0]
		c.JSON(http.StatusOK, gin.H{
			"byWhoId":     string(row[0].([]byte)),
			"name":        string(row[1].([]byte)),
			"description": string(row[2].([]byte)),
		})
	})

	r.PUT("/bywho/:byWhoId", func(c *gin.Context) {
		byWhoId := c.Param("byWhoId")
		userId := c.Query("userId")
		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
			return
		}
		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		set := []string{}
		if req.Name != nil {
			set = append(set, fmt.Sprintf("name='%s'", *req.Name))
		}
		if req.Description != nil {
			set = append(set, fmt.Sprintf("description='%s'", *req.Description))
		}
		if len(set) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		query := fmt.Sprintf(
			"UPDATE by_who SET %s WHERE by_who_id='%s' AND user_id='%s'",
			joinStrings(set, ", "), byWhoId, userId,
		)
		res, row, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
		// Check if any row was actually updated
		if res == nil || row.AffectedRows() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "by_who not found or not owned by user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	r.DELETE("/bywho/:byWhoId", func(c *gin.Context) {
		byWhoId := c.Param("byWhoId")
		userId := c.Query("userId")
		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
			return
		}
		_, _, err := db.Query(fmt.Sprintf("DELETE FROM by_who WHERE by_who_id='%s' AND user_id='%s'", byWhoId, userId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
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
