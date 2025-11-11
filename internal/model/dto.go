package model

import "time"

// DTOs (Data Transfer Objects) 用于API的请求和响应，实现API契约与数据库模型的解耦

// === Admin DTOs ===

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// === Activity DTOs ===

// CreateActivityRequest 创建活动请求
type CreateActivityRequest struct {
	Title                string    `json:"title" binding:"required"`
	Type                 string    `json:"type" binding:"required"`
	Description          string    `json:"description"`
	StartTime            time.Time `json:"start_time" binding:"required"`
	Location             string    `json:"location" binding:"required"`
	RegistrationDeadline time.Time `json:"registration_deadline" binding:"required"`
	MaxParticipants      int       `json:"max_participants" binding:"gte=0"` // 必须大于等于0
	LiveURL              string    `json:"live_url"`
	AttachmentURL        string    `json:"attachment_url"`
}

// UpdateActivityRequest 更新活动请求
// 使用指针类型允许部分更新 (Partial Update)
type UpdateActivityRequest struct {
	Title                *string    `json:"title"`
	Type                 *string    `json:"type"`
	Description          *string    `json:"description"`
	StartTime            *time.Time `json:"start_time"`
	Location             *string    `json:"location"`
	RegistrationDeadline *time.Time `json:"registration_deadline"`
	MaxParticipants      *int       `json:"max_participants" binding:"omitempty,gte=0"`
	LiveURL              *string    `json:"live_url"`
	AttachmentURL        *string    `json:"attachment_url"`
	Status               *string    `json:"status"` // 用于手动更新状态
}

// ActivityResponse 活动的通用响应
type ActivityResponse struct {
	ID                   uint           `json:"id"`
	AdminID              uint           `json:"admin_id"`
	Title                string         `json:"title"`
	Type                 string         `json:"type"`
	Description          string         `json:"description"`
	StartTime            time.Time      `json:"start_time"`
	Location             string         `json:"location"`
	RegistrationDeadline time.Time      `json:"registration_deadline"`
	MaxParticipants      int            `json:"max_participants"`
	RegisteredCount      int            `json:"registered_count"`
	Status               ActivityStatus `json:"status"`
	LiveURL              string         `json:"live_url,omitempty"`
	AttachmentURL        string         `json:"attachment_url,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
}

// ListActivitiesParams 列表查询参数
type ListActivitiesParams struct {
	Page     int            `form:"page,default=1"`       // 页码
	PageSize int            `form:"page_size,default=10"` // 每页大小
	Type     string         `form:"type"`                 // 按类型过滤
	Status   ActivityStatus `form:"status"`               // 按状态过滤
	DateFrom time.Time      `form:"date_from"`            // 按时间范围过滤
	DateTo   time.Time      `form:"date_to"`
}

// === Registration DTOs ===

// CreateRegistrationRequest 参与者报名请求
type CreateRegistrationRequest struct {
	ParticipantName    string `json:"participant_name" binding:"required"`
	ParticipantPhone   string `json:"participant_phone" binding:"required"`
	ParticipantCollege string `json:"participant_college" binding:"required"`
}

// RegistrationResponse 报名的通用响应
type RegistrationResponse struct {
	ID                 uint      `json:"id"`
	ActivityID         uint      `json:"activity_id"`
	ParticipantName    string    `json:"participant_name"`
	ParticipantPhone   string    `json:"participant_phone"`
	ParticipantCollege string    `json:"participant_college"`
	RegisteredAt       time.Time `json:"registered_at"`
	IsSignedIn         bool      `json:"is_signed_in"`
}
