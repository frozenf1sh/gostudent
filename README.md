# 学生活动管理系统

一个基于 Go + Gin 框架开发的学生活动管理系统，支持活动发布、报名、签到等功能。

## 功能特性

### 公共接口
- 活动列表查询
- 活动详情查询
- 活动报名
- 活动签到
- 签到Token获取

### 管理接口
- 管理员登录
- 活动CRUD（创建、查询、更新、删除）
- 活动发布
- 报名记录管理
- 签到状态修改
- 后台统计数据

### 系统功能
- 活动状态自动更新（自动将过期活动标记为已结束）
- JWT认证
- Redis缓存（用于存储签到Token）
- 日志记录

## 配置说明

项目配置文件 `config.yaml` 位于根目录，包含以下字段：

| 配置项 | 说明 |
|--------|------|
| `server.port` | 服务器监听端口 |
| `server.host` | 服务器监听地址 |
| `database.driver` | 数据库驱动（支持mysql） |
| `database.host` | 数据库地址 |
| `database.port` | 数据库端口 |
| `database.name` | 数据库名称 |
| `database.password` | 数据库密码 |
| `redis.host` | Redis地址 |
| `redis.port` | Redis端口 |
| `redis.password` | Redis密码 |
| `redis.db` | Redis数据库索引 |
| `jwt.secret` | JWT密钥 |
| `jwt.admin_expires_in` | 管理员Token过期时间 |
| `jwt.sign_in_expires_in` | 签到Token过期时间 |
| `log_file` | 日志文件路径 |
| `admin.username` | 默认管理员用户名 |
| `admin.password` | 默认管理员密码 |
| `cors.allow_origins` | 允许的跨域来源 |
| `cors.allow_methods` | 允许的HTTP方法 |
| `activity_status_update_interval` | 活动状态自动更新间隔 |

## API文档

### 公共接口（无需认证）

#### GET /api/v1/activities
活动列表查询

**请求参数（Query）：**
- `page`: 页码，默认1
- `page_size`: 每页大小，默认10
- `title`: 活动名称关键词过滤
- `type`: 活动类型过滤
- `status`: 活动状态过滤
- `date_from`: 开始时间范围过滤
- `date_to`: 结束时间范围过滤

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "admin_id": 1,
        "title": "Go语言技术分享会",
        "type": "技术讲座",
        "description": "分享Go语言的最新特性",
        "start_time": "2023-11-15T14:00:00+08:00",
        "end_time": "2023-11-15T16:00:00+08:00",
        "location": "学术报告厅",
        "registration_deadline": "2023-11-14T23:59:59+08:00",
        "max_participants": 100,
        "registered_count": 50,
        "status": "published",
        "live_url": "",
        "attachment_url": "",
        "created_at": "2023-11-10T10:00:00+08:00"
      }
    ],
    "total": 1,
    "page": 1
  }
}
```

---

#### GET /api/v1/activities/:activity_id
活动详情查询

**路径参数：**
- `activity_id`: 活动ID

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "admin_id": 1,
    "title": "Go语言技术分享会",
    "type": "技术讲座",
    "description": "分享Go语言的最新特性",
    "start_time": "2023-11-15T14:00:00+08:00",
    "end_time": "2023-11-15T16:00:00+08:00",
    "location": "学术报告厅",
    "registration_deadline": "2023-11-14T23:59:59+08:00",
    "max_participants": 100,
    "registered_count": 50,
    "status": "published",
    "live_url": "",
    "attachment_url": "",
    "created_at": "2023-11-10T10:00:00+08:00"
  }
}
```

---

#### POST /api/v1/activities/:activity_id/register
活动报名

**路径参数：**
- `activity_id`: 活动ID

**请求示例：**
```json
{
  "participant_name": "张三",
  "participant_phone": "13800138000",
  "participant_college": "计算机学院"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "activity_id": 1,
    "participant_name": "张三",
    "participant_phone": "13800138000",
    "participant_college": "计算机学院",
    "registered_at": "2023-11-10T15:30:00+08:00",
    "is_signed_in": false
  }
}
```

---

#### POST /api/v1/activities/:activity_id/signin
活动签到

**路径参数：**
- `activity_id`: 活动ID

**请求示例：**
```json
{
  "phone": "13800138000",
  "token": "abc123def456"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "签到成功",
  "data": null
}
```

---

#### GET /api/v1/activities/:activity_id/signin-token
获取签到Token

