package main

import (
	"backend/config"
	"backend/database"
	"backend/routes"
	"crypto/tls"
	"fmt"

	// "backend/routes"
	"log"
	"net/http"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {

	envFile := ".env"
	if _, err := os.Stat(".env.local"); err == nil {
		envFile = ".env.local"
	}

	err := godotenv.Load(envFile)

	if err != nil {
		log.Fatalf("Error loading %s file: %s", envFile, err)
	}

	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	config.LoadConfig()

	database.ConnectToPostgres(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "welcome to the Gin server!",
		})
	})

	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr:     []string{"kow2h8xlj8.ap-south-1.aws.clickhouse.cloud:9440"}, // 9440 is a secure native TCP port
		Protocol: clickhouse.Native,
		TLS:      &tls.Config{}, // enable secure TLS
		Auth: clickhouse.Auth{
			Username: "default",
			Password: "cXtWg9Z_Ccowu",
		},
	})

	// Inserting data into a table
	insertQuery := `
		CREATE TABLE IF NOT EXISTS test_table (
			id UInt32,
			name String
		) ENGINE = MergeTree ORDER BY id;
	`

	// Execute the create table query
	if _, err := conn.Exec(insertQuery); err != nil {
		log.Fatalf("An error creating table: %s", err)
	}

	// Insert data into the table
	insertData := `
		INSERT INTO test_table (id, name) VALUES (?, ?)
	`

	rowsAffected, err := conn.Exec(insertData, 1, "John Doe")

	if err != nil {
		log.Fatalf("An error inserting data: %s", err)
	}

	fmt.Printf("Rows inserted: %d\n", rowsAffected)

	// Fetch data from the table
	selectQuery := `SELECT id, country FROM chart_data.covid_data`
	selectQuery2 := `SELECT * FROM chart_data.covid_data`

	rows, err := conn.Query(selectQuery)
	rows2, err2 := conn.Query(selectQuery2)
	if err != nil {
		log.Fatalf("An error executing SELECT query: %s", err)
	}
	if err2 != nil {
		log.Fatalf("An error executing SELECT query: %s", err2)
	}

	defer rows.Close()
	defer rows2.Close()

	// Iterate over the rows and print the results
	fmt.Println("Data from chart_data.covid_data:")
	fmt.Println("Data from chart_data:", rows2)

	for rows.Next() {
		var id uint64
		var country string

		if err := rows.Scan(&id, &country); err != nil {
			log.Fatalf("Error scanning row: %s", err)
		}
		fmt.Printf("ID: %d, Country: %s\n", id, country)
	}

	// Check for any error during row iteration
	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating over rows: %s", err)
	}

	// row := conn.QueryRow("SELECT 1")
	// var col uint8
	// if err := row.Scan(&col); err != nil {
	// 	fmt.Printf("An error while reading the data: %s", err)
	// } else {
	// 	fmt.Printf("Result: %d", col)
	// }

	r.Use(CORSMiddleware())

	api := r.Group("/api")
	routes.SetupChartRoutes(api)
	routes.SetupCovidRoutes(api)

	r.Run(":8000")
}
