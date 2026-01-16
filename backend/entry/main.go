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

	r.POST("/entry/add", func(c *gin.Context) {
		var req struct {
			UserId          string  `json:"userId"`
			TransactionType string  `json:"transaction_type"`
			Owe             bool    `json:"owe"`
			Tid             string  `json:"tid"`
			Date            string  `json:"date"`
			Reason          string  `json:"reason"`
			ByWho           string  `json:"by_who"`
			Category        string  `json:"category"`
			Amount          float64 `json:"amount"`
			OweList         []struct {
				UserId  string  `json:"userid"`
				OweType string  `json:"owe_type"`
				Amount  float64 `json:"amount"`
			} `json:"oweList"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		tid := req.Tid
		if tid == "" {
			tid = generateUID()[:10]
		}
		// Insert entry
		insertEntry := fmt.Sprintf(
			"INSERT INTO entries(tid, user_id, transaction_type, owe, date, reason, by_who, category, amount) VALUES ('%s', '%s', '%s', %t, '%s', '%s', '%s', '%s', %f)",
			tid, req.UserId, req.TransactionType, req.Owe, req.Date, req.Reason, req.ByWho, req.Category, req.Amount,
		)
		_, _, err := db.Query(insertEntry)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "insert entry failed"})
			return
		}
		// Insert oweList
		for _, owe := range req.OweList {
			insertOwe := fmt.Sprintf(
				"INSERT INTO owe_list(tid, userid, owe_type, amount) VALUES ('%s', '%s', '%s', %f)",
				tid, owe.UserId, owe.OweType, owe.Amount,
			)
			_, _, err := db.Query(insertOwe)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "insert oweList failed"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"userId": req.UserId})
	})

	r.GET("/entry/list", func(c *gin.Context) {
		userId := c.Query("userId")
		dateFrom := c.Query("dateFrom")
		dateTo := c.Query("dateTo")
		where := []string{}
		if userId != "" {
			where = append(where, fmt.Sprintf("user_id='%s'", userId))
		}
		if dateFrom != "" {
			where = append(where, fmt.Sprintf("date>='%s'", dateFrom))
		}
		if dateTo != "" {
			where = append(where, fmt.Sprintf("date<='%s'", dateTo))
		}
		query := "SELECT tid, transaction_type, date, reason, by_who, category, amount FROM entries"
		if len(where) > 0 {
			query += " WHERE " + joinStrings(where, " AND ")
		}
		rows, _, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		var result []gin.H
		for _, row := range rows {
			tid := string(row[0].([]byte))
			// Fetch oweList for this tid
			oweRows, _, _ := db.Query(fmt.Sprintf("SELECT userid, owe_type, amount FROM owe_list WHERE tid='%s'", tid))
			oweList := []gin.H{}
			for _, o := range oweRows {
				oweList = append(oweList, gin.H{
					"userid":   string(o[0].([]byte)),
					"owe_type": string(o[1].([]byte)),
					"amount":   string(o[2].([]byte)),
				})
			}
			result = append(result, gin.H{
				"tid":              tid,
				"transaction_type": string(row[1].([]byte)),
				"date":             string(row[2].([]byte)),
				"reason":           string(row[3].([]byte)),
				"by_who":           string(row[4].([]byte)),
				"category":         string(row[5].([]byte)),
				"amount":           string(row[6].([]byte)),
				"oweList":          oweList,
			})
		}
		c.JSON(http.StatusOK, result)
	})

	r.GET("/entry/:tid", func(c *gin.Context) {
		tid := c.Param("tid")
		userId := c.Query("userId")
		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
			return
		}
		query := fmt.Sprintf("SELECT tid, transaction_type, date, reason, by_who, category, amount FROM entries WHERE tid='%s' AND user_id='%s'", tid, userId)
		rows, _, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if len(rows) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
			return
		}
		row := rows[0]
		oweRows, _, _ := db.Query(fmt.Sprintf("SELECT userid, owe_type, amount FROM owe_list WHERE tid='%s'", tid))
		oweList := []gin.H{}
		for _, o := range oweRows {
			oweList = append(oweList, gin.H{
				"userid":   string(o[0].([]byte)),
				"owe_type": string(o[1].([]byte)),
				"amount":   string(o[2].([]byte)),
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"tid":              string(row[0].([]byte)),
			"transaction_type": string(row[1].([]byte)),
			"date":             string(row[2].([]byte)),
			"reason":           string(row[3].([]byte)),
			"by_who":           string(row[4].([]byte)),
			"category":         string(row[5].([]byte)),
			"amount":           string(row[6].([]byte)),
			"oweList":          oweList,
		})
	})

	r.DELETE("/entry/:tid", func(c *gin.Context) {
		tid := c.Param("tid")
		userId := c.Query("userId")
		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
			return
		}
		// Check if entry exists and belongs to user
		rows, _, err := db.Query(fmt.Sprintf("SELECT tid FROM entries WHERE tid='%s' AND user_id='%s'", tid, userId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if len(rows) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "entry not found or not owned by user"})
			return
		}
		_, _, err = db.Query(fmt.Sprintf("DELETE FROM owe_list WHERE tid='%s'", tid))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete owe_list failed"})
			return
		}
		_, _, err = db.Query(fmt.Sprintf("DELETE FROM entries WHERE tid='%s' AND user_id='%s'", tid, userId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete entry failed"})
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
