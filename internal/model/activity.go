package model

import "time"

// Activity 对应 'activities' 表，存储活动信息
type Activity struct {
	ID                   uint           `gorm:"primarykey"`
	AdminID              uint           `gorm:"not null" json:"admin_id"` // 创建活动的管理员ID
	Admin                Admin          `gorm:"foreignKey:AdminID" json:"admin"`
	Title                string         `gorm:"type:varchar(255);not null" json:"title"`                 // 活动名称
	Type                 string         `gorm:"type:varchar(50);not null" json:"type"`                   // 活动类型 (讲座, 宣讲会等)
	Description          string         `gorm:"type:text" json:"description"`                            // 活动简介
	StartTime            time.Time      `gorm:"not null" json:"start_time"`                              // 活动时间
	Location             string         `gorm:"type:varchar(255);not null" json:"location"`              // 活动地点
	RegistrationDeadline time.Time      `gorm:"not null" json:"registration_deadline"`                   // 报名截止时间
	MaxParticipants      int            `gorm:"not null;default:0" json:"max_participants"`              // 人数上限 (0表示不限制)
	RegisteredCount      int            `gorm:"not null;default:0" json:"registered_count"`              // 已报名人数
	Status               ActivityStatus `gorm:"type:varchar(20);not null;default:'DRAFT'" json:"status"` // 活动状态

	// 可选字段
	LiveURL       string `gorm:"type:varchar(512)" json:"live_url"`       // 直播链接
	AttachmentURL string `gorm:"type:varchar(512)" json:"attachment_url"` // 附件链接

	// 关联关系：一个活动有多条报名记录
	Registrations []Registration `gorm:"foreignKey:ActivityID" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 自定义Gorm的表名
func (Activity) TableName() string {
	return "activities"
}

// ActivityStatus 定义了活动的状态
type ActivityStatus string

const (
	// ActivityStatusDraft 草稿状态，未发布
	ActivityStatusDraft ActivityStatus = "DRAFT"
	// ActivityStatusPublished 已发布，报名中
	ActivityStatusPublished ActivityStatus = "PUBLISHED"
	// ActivityStatusFinished 已结束
	ActivityStatusFinished ActivityStatus = "FINISHED"
	// ActivityStatusCancelled 已取消
	ActivityStatusCancelled ActivityStatus = "CANCELLED"
)
