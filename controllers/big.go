package controllers

import (
	"myapp/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BidController struct {
	DB *gorm.DB
}

type CreateBidRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description" binding:"required"`
	TenderID    uuid.UUID         `json:"tenderId" binding:"required"`
	AuthorType  models.AuthorType `json:"authorType" binding:"required"`
	AuthorID    uuid.UUID         `json:"authorId" binding:"required"`
}

type UpdateBidRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

func (ctrl BidController) CreateBid(c *gin.Context) {
	var req CreateBidRequest
	var tender models.Tender
	var employee models.Employee
	var orgResp models.OrganizationResponsible

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid request body"})
		return
	}

	// Проверка существования пользователя
	if err := ctrl.DB.Where("id = ?", req.AuthorID).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "User does not exist"})
		return
	}

	// Проверка существования тендера
	if err := ctrl.DB.Where("id = ?", req.TenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка статуса тендера
	if tender.Status != models.Published {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Tender is not published"})
		return
	}

	// Проверка авторизации
	if req.AuthorType == models.AuthorOrganization {
		// Проверка, что пользователь является ответственным лицом какой-либо организации
		if err := ctrl.DB.Where("user_id = ?", req.AuthorID).First(&orgResp).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for any organization"})
			return
		}
	}

	bid := models.Bid{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Status:      models.BidCreated,
		TenderID:    req.TenderID,
		AuthorType:  req.AuthorType,
		AuthorID:    req.AuthorID,
		Version:     1,
		CreatedAt:   time.Now(),
	}

	if err := ctrl.DB.Create(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to create bid"})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func (ctrl BidController) GetUserBids(c *gin.Context) {
	var bids []models.Bid
	var employee models.Employee
	var err error

	// Получение параметров запроса
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")
	username := c.Query("username")

	// Проверка обязательных параметров
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Missing required parameter(s)"})
		return
	}

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

	// Проверка существования пользователя
	if err := ctrl.DB.Where("username = ?", username).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "User does not exist"})
		return
	}

	query := ctrl.DB.Limit(limit).Offset(offset).Order("name").Where("author_id = ?", employee.ID)

	if err = query.Find(&bids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to retrieve bids"})
		return
	}

	c.JSON(http.StatusOK, bids)
}

func (ctrl BidController) GetTenderBids(c *gin.Context) {
	var bids []models.Bid
	var tender models.Tender
	var employee models.Employee
	var orgResp models.OrganizationResponsible
	var err error

	tenderID := c.Param("id")
	username := c.Query("username")

	// Получение параметров запроса
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	// Проверка обязательных параметров
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Missing required parameter(s)"})
		return
	}

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

	// Проверка существования пользователя
	if err := ctrl.DB.Where("username = ?", username).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "User does not exist"})
		return
	}

	// Проверка существования тендера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка, что пользователь является ответственным лицом организации, которая разместила тендер
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this organization"})
		return
	}

	query := ctrl.DB.Limit(limit).Offset(offset).Order("name").Where("tender_id = ? AND status = ?", tenderID, models.Published)

	if err = query.Find(&bids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to retrieve bids"})
		return
	}

	c.JSON(http.StatusOK, bids)
}

func (ctrl BidController) GetBidStatus(c *gin.Context) {
	var bid models.Bid
	var employee models.Employee
	var orgResp models.OrganizationResponsible

	bidID := c.Param("id")
	username := c.Query("username")

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

	// Проверка авторизации
	if bid.AuthorType == models.AuthorOrganization {
		if bid.AuthorID == employee.ID {
			// Пользователь является автором предложения
			c.JSON(http.StatusOK, bid.Status)
			return
		} else {
			// Проверка, что пользователь является ответственным лицом в той же организации
			if err := ctrl.DB.Where("user_id = ? AND organization_id = (SELECT organization_id FROM organization_responsibles WHERE user_id = ?)", employee.ID, bid.AuthorID).First(&orgResp).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized"})
				return
			}
		}
	} else {
		// Проверка, что пользователь является автором предложения
		if bid.AuthorID != employee.ID {
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this bid"})
			return
		}
	}

	c.JSON(http.StatusOK, bid.Status)
}

