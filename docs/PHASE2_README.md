# Phase 2: JWT Authentication + PostgreSQL + Redis

**Status**: Foundation Complete ✅
**Version**: 1.0.0
**Date**: 2025-11-20

## 📋 Overview

Phase 2 implements enterprise-grade backend infrastructure with dual authentication, persistent storage, and caching layer.

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      CLIENT LAYER                                │
│  ┌──────────────────┐              ┌──────────────────┐         │
│  │  Mobile Apps     │              │  Admin Dashboard │         │
│  │  Firebase Auth   │              │  JWT Auth        │         │
│  └────────┬─────────┘              └────────┬─────────┘         │
└───────────┼──────────────────────────────────┼───────────────────┘
            │                                  │
            ↓                                  ↓
┌─────────────────────────────────────────────────────────────────┐
│                   API GATEWAY (Go Server)                        │
│  • Firebase Auth Middleware                                      │
│  • Admin JWT Middleware                                          │
│  • Rate Limiting                                                 │
│  • RBAC                                                          │
└─────────────────┬───────────────────────────┬───────────────────┘
                  │                           │
         ┌────────┴────────┐         ┌───────┴────────┐
         ↓                 ↓         ↓                ↓
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│   PostgreSQL     │  │      Redis       │  │    Firebase      │
│   + PostGIS      │  │                  │  │                  │
│                  │  │  • Sessions      │  │  • Anonymous     │
│  • Users         │  │  • Cache         │  │    Auth          │
│  • Admins        │  │  • Rate Limits   │  │  • Verification  │
│  • Sessions      │  │  • Pub/Sub       │  │                  │
│  • Analytics     │  │                  │  │                  │
└──────────────────┘  └──────────────────┘  └──────────────────┘
```

## 🎯 What's Implemented

### ✅ Database Schema (PostgreSQL)

**Created Migrations:**
- `001_create_users.up.sql` - Users table with Firebase UID mapping
- `002_create_admins_roles.up.sql` - Admin users and RBAC system
- `003_create_sessions.up.sql` - User sessions and station analytics

**Key Features:**
- UUID primary keys
- PostGIS support for geographic queries
- Automatic `updated_at` triggers
- Comprehensive indexes for performance
- Foreign key constraints
- RBAC (Role-Based Access Control)

**Tables Created:**
1. **users** - Mobile app users (Firebase authentication)
2. **admins** - Admin users (JWT authentication)
3. **roles** - Role and permission definitions
4. **user_sessions** - Station check-in sessions
5. **station_analytics** - Daily aggregated metrics

### ✅ Development Environment

**docker-compose.yml:**
- PostgreSQL 15 with PostGIS 3.3
- Redis 7
- Network isolation
- Volume persistence
- Health checks

**Configuration:**
- `.env.example` with all required variables
- Separate dev/prod configurations
- Secrets management ready

## 🚀 Quick Start

### 1. Prerequisites

```bash
# Required
- Docker & Docker Compose
- Go 1.23+
- Node.js 18+ (for tests)

# Optional (for production)
- Firebase project
- PostgreSQL 15+
- Redis 7+
```

### 2. Setup Development Environment

```bash
# Clone and navigate to project
cd messenger_protocol_research

# Copy environment variables
cp .env.example .env

# Edit .env and set your values
vi .env

# Start database and Redis
docker-compose up -d postgres redis

# Wait for services to be healthy
docker-compose ps

# Run migrations (TODO: implement migration tool)
# For now, manually run SQL files in migrations/ directory
```

### 3. Access Services

```bash
# PostgreSQL
psql -h localhost -p 5432 -U trainblink_user -d trainblink

# Redis CLI
redis-cli -h localhost -p 6379 -a dev_password

