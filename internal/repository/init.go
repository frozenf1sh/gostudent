package repository

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/frozenf1sh/gostudent/internal/model"
	"github.com/frozenf1sh/gostudent/pkg/fishlogger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GormInit() (db *gorm.DB) {
	var err error
	// 拼接dsn
	DSN := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.GlobalConfig.Database.Name,
		config.GlobalConfig.Database.Password,
		config.GlobalConfig.Database.Host,
		config.GlobalConfig.Database.Port,
		config.GlobalConfig.Database.Database)

	// 连接数据库，并设置gorm日志
	gormAdapter := fishlogger.NewGormSlogAdapter(fishlogger.AppLogger)
	db, err = gorm.Open(mysql.Open(DSN), &gorm.Config{
		Logger:      gormAdapter.LogMode(logger.Warn),
		PrepareStmt: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	//设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("无法获取底层DB对象", "reason", err)
	}
	sqlDB.SetMaxIdleConns(25)                 // 最大允许的空闲连接
	sqlDB.SetMaxOpenConns(100)                // 最大连接数
	sqlDB.SetConnMaxLifetime(20 * time.Hour)  // 示例：连接可复用 1 小时}
	sqlDB.SetConnMaxIdleTime(4 * time.Minute) // 示例：连接可复用 1 小时}

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		slog.Error("连接数据库失败", "reason", err)
		panic("连接数据库失败")
	}

	slog.Info("已连接到数据库")

	slog.Info("开始数据库自动迁移")
	err = db.AutoMigrate(&model.Admin{})
	err = errors.Join(err, db.AutoMigrate(&model.Activity{}))
	err = errors.Join(err, db.AutoMigrate(&model.Registration{}))
	if err != nil {
		slog.Error("数据库自动迁移失败", "reason", err)
		os.Exit(1)
	}
	slog.Info("数据库自动迁移成功！")

	return db
}
