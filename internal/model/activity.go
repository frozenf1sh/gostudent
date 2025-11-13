package model

import "time"

// 定义活动的 4 种状态
type ActivityStatus string

const (
	ActivityStatusDraft     ActivityStatus = "DRAFT"     // 未发布
	ActivityStatusPublished ActivityStatus = "PUBLISHED" // 已发布报名中
	ActivityStatusClosed    ActivityStatus = "CLOSED"    // 已截止报名
	ActivityStatusFinished  ActivityStatus = "FINISHED"  // 活动已结束
)

// 对应 'activities' 表，存储活动信息
type Activity struct {
	// 活动相关
	ID          uint           `gorm:"primarykey"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`                 // 活动名称
	Type        string         `gorm:"type:varchar(50);not null" json:"type"`                   // 活动类型 (讲座, 宣讲会等)
	Description string         `gorm:"type:text" json:"description"`                            // 活动简介
	StartTime   time.Time      `gorm:"not null" json:"start_time"`                              // 活动时间
	EndTime     time.Time      `gorm:"not null" json:"end_time"`                                // 活动时间
	Location    string         `gorm:"type:varchar(255);not null" json:"location"`              // 活动地点
	Status      ActivityStatus `gorm:"type:varchar(20);not null;default:'DRAFT'" json:"status"` // 活动状态

	// 报名相关
	RegistrationDeadline time.Time `gorm:"not null" json:"registration_deadline"`      // 报名截止时间
	MaxParticipants      int       `gorm:"not null;default:0" json:"max_participants"` // 人数上限 (0表示不限制)
	RegisteredCount      int       `gorm:"not null;default:0" json:"registered_count"` // 已报名人数

	// 链接
	LiveURL       string `gorm:"type:varchar(512)" json:"live_url"`       // 直播链接
	AttachmentURL string `gorm:"type:varchar(512)" json:"attachment_url"` // 附件链接

	// 1对1关联：活动-管理员
	Admin   Admin `gorm:"foreignKey:AdminID" json:"admin"`
	AdminID uint  `gorm:"not null" json:"admin_id"` // 外键：创建活动的管理员ID

	// 1对n关联：活动-报名记录
	Registrations []Registration `gorm:"foreignKey:ActivityID;OnDelete:CASCADE" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