func (ctrl BidController) UpdateBidStatus(c *gin.Context) {
	var bid models.Bid
	var employee models.Employee
	var orgResp models.OrganizationResponsible

	bidID := c.Param("bidID")
	username := c.Query("username")
	statusStr := c.Query("status")

	// Проверка обязательных параметров
	if username == "" || statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Missing required parameter(s)"})
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

	// Проверка авторизации
	if bid.AuthorType == models.AuthorOrganization {
		if bid.AuthorID == employee.ID {
			// Пользователь является автором предложения
		} else {
			// Проверка, что пользователь является ответственным лицом в той же организации
			if err := ctrl.DB.Where("user_id = ? AND organization_id = (SELECT organization_id FROM organization_responsibles WHERE user_id = ?)", employee.ID, bid.AuthorID).First(&orgResp).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized"})
				return
			}
		}
	} else {
		// Проверка, что пользователь является автором предложения
		if bid.AuthorID != employee.ID {
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this bid"})
			return
		}
	}

	// Проверка корректности статуса

	var newStatus models.BidStatus
	switch statusStr {
	case string(models.Created):
		newStatus = models.BidCreated
	case string(models.Published):
		newStatus = models.BidPublished
	case string(models.BidCanceled):
		newStatus = models.BidCanceled
	default:
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid status parameter"})
		return
	}

	// Обновление статуса предложения
	bid.Status = newStatus

	if err := ctrl.DB.Model(&bid).Update("status", newStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update bid status"})
		return
	}

	// Повторная загрузка предложения для получения актуальной версии после срабатывания триггера
	if err := ctrl.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to reload bid"})
		return
	}
	c.JSON(http.StatusOK, bid)
}

func (ctrl BidController) EditBid(c *gin.Context) {
	var bid models.Bid
	var employee models.Employee
	var orgResp models.OrganizationResponsible
	var req UpdateBidRequest

	bidID := c.Param("bidID")
	username := c.Query("username")

	// Проверка обязательных параметров
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Missing required parameter(s)"})
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

	// Проверка авторизации
	if bid.AuthorType == models.AuthorOrganization {
		if bid.AuthorID == employee.ID {
			// Пользователь является автором предложения
		} else {
			// Проверка, что пользователь является ответственным лицом в той же организации
			if err := ctrl.DB.Where("user_id = ? AND organization_id = (SELECT organization_id FROM organization_responsibles WHERE user_id = ?)", employee.ID, bid.AuthorID).First(&orgResp).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized"})
				return
			}
		}
	} else {
		// Проверка, что пользователь является автором предложения
		if bid.AuthorID != employee.ID {
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this bid"})
			return
		}
	}

	// Обновление полей предложения
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid request body"})
		return
	}

	// Проверка допустимых значений для поля status
	validStatuses := map[string]bool{
		"Created":   true,
		"Published": true,
		"Closed":    true,
	}

	if req.Status != "" && !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid status value"})
		return
	}

	if err := ctrl.DB.Model(&bid).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update bid"})
		return
	}

	// Повторная загрузка предложения для получения актуальной версии после срабатывания триггера
	if err := ctrl.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to reload bid"})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func (ctrl BidController) RollbackBid(c *gin.Context) {
	var bid models.Bid
	var bidHistory models.BidHistory
	var employee models.Employee
	var orgResp models.OrganizationResponsible

	bidID := c.Param("bidID")
	versionStr := c.Param("version")
	username := c.Query("username")

	// Проверка обязательных параметров
	if username == "" || versionStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Missing required parameter(s)"})
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

	// Проверка авторизации
	if bid.AuthorType == models.AuthorOrganization {
		if bid.AuthorID == employee.ID {
			// Пользователь является автором предложения
		} else {
			// Проверка, что пользователь является ответственным лицом в той же организации
			if err := ctrl.DB.Where("user_id = ? AND organization_id = (SELECT organization_id FROM organization_responsibles WHERE user_id = ?)", employee.ID, bid.AuthorID).First(&orgResp).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized"})
				return
			}
		}
	} else {
		// Проверка, что пользователь является автором предложения
		if bid.AuthorID != employee.ID {
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this bid"})
			return
		}
	}

	// Преобразование версии в int
	version, err := strconv.Atoi(versionStr)
	if err != nil || version <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid version parameter"})
		return
	}

	// Поиск истории предложения по версии
	if err := ctrl.DB.Where("bid_id = ? AND version = ?", bidID, version).First(&bidHistory).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid version not found"})
		return
	}

	// Откат предложения к указанной версии
	bid.Name = bidHistory.Name
	bid.Description = bidHistory.Description
	bid.Status = bidHistory.Status

	if err := ctrl.DB.Model(&bid).Updates(bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to rollback bid"})
		return
	}

	// Повторная загрузка предложения для получения актуальной версии после срабатывания триггера
	if err := ctrl.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to reload bid"})
		return
	}

	c.JSON(http.StatusOK, bid)
}
