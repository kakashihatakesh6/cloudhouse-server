package handlers

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
)

// Global ClickHouse connection
var conn clickhouse.Conn

// Function to open ClickHouse DB connection
// func initClickHouseConnection() {
// 	var err error
// 	conn := clickhouse.OpenDB(&clickhouse.Options{
// 		Addr:     []string{"kow2h8xlj8.ap-south-1.aws.clickhouse.cloud:9440"}, // 9440 is a secure native TCP port
// 		Protocol: clickhouse.Native,
// 		TLS:      &tls.Config{}, // enable secure TLS
// 		Auth: clickhouse.Auth{
// 			Username: "default",
// 			Password: "cXtWg9Z_Ccowu",
// 		},
// 	})
// 	// Verify connection
// 	if err = conn.Ping(); err != nil {
// 		log.Fatalf("Failed to connect to ClickHouse: %v", err)
// 	}
// 	log.Println("Connected to ClickHouse!")
// }

// Fetch data from the 'covid_data' table
func FetchDataFromClickHouse(c *gin.Context) {

	// Initialize the connection to ClickHouse
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr:     []string{"kow2h8xlj8.ap-south-1.aws.clickhouse.cloud:9440"}, // 9440 is a secure native TCP port
		Protocol: clickhouse.Native,
		TLS:      &tls.Config{}, // enable secure TLS
		Auth: clickhouse.Auth{
			Username: "default",
			Password: "cXtWg9Z_Ccowu",
		},
	})
	if err2 := conn.Ping(); err2 != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err2)
	}

	log.Println("Connected to ClickHouse!")

	// Query the data from ClickHouse table
	query := `SELECT id, time, country, metric, value FROM chart_data.mock_data`
	rows, err := conn.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error executing SELECT query: %s", err),
		})
		return
	}
	defer rows.Close()

	// Prepare response
	var result []map[string]interface{}

	for rows.Next() {
		var id uint64
		var time string
		var country string
		var metric string
		var value float64
		if err := rows.Scan(&id, &time, &country, &metric, &value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error scanning row: %s", err),
			})
			return
		}
		result = append(result, map[string]interface{}{
			"id":      id,
			"time": time,
			"country": country,
			"metric": metric,
			"value": value,
		})
	}

	println("results: ", result)

	// Handle any iteration error
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error iterating over rows: %s", err),
		})
		return
	}

	// Send the result as JSON
	c.JSON(http.StatusOK, result)
}

func GetFilteredCovidData(c *gin.Context) {
	// Get parameters from the request
	country := c.Param("country")
	metric := c.Param("metric")

	// Initialize the connection to ClickHouse
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr:     []string{"kow2h8xlj8.ap-south-1.aws.clickhouse.cloud:9440"}, // 9440 is a secure native TCP port
		Protocol: clickhouse.Native,
		TLS:      &tls.Config{}, // enable secure TLS
		Auth: clickhouse.Auth{
			Username: "default",
			Password: "cXtWg9Z_Ccowu",
		},
	})

	if err := conn.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to connect to ClickHouse: %v", err),
		})
		return
	}

	// Query with parameters
	query := `
		SELECT id, time, country, metric, value 
		FROM chart_data.mock_data 
		WHERE country = ? AND metric = ?
		ORDER BY time`
	
	rows, err := conn.Query(query, country, metric)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error executing SELECT query: %s", err),
		})
		return
	}
	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var id uint64
		var time string
		var country string
		var metric string
		var value float64
		
		if err := rows.Scan(&id, &time, &country, &metric, &value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error scanning row: %s", err),
			})
			return
		}
		
		result = append(result, map[string]interface{}{
			"id":      id,
			"time":    time,
			"country": country,
			"metric":  metric,
			"value":   value,
		})

		fmt.Sprintf("new data %s: ", result)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error iterating over rows: %s", err),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
