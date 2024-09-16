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

type TenderController struct {
	DB *gorm.DB
}

type CreateTenderRequest struct {
	Name            string             `json:"name" binding:"required"`
	Description     string             `json:"description" binding:"required"`
	ServiceType     models.ServiceType `json:"serviceType" binding:"required"`
	OrganizationID  uuid.UUID          `json:"organizationId" binding:"required"`
	CreatorUsername string             `json:"creatorUsername" binding:"required"`
}

type UpdateTenderRequest struct {
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	ServiceType models.ServiceType `json:"serviceType,omitempty"`
	Status      models.Status      `json:"status,omitempty"`
}

func (ctrl TenderController) GetTenders(c *gin.Context) {
	var tenders []models.Tender
	var err error

	// Получение параметров запроса
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")
	serviceTypes := c.QueryArray("service_type")

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

	// Поскольку в этой ручке нет параметра username, то отображаются только опубликованные тендеры
	// которые доступны всем пользователям
	query := ctrl.DB.Limit(limit).Offset(offset).Order("name").Where("status = ?", models.Published)

	if len(serviceTypes) > 0 {
		query = query.Where("service_type IN ?", serviceTypes)
	}

	if err = query.Find(&tenders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to retrieve tenders"})
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func (ctrl TenderController) CreateTender(c *gin.Context) {
	var req CreateTenderRequest
	var employee models.Employee
	var orgResp models.OrganizationResponsible
	var organization models.Organization

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid request body"})
		return
	}

	// Проверка существования организации
	if err := ctrl.DB.Where("id = ?", req.OrganizationID).First(&organization).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Organization does not exist"})
		return
	}

	// Проверка существования пользователя
	if err := ctrl.DB.Where("username = ?", req.CreatorUsername).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "User does not exist"})
		return
	}

	// Проверка, что пользователь является ответственным лицом организации
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, req.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this organization"})
		return
	}

	// Проверка корректности ServiceType
	validServiceTypes := map[models.ServiceType]bool{
		models.Construction: true,
		models.Delivery:     true,
		models.Manufacture:  true,
	}

	if !validServiceTypes[req.ServiceType] {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid service type"})
		return
	}

	tender := models.Tender{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		ServiceType:    req.ServiceType,
		Status:         models.Created,
		OrganizationID: req.OrganizationID,
		CreatedAt:      time.Now(),
	}

	if err := ctrl.DB.Create(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to create tender"})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func (ctrl TenderController) GetUserTenders(c *gin.Context) {
	var tenders []models.Tender
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

	query := ctrl.DB.Limit(limit).Offset(offset).Order("name").Where("organization_id IN (?)", ctrl.DB.Table("organization_responsibles").Select("organization_id").Where("user_id = ?", employee.ID))

	if err = query.Find(&tenders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to retrieve tenders"})
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func (ctrl TenderController) GetTenderStatus(c *gin.Context) {
	var tender models.Tender
	var employee models.Employee
	var orgResp models.OrganizationResponsible

	tenderID := c.Param("tenderId")
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

	// Проверка существования тендера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка, что пользователь является ответственным лицом организации
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this organization"})
		return
	}

	c.JSON(http.StatusOK, tender.Status)
}

func (ctrl TenderController) UpdateTenderStatus(c *gin.Context) {
	var tender models.Tender
	var employee models.Employee
	var orgResp models.OrganizationResponsible

	tenderID := c.Param("tenderId")
	statusStr := c.Query("status")
	username := c.Query("username")

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

	// Проверка существования тендера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка, что пользователь является ответственным лицом организации
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this organization"})
		return
	}

	// Проверка корректности статуса
	var newStatus models.Status
	switch statusStr {
	case string(models.Created):
		newStatus = models.Created
	case string(models.Published):
		newStatus = models.Published
	case string(models.Closed):
		newStatus = models.Closed
	default:
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid status parameter"})
		return
	}

	// Обновление статуса тендера
	tender.Status = newStatus

	if err := ctrl.DB.Model(&tender).Update("status", newStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update tender status"})
		return
	}

	// Повторная загрузка тендера для получения актуальной версии тендера после срабатывания триггера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to reload tender"})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func (ctrl TenderController) EditTender(c *gin.Context) {
	var tender models.Tender
	var employee models.Employee
	var orgResp models.OrganizationResponsible
	var req UpdateTenderRequest

	tenderID := c.Param("tenderId")
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

	// Проверка существования тендера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка, что пользователь является ответственным лицом организации
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this organization"})
		return
	}

	// Обновление полей тендера
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid request body"})
		return
	}

	// Обновление полей тендера
	if err := ctrl.DB.Model(&tender).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update tender"})
		return
	}

	// Повторная загрузка тендера для получения актуальной версии тендера после срабатывания триггера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to reload tender"})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func (ctrl TenderController) RollbackTender(c *gin.Context) {
	var tender models.Tender
	var tenderHistory models.TenderHistory
	var employee models.Employee
	var orgResp models.OrganizationResponsible

	tenderID := c.Param("tenderId")
	versionStr := c.Param("version")
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

	// Проверка существования тендера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка, что пользователь является ответственным лицом организации
	if err := ctrl.DB.Where("user_id = ? AND organization_id = ?", employee.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized for this organization"})
		return
	}

	// Преобразование версии в int
	version, err := strconv.Atoi(versionStr)
	if err != nil || version <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid version parameter"})
		return
	}

	// Поиск истории тендера по версии
	if err := ctrl.DB.Where("tender_id = ? AND version = ?", tenderID, version).First(&tenderHistory).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender version not found"})
		return
	}

	// Откат тендера к указанной версии
	tender.Name = tenderHistory.Name
	tender.Description = tenderHistory.Description
	tender.ServiceType = tenderHistory.ServiceType
	tender.Status = tenderHistory.Status

	if err := ctrl.DB.Model(&tender).Updates(tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to rollback tender"})
		return
	}

	// Повторная загрузка тендера для получения актуальной версии тендера после срабатывания триггера
	if err := ctrl.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to reload tender"})
		return
	}

	c.JSON(http.StatusOK, tender)
}
