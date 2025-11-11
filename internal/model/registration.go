package model

import "time"

// Registration 对应 'registrations' 表，存储报名信息
// UniqueIndex约束：同一个活动(ActivityID)中，参与者手机号(ParticipantPhone)必须是唯一的。
type Registration struct {
	ID                 uint      `gorm:"primarykey"`
	ParticipantName    string    `gorm:"type:varchar(100);not null" json:"participant_name"`                                // 参与者姓名
	ParticipantPhone   string    `gorm:"type:varchar(20);uniqueIndex:idx_activity_phone;not null" json:"participant_phone"` // 参与者手机号
	ParticipantCollege string    `gorm:"type:varchar(100);not null" json:"participant_college"`                             // 参与者学院
	RegisteredAt       time.Time `gorm:"autoCreateTime" json:"registered_at"`                                               // 报名时间

	// 关联：活动
	ActivityID uint     `gorm:"uniqueIndex:idx_activity_phone;not null" json:"activity_id"` // 外键：活动ID
	Activity   Activity `gorm:"foreignKey:ActivityID" json:"activity"`

	// 可选的签到功能字段
	IsSignedIn bool      `gorm:"not null;default:false" json:"is_signed_in"` // 是否已签到
	SignedInAt time.Time `gorm:"null" json:"signed_in_at"`                   // 签到时间
}
