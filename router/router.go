package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"myapp/controllers"
	"myapp/handlers"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()
	reviewController := controllers.ReviewController{DB: db}
	decisionController := controllers.DecisionController{DB: db}
	tenderController := controllers.TenderController{DB: db}
	bidController := controllers.BidController{DB: db}

	// Маршрут для проверки доступности сервера
	router.GET("/api/ping", handlers.PingHandler)

	// Маршруты для тендеров
	router.GET("/api/tenders", tenderController.GetTenders)
	router.POST("/api/tenders/new", tenderController.CreateTender)
	router.GET("/api/tenders/my", tenderController.GetUserTenders)
	router.GET("/api/tenders/:tenderId/status", tenderController.GetTenderStatus)
	router.PUT("/api/tenders/:tenderId/status", tenderController.UpdateTenderStatus)
	router.PATCH("/api/tenders/:tenderId/edit", tenderController.EditTender)
	router.PUT("/api/tenders/:tenderId/rollback/:version", tenderController.RollbackTender)

	// Маршруты для предложений
	router.POST("/api/bids/new", bidController.CreateBid)
	router.GET("/api/bids/my", bidController.GetUserBids)
	router.GET("/api/bids/:id/*action", func(c *gin.Context) {
		action := c.Param("action")

		if action == "/list" {
			bidController.GetTenderBids(c)
		} else if action == "/status" {
			bidController.GetBidStatus(c)
		} else if action == "/reviews" {
			reviewController.GetReviews(c)
		} else {
			c.JSON(400, gin.H{"error": "Invalid action"})
		}
	})
	router.PUT("/api/bids/:bidID/status", bidController.UpdateBidStatus)
	router.PATCH("/api/bids/:bidID/edit", bidController.EditBid)
	router.PUT("/api/bids/:bidID/rollback/:version", bidController.RollbackBid)

	// Маршруты для решений по предложениям
	router.PUT("/api/bids/:bidID/submit_decision", decisionController.SubmitDecision)

	// Маршруты для отзывов
	router.PUT("/api/bids/:bidID/feedback", reviewController.SubmitFeedback)

	return router
}
