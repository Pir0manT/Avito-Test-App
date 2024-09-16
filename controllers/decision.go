package controllers

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"myapp/models"
)

type DecisionController struct {
	DB *gorm.DB
}

func (ctrl DecisionController) SubmitDecision(c *gin.Context) {
	var bid models.Bid
	var tender models.Tender
	var employee models.Employee
	var orgResp models.OrganizationResponsible
	var decision models.Decision

	bidID := c.Param("bidID")
	username := c.Query("username")
	decisionType := c.Query("decision")

	// Проверка обязательных параметров
	if username == "" || decisionType == "" {
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

	// Проверка статуса тендера
	if tender.Status != models.Published {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Tender is not published"})
		return
	}

	// Проверка авторизации
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized"})
		return
	}

	// Проверка допустимых значений для поля decision
	validDecisions := map[string]bool{
		string(models.Approved): true,
		string(models.Rejected): true,
	}

	if !validDecisions[decisionType] {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid decision value"})
		return
	}

	// Создание решения
	decision = models.Decision{
		BidID:        bid.ID,
		AuthorID:     employee.ID,
		DecisionType: models.DecisionType(decisionType),
	}

	if err := ctrl.DB.Create(&decision).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to submit decision"})
		return
	}

	// Обновление статуса предложения, если решение отклонено
	if decision.DecisionType == models.Rejected {
		bid.Status = models.BidCanceled
		if err := ctrl.DB.Model(&bid).Update("status", models.BidCanceled).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update bid status"})
			return
		}
	}

	// Проверка кворума для одобренных решений
	var approvedCount int64
	ctrl.DB.Model(&models.Decision{}).Where("bid_id = ? AND decision_type = ?", bid.ID, models.Approved).Count(&approvedCount)

	var responsibleCount int64
	ctrl.DB.Model(&models.OrganizationResponsible{}).Where("organization_id = ?", tender.OrganizationID).Count(&responsibleCount)

	quorum := int64(math.Min(3, float64(responsibleCount)))

	if approvedCount >= quorum {
		tender.Status = models.Closed
		if err := ctrl.DB.Model(&tender).Update("status", models.Closed).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update tender status"})
			return
		}
	}

	// Повторная загрузка предложения для получения актуальной версии после срабатывания триггера
	if err := ctrl.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to reload bid"})
		return
	}

	c.JSON(http.StatusOK, bid)
}
