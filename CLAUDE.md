# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview
A student activity management system built with Go, Gin, and GORM. The system manages activity lifecycle,报名 (registration), and 签到 (sign-in) for university events.

## High-Level Architecture

### Layered Architecture
The system follows a clear layered architecture with dependency injection:

```
┌─────────────────────────────────────────────────────────┐
│                      API Layer                         │
│  Internal: handler/                                    │
│  - admin.go, activity.go, registration.go, dashboard.go│
├─────────────────────────────────────────────────────────┤
│                     Service Layer                       │
│  Internal: service/                                     │
│  - admin_service.go, activity_service.go, registration_service.go │
├─────────────────────────────────────────────────────────┤
│                   Repository Layer                      │
│  Internal: repository/                                  │
│  - admin.go, activity.go, registration.go              │
├─────────────────────────────────────────────────────────┤
│                      Database                           │
│  MySQL database accessed via GORM                      │
└─────────────────────────────────────────────────────────┘
```

### Key Components

#### Entry Point
- `main.go`: Initializes the application, sets up dependencies, starts the web server, and runs background tasks.

#### Routing
- `internal/router/router.go`: Defines all API endpoints with authentication middleware.
  - Public endpoints: Activity list/query, registration, sign-in, admin login
  - Admin endpoints: Activity management (CRUD/publish), registration management, dashboard (JWT protected)

#### Background Tasks
- Activity status auto-updater: Runs periodically to update activity states (from `service/activity_service.go`)

#### Authentication
- JWT-based authentication for admin endpoints (`internal/middleware/jwt.go`)

#### Logging
- Dual-channel logging system (`pkg/fishlogger/`)

## Development Commands

### Run the Application
```bash
go run main.go
```

### Dependencies
```bash
go mod tidy
```

### Configuration
- `config.yaml`: Main configuration file for server settings, database connection, and JWT secret

### Project Structure
```
├── config.yaml
├── internal/
│   ├── config/      # Configuration loading
│   ├── handler/     # API handlers
│   ├── middleware/  # Middleware (CORS, JWT)
│   ├── model/       # Database models and DTOs
│   ├── repository/  # Data access layer
│   ├── router/      # API routing
│   └── service/     # Business logic layer
├── pkg/
│   ├── fishlogger/  # Logging package
│   └── utils/       # Utility functions
└── doc/             # Documentation
```

## Key API Endpoints

### Public Endpoints
- GET `/api/v1/activities` - List activities
- GET `/api/v1/activities/:activity_id` - Get activity by ID
- POST `/api/v1/activities/:activity_id/register` - Register for an activity
- POST `/api/v1/activities/:activity_id/signin` - Sign in to an activity
- POST `/api/v1/admin/login` - Admin login

### Admin Endpoints (JWT Protected)
- POST `/api/v1/admin/activities` - Create activity
- GET `/api/v1/admin/activities` - List all activities
- PUT `/api/v1/admin/activities/:activity_id` - Update activity
- DELETE `/api/v1/admin/activities/:activity_id` - Delete activity
- POST `/api/v1/admin/activities/:activity_id/publish` - Publish activity
- GET `/api/v1/admin/dashboard` - Dashboard statistics
