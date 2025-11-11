# 高校学生活动全流程管理系统 - 数据结构定义

> 本文档定义了系统后端（Golang Gorm/Gin）使用的核心数据模型（Model）和数据传输对象（DTO），以及它们在 MySQL 数据库中的表结构。

---

## 1. 核心模型（Gorm Model & 数据库表）

> 这部分结构与 MySQL 数据库中的表结构一一对应，用于 Gorm 操作。

### 1.1 管理员模型 (Admin)

| 数据库表: `admins` |
| --- |

| 字段名 | 类型 (Go) | 类型 (MySQL) | 约束 | 描述 |
| --- | --- | --- | --- | --- |
| ID | `uint` | `BIGINT` | `PRIMARY KEY, AUTO_INCREMENT` | 管理员 ID |
| Username | `string` | `VARCHAR(100)` | `NOT NULL, UNIQUE` | 登录用户名 |
| PasswordHash | `string` | `VARCHAR(255)` | `NOT NULL` | 加密后的密码 |
| CreatedAt | `time.Time` | `DATETIME(3)` | | 创建时间 |
| UpdatedAt | `time.Time` | `DATETIME(3)` | | 更新时间 |

**Go 结构示例**（内部使用）:

```go
type Admin struct {
    ID           uint
    Username     string
    PasswordHash string
    // ... 省略时间戳
}
```

---

### 1.2 活动模型 (Activity)

| 数据库表: `activities` |
| --- |

| 字段名 | 类型 (Go) | 类型 (MySQL) | 约束 | 描述 |
| --- | --- | --- | --- | --- |
| ID | `uint` | `BIGINT` | `PRIMARY KEY, AUTO_INCREMENT` | 活动 ID |
| AdminID | `uint` | `BIGINT` | `FOREIGN KEY` | 创建者管理员 ID |
| Title | `string` | `VARCHAR(255)` | `NOT NULL` | 活动名称 |
| StartTime | `time.Time` | `DATETIME(3)` | `NOT NULL` | 活动开始时间 |
| RegistrationDeadline | `time.Time` | `DATETIME(3)` | `NOT NULL` | 报名截止时间 |
| MaxParticipants | `int` | `INT` | `DEFAULT 0` | 人数上限 (0 为不限) |
| RegisteredCount | `int` | `INT` | `DEFAULT 0` | 已报名人数 |
| Status | `model.ActivityStatus` | `VARCHAR(20)` | `DEFAULT 'DRAFT'` | 状态 (DRAFT/PUBLISHED/FINISHED/CANCELLED) |
| Location | `string` | `VARCHAR(255)` | `NOT NULL` | 地点 |
| Description | `string` | `TEXT` | | 简介 |
| ... | ... | ... | | 其它字段 |

**Go 结构示例**（内部使用）:

```go
type Activity struct {
    ID              uint
    AdminID         uint
    Title           string
    StartTime       time.Time
    MaxParticipants int
    RegisteredCount int
    Status          ActivityStatus
    // ... 省略其他字段
}
```

---

### 1.3 报名模型 (Registration)

| 数据库表: `registrations` |
| --- |

| 字段名 | 类型 (Go) | 类型 (MySQL) | 约束 | 描述 |
| --- | --- | --- | --- | --- |
| ID | `uint` | `BIGINT` | `PRIMARY KEY, AUTO_INCREMENT` | 报名记录 ID |
| ActivityID | `uint` | `BIGINT` | `FOREIGN KEY` | 关联活动 ID |
| ParticipantName | `string` | `VARCHAR(100)` | `NOT NULL` | 参与者姓名 |
| ParticipantPhone | `string` | `VARCHAR(20)` | `NOT NULL` | 参与者手机号 |
| ParticipantCollege | `string` | `VARCHAR(100)` | `NOT NULL` | 参与者学院 |
| RegisteredAt | `time.Time` | `DATETIME(3)` | | 报名时间 |
| IsSignedIn | `bool` | `TINYINT(1)` | `DEFAULT 0` | 是否已签到 |
| SignedInAt | `time.Time` | `DATETIME(3)` | `NULL` | 签到时间 |
| **Unique Key** | | | | `(activity_id, participant_phone)` 确保不重复报名 |

**Go 结构示例**（内部使用）:

```go
type Registration struct {
    ID               uint
    ActivityID       uint
    ParticipantName  string
    ParticipantPhone string // 唯一性检查的关键字段
    IsSignedIn       bool
    RegisteredAt     time.Time
    // ... 省略其他字段
}
```

---

## 2. 数据传输对象 (DTOs) 示例

> DTOs 用于定义 API 请求和响应的 JSON 格式。

### 2.1 创建活动请求 (CreateActivityRequest) - JSON Body

> 这是管理员在创建活动时需要提交的 JSON 数据。

| 字段名 | 类型 | 描述 | 校验规则 (Gin Binding) |
| --- | --- | --- | --- |
| title | `string` | 活动名称 | `required` |
| type | `string` | 活动类型 | `required` |
| start_time | `string` | 活动开始时间 (ISO 8601 格式) | `required` |
| registration_deadline | `string` | 报名截止时间 (ISO 8601 格式) | `required` |
| max_participants | `number` | 人数上限 | `gte=0` |
| location | `string` | 地点 | `required` |
| description | `string` | 简介 | |

**JSON 示例** (`POST /api/v1/admin/activities`):

```json
{
  "title": "2025届就业指导讲座",
  "type": "讲座",
  "description": "特邀行业专家分享职业规划经验。",
  "start_time": "2025-12-15T19:00:00+08:00",
  "location": "A楼报告厅",
  "registration_deadline": "2025-12-14T23:59:59+08:00",
  "max_participants": 200,
  "live_url": "http://live.example.com/job"
}
```

---

### 2.2 活动详情响应 (ActivityResponse) - JSON Response

> 这是前端获取活动详情时收到的 JSON 数据。

| 字段名 | 类型 | 描述 |
| --- | --- | --- |
| id | `number` | 活动 ID |
| title | `string` | 活动名称 |
| status | `string` | 活动状态 (PUBLISHED, DRAFT 等) |
| start_time | `string` | 活动开始时间 |
| registered_count | `number` | 已报名人数 |
| max_participants | `number` | 人数上限 |
| ... | ... | 其它字段 |

**JSON 示例** (`GET /api/v1/public/activities/123`):

```json
{
  "id": 123,
  "admin_id": 1,
  "title": "2025届就业指导讲座",
  "type": "讲座",
  "status": "PUBLISHED",
  "start_time": "2025-12-15T19:00:00+08:00",
  "location": "A楼报告厅",
  "registration_deadline": "2025-12-14T23:59:59+08:00",
  "max_participants": 200,
  "registered_count": 85,
  "created_at": "2025-11-01T10:00:00+08:00"
}
```

---

### 2.3 参与者报名请求 (CreateRegistrationRequest) - JSON Body

> 这是活动参与者在报名页面提交的 JSON 数据。

| 字段名 | 类型 | 描述 | 校验规则 (Gin Binding) |
| --- | --- | --- | --- |
| participant_name | `string` | 姓名 | `required` |
| participant_phone | `string` | 手机号 | `required` |
| participant_college | `string` | 学院/部门 | `required` |

**JSON 示例** (`POST /api/v1/public/activities/123/register`):

```json
{
  "participant_name": "张三",
  "participant_phone": "13812345678",
  "participant_college": "计算机学院"
}
```
