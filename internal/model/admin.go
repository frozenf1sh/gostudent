package model

import "time"

// Admin 对应 'admins' 表，存储管理员信息
type Admin struct {
	ID           uint      `gorm:"primarykey"`
	Username     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"` // 用户名
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`                    // 存储哈希后的密码
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系：一个管理员可以创建多个活动
	Activities []Activity `gorm:"foreignKey:AdminID" json:"-"`
}

// TableName 自定义Gorm的表名
func (Admin) TableName() string {
	return "admins"
}
