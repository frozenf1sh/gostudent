package handler

import (
	"github.com/frozenf1sh/gostudent/internal/service"
)

// Handler 接口定义了所有 Handler 层的聚合
type Handler interface {
	Admin() AdminHandler
	Activity() ActivityHandler
	Registration() RegistrationHandler
}

// handlerImpl 实现了 Handler 接口
type handlerImpl struct {
	adminHandler        AdminHandler
	activityHandler     ActivityHandler
	registrationHandler RegistrationHandler
}

// NewHandler 创建 Handler 实例，并注入 Service 依赖
func NewHandler(svc service.Service) Handler {
	return &handlerImpl{
		adminHandler:        NewAdminHandler(svc.Admin()),
		activityHandler:     NewActivityHandler(svc.Activity()),
		registrationHandler: NewRegistrationHandler(svc.Registration()),
	}
}

// Admin 返回 AdminHandler 实例
func (h *handlerImpl) Admin() AdminHandler {
	return h.adminHandler
}

// Activity 返回 ActivityHandler 实例
func (h *handlerImpl) Activity() ActivityHandler {
	return h.activityHandler
}

// Registration 返回 RegistrationHandler 实例
func (h *handlerImpl) Registration() RegistrationHandler {
	return h.registrationHandler
}