**路径参数：**
- `activity_id`: 活动ID

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "abc123def456"
  }
}
```

---

#### POST /api/v1/admin/login
管理员登录

**请求示例：**
```json
{
  "username": "admin",
  "password": "password"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "username": "admin"
  }
}
```

### 管理接口（需JWT认证）

#### POST /api/v1/admin/activities
创建活动

**请求示例：**
```json
{
  "title": "Go语言技术分享会",
  "type": "技术讲座",
  "description": "分享Go语言的最新特性",
  "start_time": "2023-11-15T14:00:00+08:00",
  "end_time": "2023-11-15T16:00:00+08:00",
  "location": "学术报告厅",
  "registration_deadline": "2023-11-14T23:59:59+08:00",
  "max_participants": 100,
  "live_url": "",
  "attachment_url": ""
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "admin_id": 1,
    "title": "Go语言技术分享会",
    "type": "技术讲座",
    "description": "分享Go语言的最新特性",
    "start_time": "2023-11-15T14:00:00+08:00",
    "end_time": "2023-11-15T16:00:00+08:00",
    "location": "学术报告厅",
    "registration_deadline": "2023-11-14T23:59:59+08:00",
    "max_participants": 100,
    "registered_count": 0,
    "status": "draft",
    "live_url": "",
    "attachment_url": "",
    "created_at": "2023-11-10T10:00:00+08:00"
  }
}
```

---

#### GET /api/v1/admin/activities
管理员查询所有活动

**请求参数（同公共接口的活动列表查询）**

---

#### GET /api/v1/admin/activities/:activity_id
管理员查询单个活动详情

**路径参数（同公共接口的活动详情查询）**

---

#### PUT /api/v1/admin/activities/:activity_id
更新活动

**路径参数：**
- `activity_id`: 活动ID

**请求示例（部分更新）：**
```json
{
  "title": "Go语言技术分享会（更新）",
  "max_participants": 150
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "admin_id": 1,
    "title": "Go语言技术分享会（更新）",
    "type": "技术讲座",
    "description": "分享Go语言的最新特性",
    "start_time": "2023-11-15T14:00:00+08:00",
    "end_time": "2023-11-15T16:00:00+08:00",
    "location": "学术报告厅",
    "registration_deadline": "2023-11-14T23:59:59+08:00",
    "max_participants": 150,
    "registered_count": 50,
    "status": "published",
    "live_url": "",
    "attachment_url": "",
    "created_at": "2023-11-10T10:00:00+08:00"
  }
}
```

---

#### DELETE /api/v1/admin/activities/:activity_id
删除活动

**路径参数：**
- `activity_id`: 活动ID

**响应示例：**
```json
{
  "code": 200,
  "message": "活动删除成功",
  "data": null
}
```

---

#### POST /api/v1/admin/activities/:activity_id/publish
发布活动

**路径参数：**
- `activity_id`: 活动ID

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "admin_id": 1,
    "title": "Go语言技术分享会",
    "type": "技术讲座",
    "description": "分享Go语言的最新特性",
    "start_time": "2023-11-15T14:00:00+08:00",
    "end_time": "2023-11-15T16:00:00+08:00",
    "location": "学术报告厅",
    "registration_deadline": "2023-11-14T23:59:59+08:00",
    "max_participants": 100,
    "registered_count": 50,
    "status": "published",
    "live_url": "",
    "attachment_url": "",
    "created_at": "2023-11-10T10:00:00+08:00"
  }
}
```

---

#### GET /api/v1/admin/activities/:activity_id/registrations
查询活动报名记录

**路径参数：**
- `activity_id`: 活动ID

**请求参数（Query）：**
- `page`: 页码，默认1
- `page_size`: 每页大小，默认10
- `phone`: 参与者手机号过滤
- `is_signed_in`: 签到状态过滤

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "activity_id": 1,
        "participant_name": "张三",
        "participant_phone": "13800138000",
        "participant_college": "计算机学院",
        "registered_at": "2023-11-10T15:30:00+08:00",
        "is_signed_in": false
      }
    ],
    "total": 1,
    "page": 1
  }
}
```

---

#### GET /api/v1/admin/registrations/:registration_id
查询单个报名记录

**路径参数：**
- `registration_id`: 报名记录ID

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "activity_id": 1,
    "participant_name": "张三",
    "participant_phone": "13800138000",
    "participant_college": "计算机学院",
    "registered_at": "2023-11-10T15:30:00+08:00",
    "is_signed_in": false
  }
}
```

---

#### GET /api/v1/admin/registrations
查询所有报名记录

**请求参数（Query）：**
- `page`: 页码，默认1
- `page_size`: 每页大小，默认10
- `activity_id`: 活动ID过滤
- `phone`: 参与者手机号过滤
- `is_signed_in`: 签到状态过滤

**响应示例：** 同活动报名记录查询

---

#### PUT /api/v1/admin/registrations/:registration_id/sign_in
修改签到状态

**路径参数：**
- `registration_id`: 报名记录ID

**请求示例：**
```json
{
  "is_signed_in": true
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "签到状态更新成功",
  "data": null
}
```

---

#### GET /api/v1/admin/dashboard
获取后台统计数据

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_activities": 10,
    "published_activities": 8,
    "total_registrations": 500,
    "today_registrations": 20
  }
}
```

## 运行说明

1. 确保已安装 Go 环境
2. 配置 `config.yaml` 文件
3. 执行以下命令运行项目：
   ```
   go run main.go
   ```

## 技术栈

- Go 1.25.1
- Gin 框架
- GORM  ORM
- MySQL
- Redis
- JWT
