package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"myapp/models"
)

type ReviewController struct {
	DB *gorm.DB
}

type ReviewResponse struct {
	ID        string `json:"id"`
	Feedback  string `json:"description"`
	CreatedAt string `json:"createdAt"`
}

func (ctrl ReviewController) SubmitFeedback(c *gin.Context) {
	var bid models.Bid
	var tender models.Tender
	var employee models.Employee
	var orgResp models.OrganizationResponsible
	var review models.Review

	bidID := c.Param("bidID")
	username := c.Query("username")
	bidFeedback := c.Query("bidFeedback")

	// Проверка обязательных параметров
	if username == "" || bidFeedback == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Missing required parameters"})
		return
	}

	// Проверка существования пользователя
	if err := ctrl.DB.Where("username = ?", username).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "User does not exist"})
		return
	}

	// Проверка существования предложения
	if err := ctrl.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	// Проверка статуса предложения
	if bid.Status != models.BidPublished {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid is not published"})
		return
	}

	// Проверка существования тендера
	if err := ctrl.DB.Where("id = ?", bid.TenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка авторизации
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized"})
		return
	}

	// Создание отзыва
	review = models.Review{
		BidID:       bid.ID,
		BidAuthorID: employee.ID,
		Description: bidFeedback,
	}

	if err := ctrl.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func (ctrl ReviewController) GetReviews(c *gin.Context) {
	var tender models.Tender
	var employee models.Employee
	var orgResp models.OrganizationResponsible
	var reviews []models.Review
	var reviewResponses []ReviewResponse

	tenderID := c.Param("id")
	authorUsername := c.Query("authorUsername")
	requesterUsername := c.Query("requesterUsername")
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	// Проверка обязательных параметров
	if authorUsername == "" || requesterUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Missing required parameters"})
		return
	}

	// Преобразование limit и offset в int
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid offset parameter"})
		return
	}

	// Проверка существования пользователя, запрашивающего отзывы
	if err := ctrl.DB.Where("username = ?", requesterUsername).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Requester does not exist"})
		return
	}

	// Проверка существования тендера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка авторизации
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized"})
		return
	}

	// Проверка существования автора предложений
	var author models.Employee
	if err := ctrl.DB.Where("username = ?", authorUsername).First(&author).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Author not found"})
		return
	}

	// Получение отзывов на предложения автора
	if err := ctrl.DB.Where("bid_author_id = ?", author.ID).Limit(limit).Offset(offset).Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to retrieve reviews"})
		return
	}

	// Формирование ответа
	for _, review := range reviews {
		reviewResponses = append(reviewResponses, ReviewResponse{
			ID:        review.ID.String(),
			Feedback:  review.Description,
			CreatedAt: review.CreatedAt.String(),
		})
	}

	c.JSON(http.StatusOK, reviewResponses)
}
