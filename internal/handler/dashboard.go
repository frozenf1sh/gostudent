package handler

import (
	"time"

	"github.com/frozenf1sh/gostudent/internal/model"
	"github.com/frozenf1sh/gostudent/internal/repository"
	"github.com/frozenf1sh/gostudent/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler interface {
	GetDashboardData(c *gin.Context)
}

type dashboardHandlerImpl struct {
	db               *gorm.DB
	activityRepo     repository.ActivityRepository
	registrationRepo repository.RegistrationRepository
}

func NewDashboardHandler(db *gorm.DB, aRepo repository.ActivityRepository, rRepo repository.RegistrationRepository) DashboardHandler {
	return &dashboardHandlerImpl{
		db:               db,
		activityRepo:     aRepo,
		registrationRepo: rRepo,
	}
}

func (h *dashboardHandlerImpl) GetDashboardData(c *gin.Context) {
	ctx := c.Request.Context()

	var totalActivities int64
	h.db.WithContext(ctx).Model(&model.Activity{}).Count(&totalActivities)

	var publishedActivities int64
	h.db.WithContext(ctx).Model(&model.Activity{}).
		Where("status = ?", model.ActivityStatusPublished).Count(&publishedActivities)

	var totalRegistrations int64
	h.db.WithContext(ctx).Model(&model.Registration{}).Count(&totalRegistrations)

	var todayRegistrations int64
	today := time.Now().Format("2006-01-02")
	h.db.WithContext(ctx).Model(&model.Registration{}).
		Where("DATE(registered_at) = ?", today).Count(&todayRegistrations)

	resp := model.DashboardResponse{
		TotalActivities:     totalActivities,
		PublishedActivities: publishedActivities,
		TotalRegistrations:  totalRegistrations,
		TodayRegistrations:  todayRegistrations,
	}
	utils.Success(c, resp)
}
