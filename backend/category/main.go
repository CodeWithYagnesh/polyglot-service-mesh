package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
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
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading: %v", err)
	}

	dbURL := getEnv("DATABASE_URL", "172.3.0.100:4000")
	dbName := getEnv("DATABASE_NAME", "yagnesh")
	dbUser := getEnv("DB_USERNAME", "root")
	dbPass := getEnv("DB_PASSWORD", "")
	port := getEnv("PORT", "8080")
	log.Printf("Using DB user: %s", dbUser)
	db := mysql.New("tcp", "", dbURL, dbUser, dbPass, dbName)
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	r := gin.Default()
	// Category API Endpoints

	r.POST("/category/add", func(c *gin.Context) {
		var req struct {
			UserId      string `json:"userId"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.BindJSON(&req); err != nil || req.UserId == "" || req.Name == "" {
			log.Printf("Invalid add category request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		categoryId := generateUID()[:10]
		insert :=
			"INSERT INTO category(category_id, user_id, name, description) VALUES ('%s', '%s', '%s', '%s')"
		query :=
			sprintf(insert, categoryId, req.UserId, req.Name, req.Description)
		_, _, err := db.Query(query)
		if err != nil {
			log.Printf("Insert category failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "insert category failed"})
			return
		}
		log.Printf("Category added: %s by user %s", categoryId, req.UserId)
		c.JSON(http.StatusOK, gin.H{"categoryId": categoryId})
	})

	r.GET("/category/list", func(c *gin.Context) {
		userId := c.Query("userId")
		if userId == "" {
			log.Print("userId required for category list")
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
			return
		}
		query :=
			"SELECT category_id, name, description FROM category WHERE user_id='%s'"
		sql := sprintf(query, userId)
		rows, _, err := db.Query(sql)
		if err != nil {
			log.Printf("DB error on category list: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		var result []gin.H
		for _, row := range rows {
			result = append(result, gin.H{
				"categoryId":  string(row[0].([]byte)),
				"name":        string(row[1].([]byte)),
				"description": string(row[2].([]byte)),
			})
		}
		log.Printf("Listed category for user %s", userId)
		c.JSON(http.StatusOK, result)
	})

	r.GET("/category/:categoryId", func(c *gin.Context) {
		categoryId := c.Param("categoryId")
		userId := c.Query("userId")
		if userId == "" {
			log.Print("userId required for get category")
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
			return
		}
		query :=
			"SELECT category_id, name, description FROM category WHERE category_id='%s' AND user_id='%s'"
		sql := sprintf(query, categoryId, userId)
		rows, _, err := db.Query(sql)
		if err != nil {
			log.Printf("DB error on get category: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if len(rows) == 0 {
			log.Printf("Category not found: %s for user %s", categoryId, userId)
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		row := rows[0]
		log.Printf("Fetched category %s for user %s", categoryId, userId)
		c.JSON(http.StatusOK, gin.H{
			"categoryId":  string(row[0].([]byte)),
			"name":        string(row[1].([]byte)),
			"description": string(row[2].([]byte)),
		})
	})

	r.PUT("/category/:categoryId", func(c *gin.Context) {
		categoryId := c.Param("categoryId")
		var req struct {
			UserId      string `json:"userId"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.BindJSON(&req); err != nil || req.UserId == "" {
			log.Printf("Invalid update request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		updates := []string{}
		if req.Name != "" {
			updates = append(updates, sprintf("name='%s'", req.Name))
		}
		if req.Description != "" {
			updates = append(updates, sprintf("description='%s'", req.Description))
		}
		if len(updates) == 0 {
			log.Print("No fields to update in category update")
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		update :=
			"UPDATE category SET %s WHERE category_id='%s' AND user_id='%s'"
		sql := sprintf(update, joinStrings(updates, ", "), categoryId, req.UserId)
		_, _, err := db.Query(sql)
		if err != nil {
			log.Printf("Update failed for category %s: %v", categoryId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
		log.Printf("Updated category %s for user %s", categoryId, req.UserId)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	r.DELETE("/category/:categoryId", func(c *gin.Context) {
		categoryId := c.Param("categoryId")
		userId := c.Query("userId")
		if userId == "" {
			log.Print("userId required for delete category")
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
			return
		}
		del :=
			"DELETE FROM category WHERE category_id='%s' AND user_id='%s'"
		sql := sprintf(del, categoryId, userId)
		_, _, err := db.Query(sql)
		if err != nil {
			log.Printf("Delete failed for category %s: %v", categoryId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
			return
		}
		log.Printf("Deleted category %s for user %s", categoryId, userId)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

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

func sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
