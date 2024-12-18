package handlers

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
)

var conn clickhouse.Conn

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
	query := `SELECT id, time, country, metric, value FROM chart_data.covid_data`
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
			"time":    time,
			"country": country,
			"metric":  metric,
			"value":   value,
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
	// Get all possible filter parameters
	params := map[string]string{
		"country": c.Param("country"),
		"metric":  c.Param("metric"),
		"time":    c.Param("time"),
		"id":      c.Param("id"),
	}

	// Build dynamic WHERE clause
	whereClause := ""
	var queryParams []interface{}
	first := true

	for key, value := range params {
		if value != "" {
			if first {
				whereClause = "WHERE "
				first = false
			} else {
				whereClause += " AND "
			}
			whereClause += key + " = ?"
			queryParams = append(queryParams, value)
		}
	}

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

	// Query with dynamic parameters
	query := fmt.Sprintf(`
		SELECT id, time, country, metric, value 
		FROM chart_data.covid_data 
		%s
		ORDER BY time`, whereClause)

	rows, err := conn.Query(query, queryParams...)
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

func GetFilteredData(c *gin.Context) {
	// Extract query parameters
	params := map[string]string{
		"startDate": c.Query("startDate"),
		"endDate":   c.Query("endDate"),
		"country":   c.Query("country"),
		"metric":    c.Query("metric"),
	}

	// Build dynamic WHERE clause
	whereClause := ""
	var queryParams []interface{}
	first := true

	for key, value := range params {
		if value != "" {
			if first {
				whereClause = "WHERE "
				first = false
			} else {
				whereClause += " AND "
			}

			switch key {
			case "startDate":
				whereClause += "time >= ?"
			case "endDate":
				whereClause += "time <= ?"
			case "country":
				whereClause += "country = ?"
			case "metric":
				whereClause += "metric = ?"
			}
			queryParams = append(queryParams, value)
		}
	}

	// Initialize the ClickHouse connection
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{"kow2h8xlj8.ap-south-1.aws.clickhouse.cloud:9440"}, // Replace with your ClickHouse address
		Protocol: clickhouse.Native,
		TLS: &tls.Config{}, // Enable secure TLS
		Auth: clickhouse.Auth{
			Username: "default",
			Password: "cXtWg9Z_Ccowu", // Replace with your password
		},
	})

	// Ensure the connection is active
	if err := conn.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to connect to ClickHouse: %v", err),
		})
		return
	}
	defer conn.Close()

	// Query with dynamic WHERE clause
	query := fmt.Sprintf(`
		SELECT id, time, country, metric, value 
		FROM chart_data.covid_data
		%s
		ORDER BY time`, whereClause)

	// Execute the query
	rows, err := conn.Query(query, queryParams...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error executing query: %v", err),
		})
		return
	}
	defer rows.Close()

	// Parse query results
	var result []map[string]interface{}
	for rows.Next() {
		var id uint64
		var time string
		var country string
		var metric string
		var value float64

		if err := rows.Scan(&id, &time, &country, &metric, &value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error scanning row: %v", err),
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
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error iterating over rows: %v", err),
		})
		return
	}

	// Return the filtered data
	c.JSON(http.StatusOK, result)
}
