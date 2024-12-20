package routes

import (
	"backend/handlers"

	"github.com/gin-gonic/gin"
)

func SetupCovidRoutes(r *gin.RouterGroup) {
	covid := r.Group("/covid")

	{
		covid.GET("/getdata", handlers.FetchDataFromClickHouse)
		// covid.GET("/getfiltered/*params", handlers.GetFilteredCovidData)
		covid.GET("/getfiltered", handlers.GetFilteredData)
	}
}