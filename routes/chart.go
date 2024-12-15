package routes

import (
	"backend/handlers"
	"github.com/gin-gonic/gin"
)

func SetupChartRoutes(r *gin.RouterGroup) {
	chart := r.Group("/chart")

	{
		chart.GET("/", handlers.GetCharts)
	}
}