# Check health
docker-compose logs postgres
docker-compose logs redis
```

## 📦 Phase 2 Implementation Checklist

### Core Infrastructure
- [x] PostgreSQL database schema
- [x] Database migrations (up/down)
- [x] Redis cache configuration
- [x] Docker compose setup
- [x] Environment configuration template
- [ ] Database connection pool (Go)
- [ ] Redis service (Go)
- [ ] Configuration loader (Go)

### Authentication & Authorization
- [ ] Firebase Admin SDK integration
- [ ] JWT token service
- [ ] Firebase auth middleware
- [ ] Admin JWT middleware
- [ ] RBAC permission system
- [ ] Rate limiting middleware

### Services & Repositories
- [ ] User service + repository
- [ ] Admin service + repository
- [ ] Session service + repository
- [ ] Station service (update for PostgreSQL)

### API Endpoints
- [ ] `POST /api/v1/users/me` - Get user profile
- [ ] `PUT /api/v1/users/me` - Update profile
- [ ] `POST /api/v1/admin/login` - Admin login
- [ ] `POST /api/v1/admin/refresh` - Token refresh
- [ ] `GET /api/v1/admin/users` - List users
- [ ] `POST /api/v1/geofence/enter` - Enter station (update)
- [ ] `POST /api/v1/geofence/exit` - Exit station (update)

### Testing
- [ ] Unit tests for auth services
- [ ] Integration tests for API
- [ ] Load testing
- [ ] Migration rollback tests

### Documentation
- [x] Phase 2 README
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Database ER diagram
- [ ] Architecture decision records

## 🗂️ Project Structure (Planned)

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── auth/
│   │   ├── firebase.go             # Firebase authentication
│   │   └── jwt.go                  # JWT token management
│   ├── middleware/
│   │   ├── firebase_auth.go        # Firebase middleware
│   │   ├── admin_auth.go           # Admin JWT middleware
│   │   ├── rate_limit.go           # Rate limiting
│   │   ├── cors.go                 # CORS configuration
│   │   └── logging.go              # Request logging
│   ├── user/
│   │   ├── service.go              # User business logic
│   │   ├── repository.go           # User database operations
│   │   └── handler.go              # User HTTP handlers
│   ├── admin/
│   │   ├── service.go              # Admin business logic
│   │   ├── repository.go           # Admin database operations
│   │   └── handler.go              # Admin HTTP handlers
│   ├── session/
│   │   ├── service.go              # Session management
│   │   └── handler.go              # Session HTTP handlers
│   ├── cache/
│   │   └── redis.go                # Redis service
│   ├── database/
│   │   └── postgres.go             # PostgreSQL connection
│   └── config/
│       └── config.go               # Configuration management
├── migrations/
│   ├── 001_create_users.up.sql    ✅
│   ├── 001_create_users.down.sql  ✅
│   ├── 002_create_admins_roles.up.sql  ✅
│   ├── 002_create_admins_roles.down.sql  ✅
│   ├── 003_create_sessions.up.sql  ✅
│   └── 003_create_sessions.down.sql  ✅
├── go.mod
├── .env.example                    ✅
├── docker-compose.yml              ✅
└── Dockerfile
```

## 🔐 Security Features (Planned)

### Authentication
- **Firebase Anonymous Auth** for mobile users
- **JWT (HS256)** for admin users
- Token refresh mechanism
- Token revocation/blacklist

### Authorization
- Role-Based Access Control (RBAC)
- Permission wildcards (`user.*`, `*`)
- Per-endpoint permission checks

### Security Measures
- Password hashing (bcrypt, cost 12)
- Account lockout (5 failed attempts)
- Rate limiting (per-user, per-IP)
- HTTPS only (production)
- CORS whitelist
- SQL injection prevention (parameterized queries)
- Input validation and sanitization

## 📊 Database Schema

### Users Table
```sql
users (
    id UUID PRIMARY KEY,
    firebase_uid VARCHAR(128) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    avatar_emoji VARCHAR(10),
    is_active BOOLEAN DEFAULT true,
    is_banned BOOLEAN DEFAULT false,
    total_sessions INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE,
    ...
)
```

### Admins Table
```sql
admins (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(50),
    permissions JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    ...
)
```

### Roles
- **super_admin**: Full system access (`["*"]`)
- **admin**: User and content management
- **moderator**: Content moderation and reports

## 🧪 Testing

```bash
# Unit tests (when implemented)
go test ./internal/...

# Integration tests
go test -tags=integration ./tests/...

# Load tests
# Use artillery or k6 for load testing
```

## 🚀 Deployment (TODO)

### Build
```bash
docker build -t trainblink-server:latest .
```

### Production Environment
- Set strong JWT_SECRET_KEY (32+ chars)
- Use strong database passwords
- Enable SSL/TLS for PostgreSQL
- Use Redis password
- Set up Firebase in production mode
- Enable CORS whitelist

## 📝 Next Steps

1. **Implement Core Services** (Priority 1)
   - Database connection pooling
   - Redis service wrapper
   - Configuration loader

2. **Authentication Implementation** (Priority 1)
   - Firebase Admin SDK integration
   - JWT service with token generation/verification
   - Auth middlewares

3. **User/Admin Services** (Priority 2)
   - User CRUD operations
   - Admin CRUD operations
   - Session management with Redis

4. **API Endpoints** (Priority 2)
   - User profile endpoints
   - Admin authentication endpoints
   - Admin management endpoints

5. **Testing & Documentation** (Priority 3)
   - Unit tests
   - Integration tests
   - API documentation (Swagger)

## 📚 Resources

- [Complete PRD Document](./PRD_PHASE2_JWT_POSTGRES_REDIS.md)
- [Firebase Admin SDK - Go](https://firebase.google.com/docs/admin/setup)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- [GORM Documentation](https://gorm.io/docs/)
- [go-redis](https://redis.uptrace.dev/)

## 📄 License

Copyright (c) 2025 TrainBlink Team

---

**Phase 2 Status**: Foundation complete, ready for service implementation
**Last Updated**: 2025-11-20
