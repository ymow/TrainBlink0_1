# TrainBlink Matrix Bridge - Complete Implementation Roadmap v2.0

**Date**: 2025-11-20
**Status**: Backend API Development Phase
**Next Phase**: Mobile Application (Matrix SDK, Notifications)

---

## 📋 Table of Contents

1. [Current Implementation Status](#current-implementation-status)
2. [Revised Requirements](#revised-requirements)
3. [Architecture Overview](#architecture-overview)
4. [Database Schema](#database-schema)
5. [Implementation Roadmap](#implementation-roadmap)
6. [API Endpoints Specification](#api-endpoints-specification)
7. [Development Phases](#development-phases)

---

## 🎯 Current Implementation Status

### ✅ Completed Components

#### 1. MLS E2EE System (Phase 5)
```
✅ internal/mls/types.go                 - Complete MLS type system (RFC 9420)
✅ internal/mls/group_manager.go         - MLS group lifecycle management
✅ internal/mls/keypackage_service.go    - KeyPackage operations with rate limiting
✅ internal/mls/delivery_service.go      - Atomic message delivery
✅ internal/mls/handler.go               - HTTP API handlers (11 endpoints)
```

**Key Features:**
- RFC 9420 compliant MLS protocol
- Atomic sequence numbering (Redis INCR)
- Transaction-safe KeyPackage claiming
- Rate limiting (100 uploads/hour)
- 7 cipher suites supported

#### 2. Matrix Client SDK (Phase 4)
```
✅ internal/matrix/client.go             - HTTP client for Matrix C-S API
✅ internal/matrix/types.go              - Matrix event types & structures
✅ internal/matrix/sync_manager.go       - /sync loop (30s long-polling)
✅ internal/matrix/event_handler.go      - Event processing framework
✅ internal/matrix/room_manager.go       - Room management (in-memory)
✅ internal/matrix/user_manager.go       - User registration (in-memory)
✅ internal/matrix/bridge.go             - Basic bridge service (mock)
```

**Key Features:**
- Complete Matrix Client-Server API coverage
- Event system (timeline, state, ephemeral)
- Anonymous user generation
- Station → Room mapping

#### 3. Infrastructure
```
✅ PostgreSQL + PostGIS                  - Database foundation
✅ Redis                                 - Caching & rate limiting
✅ JWT Authentication                    - Token-based auth
✅ Admin System                          - Admin users & roles
✅ Geofence Foundation                   - Basic geofence models
```

---

## 🔄 Revised Requirements (v2.0)

### 🚫 Removed: Strapi CMS Integration

**Original Plan:**
- ❌ Strapi as headless CMS
- ❌ Strapi webhooks for content publishing
- ❌ External content management

### ✅ New: Custom Agentic Content Management System

**New Requirements:**

#### 1. Content Types
```yaml
Icebreaker Cards:
  - ID (UUID)
  - Title (multilingual: EN, ZH, JA)
  - Summary (rich text)
  - Category (enum: TECH, CULTURE, FOOD, TRAVEL, etc.)
  - Cover Image URL
  - Content Body (rich text, supports embeds)
  - Source URL (optional)
  - Source Name (optional)
  - Tags (array)
  - Geo Tags (station IDs)
  - Icebreaker Score (0-10)
  - Visibility (draft, published, archived)
  - Published At
  - Expires At (optional)
  - Author (admin user)
  - Created/Updated timestamps

Blog Posts:
  - ID (UUID)
  - Title (multilingual)
  - Slug (URL-friendly)
  - Excerpt (short summary)
  - Content (rich text via Puck)
  - Cover Image
  - Category
  - Tags
  - Author
  - SEO (meta title, description, keywords)
  - Visibility
  - Published At
  - View Count
  - Like Count
  - Created/Updated timestamps

Announcements:
  - ID (UUID)
  - Title
  - Message (rich text)
  - Priority (low, medium, high, urgent)
  - Target Audience (all, specific_stations, specific_users)
  - Target Station IDs (array)
  - Start Time
  - End Time
  - Created By (admin)
  - Created/Updated timestamps
```

#### 2. Puck Editor Integration (React Frontend)

**Why Puck?**
- Visual page builder for React
- Component-based editing
- Custom components support
- JSON output (easy to store/retrieve)
- Headless architecture (perfect for our API-first approach)

**Architecture:**
```
Admin Dashboard (React + Puck)
    ↓ (HTTP API)
Go Backend API
    ↓ (PostgreSQL)
Content Database
    ↓ (Matrix Bridge)
Station Rooms (Matrix)
```

**Puck Integration Points:**
```typescript
// Custom components for TrainBlink
const puckConfig = {
  components: {
    IcebreakerCard: {
      fields: {
        title: { type: "text" },
        summary: { type: "textarea" },
        category: { type: "select", options: [...] },
        coverImage: { type: "custom", render: ImageUpload },
        geoTags: { type: "custom", render: StationSelector },
        icebreakerScore: { type: "number", min: 0, max: 10 }
      }
    },
    BlogPost: { ... },
    Announcement: { ... }
  }
}
```

#### 3. Content Publishing Workflow

```
┌─────────────────────────────────────────────────────┐
│              Admin Dashboard (React)                │
│                                                     │
│  ┌─────────────────────────────────────────────┐  │
│  │    Puck Editor                              │  │
│  │    - Visual editing                         │  │
│  │    - Component selection                    │  │
│  │    - Preview                                │  │
│  │    - Geo-targeting (station selection)     │  │
│  └─────────────────────────────────────────────┘  │
│                       ↓                             │
│         Save Draft | Publish | Schedule            │
└─────────────────┬───────────────────────────────────┘
                  ↓ POST /api/v1/admin/content/publish
┌─────────────────────────────────────────────────────┐
│              Go Backend API                         │
│                                                     │
│  ┌─────────────────────────────────────────────┐  │
│  │ Content Service                             │  │
│  │ - Validate content                          │  │
│  │ - Store in PostgreSQL                       │  │
│  │ - Trigger Matrix publishing                 │  │
│  └──────────────┬──────────────────────────────┘  │
│                 ↓                                   │
│  ┌─────────────────────────────────────────────┐  │
│  │ Matrix Content Publisher                    │  │
│  │ - Format content as Matrix messages         │  │
│  │ - Geo-target specific stations              │  │
│  │ - Send to Matrix rooms                      │  │
│  │ - Track publication metrics                 │  │
│  └──────────────┬──────────────────────────────┘  │
└─────────────────┼───────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────────────────┐
│          Matrix Homeserver (Synapse)                │
│                                                     │
│   Station Rooms:                                   │
│   - #tokyo-station:trainblink.org                 │
│   - #taipei-station:trainblink.org                │
│   - #osaka-station:trainblink.org                 │
│                                                     │
│   Content appears as Matrix messages               │
└─────────────────────────────────────────────────────┘
```

---

## 🏗 Architecture Overview

### System Components

```
┌────────────────────────────────────────────────────────────┐
│                    ADMIN LAYER                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Admin Dashboard (React + Puck Editor)                │  │
│  │ - Content creation & editing                         │  │
│  │ - Station management                                 │  │
│  │ - Analytics & monitoring                             │  │
│  │ - User management                                    │  │
│  └──────────────────────────────────────────────────────┘  │
└───────────────────────┬────────────────────────────────────┘
                        ↓ HTTPS/REST API
┌────────────────────────────────────────────────────────────┐
│                   BACKEND API LAYER                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Go Server (Gin Framework)                            │  │
│  │                                                       │  │
│  │ ┌─────────────────┐  ┌──────────────────────────┐  │  │
│  │ │ Content Service │  │ Matrix Bridge Service    │  │  │
│  │ │ - CRUD          │  │ - Room Manager           │  │  │
│  │ │ - Publishing    │  │ - User Manager           │  │  │
│  │ │ - Geo-targeting │  │ - Content Publisher      │  │  │
│  │ └─────────────────┘  │ - Geofence Integration   │  │  │
│  │                       └──────────────────────────┘  │  │
│  │                                                       │  │
│  │ ┌─────────────────┐  ┌──────────────────────────┐  │  │
│  │ │ Geofence Svc    │  │ MLS E2EE Service         │  │  │
│  │ │ - Enter/Exit    │  │ - Group Management       │  │  │
│  │ │ - Matrix Notify │  │ - KeyPackage Service     │  │  │
│  │ └─────────────────┘  │ - Delivery Service       │  │  │
│  │                       └──────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
└───────────────────────┬────────────────────────────────────┘
                        ↓
┌────────────────────────────────────────────────────────────┐
│                   DATA LAYER                                │
│  ┌──────────────────┐  ┌───────────────┐  ┌────────────┐  │
│  │ PostgreSQL       │  │ Redis         │  │ Matrix     │  │
│  │ - Content        │  │ - Rate Limit  │  │ Synapse    │  │
│  │ - Matrix Rooms   │  │ - Sessions    │  │ - Rooms    │  │
│  │ - Users          │  │ - Cache       │  │ - Events   │  │
│  │ - Geofences      │  │ - MLS Seq#    │  │ - E2EE     │  │
│  │ - MLS Data       │  └───────────────┘  └────────────┘  │
│  └──────────────────┘                                      │
└────────────────────────────────────────────────────────────┘
```

### Data Flow

#### Content Publishing Flow
```
1. Admin creates content in Puck Editor
   → Rich content with components
   → Select target stations (geo-targeting)
   → Preview & publish

2. POST /api/v1/admin/content/publish
   → Validate content structure
   → Save to PostgreSQL (content table)
   → Extract geo-tags

3. Content Publisher Service
   → For each target station:
      a. Get or create Matrix room
      b. Format content as Matrix message
      c. Send to room via Matrix API
      d. Record publication

4. Matrix Synapse
   → Stores event in room
   → Syncs to connected clients
   → Delivers to mobile apps

5. Mobile Apps (Future Phase)
   → Receive via Matrix SDK
   → Display in station feed
   → Trigger notifications
```

#### Geofence → Matrix Flow
```
1. User enters station geofence
   → GPS triggers geofence service
   → POST /api/v1/geofence/enter

2. Geofence Handler
   → Validate station & user
   → Trigger Matrix integration

3. Matrix Bridge
   → Create/get Matrix user for TrainBlink user
   → Get or create station room
   → Join user to room
   → Return Matrix credentials to app

4. Mobile App
   → Initialize Matrix SDK with credentials
   → Connect to room
   → Start receiving messages
```

---

## 🗄 Database Schema

### Content Management Tables

```sql
-- ============================================================
-- Content: Icebreaker Cards
-- ============================================================

CREATE TABLE icebreaker_cards (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Multilingual Content
  title_en VARCHAR(200) NOT NULL,
  title_zh VARCHAR(200),
  title_ja VARCHAR(200),

  summary_en TEXT NOT NULL,
  summary_zh TEXT,
  summary_ja TEXT,

  content_body JSONB NOT NULL,                    -- Puck editor output

  -- Metadata
  category VARCHAR(50) NOT NULL,                  -- TECH, CULTURE, FOOD, etc.
  tags TEXT[] DEFAULT '{}',
  source_url TEXT,
  source_name VARCHAR(200),
  cover_image_url TEXT,

  -- Geo-targeting
  geo_tags TEXT[] NOT NULL,                       -- Station IDs

  -- Scoring
  icebreaker_score DECIMAL(3,1) CHECK (icebreaker_score BETWEEN 0 AND 10),

  -- Publishing
  visibility VARCHAR(20) DEFAULT 'draft',         -- draft, published, archived
  published_at TIMESTAMP WITH TIME ZONE,
  expires_at TIMESTAMP WITH TIME ZONE,

  -- Authorship
  author_id UUID REFERENCES admins(id),

  -- Analytics
  view_count BIGINT DEFAULT 0,
  share_count BIGINT DEFAULT 0,

  -- Timestamps
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_icebreaker_cards_visibility ON icebreaker_cards(visibility);
CREATE INDEX idx_icebreaker_cards_published ON icebreaker_cards(published_at);
CREATE INDEX idx_icebreaker_cards_category ON icebreaker_cards(category);
CREATE INDEX idx_icebreaker_cards_geo_tags ON icebreaker_cards USING GIN(geo_tags);
CREATE INDEX idx_icebreaker_cards_author ON icebreaker_cards(author_id);

-- ============================================================
-- Content: Blog Posts
-- ============================================================

CREATE TABLE blog_posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Multilingual Content
  title_en VARCHAR(200) NOT NULL,
  title_zh VARCHAR(200),
  title_ja VARCHAR(200),

  slug VARCHAR(200) UNIQUE NOT NULL,

  excerpt_en TEXT,
  excerpt_zh TEXT,
  excerpt_ja TEXT,

  content_body JSONB NOT NULL,                    -- Puck editor output

  -- Metadata
  category VARCHAR(50),
  tags TEXT[] DEFAULT '{}',
  cover_image_url TEXT,

  -- SEO
  meta_title VARCHAR(200),
  meta_description TEXT,
  meta_keywords TEXT[],

  -- Publishing
  visibility VARCHAR(20) DEFAULT 'draft',
  published_at TIMESTAMP WITH TIME ZONE,

  -- Authorship
  author_id UUID REFERENCES admins(id),

  -- Analytics
  view_count BIGINT DEFAULT 0,
  like_count BIGINT DEFAULT 0,

  -- Timestamps
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_blog_posts_slug ON blog_posts(slug);
CREATE INDEX idx_blog_posts_visibility ON blog_posts(visibility);
CREATE INDEX idx_blog_posts_published ON blog_posts(published_at);
CREATE INDEX idx_blog_posts_author ON blog_posts(author_id);

-- ============================================================
-- Content: Announcements
-- ============================================================

CREATE TABLE announcements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  title VARCHAR(200) NOT NULL,
  message TEXT NOT NULL,

  -- Targeting
  priority VARCHAR(20) DEFAULT 'medium',          -- low, medium, high, urgent
  target_audience VARCHAR(50) DEFAULT 'all',      -- all, specific_stations, specific_users
  target_station_ids TEXT[],
  target_user_ids UUID[],

  -- Scheduling
  start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  end_time TIMESTAMP WITH TIME ZONE,

  -- Authorship
  created_by UUID REFERENCES admins(id),

  -- Status
  is_active BOOLEAN DEFAULT true,

  -- Timestamps
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_announcements_active ON announcements(is_active);
CREATE INDEX idx_announcements_priority ON announcements(priority);
CREATE INDEX idx_announcements_start_time ON announcements(start_time);
CREATE INDEX idx_announcements_target_stations ON announcements USING GIN(target_station_ids);
```

### Matrix Integration Tables

```sql
-- ============================================================
-- Matrix: Rooms (Station → Matrix Room Mapping)
-- ============================================================

CREATE TABLE matrix_rooms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  station_id VARCHAR(50) NOT NULL UNIQUE,         -- Reference to stations table
  room_id VARCHAR(255) NOT NULL UNIQUE,           -- !abc123:trainblink.org
  room_alias VARCHAR(255) UNIQUE,                 -- #tokyo-station:trainblink.org

  -- Room Info
  room_name VARCHAR(200) NOT NULL,
  room_topic TEXT,
  room_avatar_url TEXT,

  -- Settings
  is_public BOOLEAN DEFAULT true,
  is_encrypted BOOLEAN DEFAULT false,
  join_rule VARCHAR(50) DEFAULT 'public',         -- public, invite
  history_visibility VARCHAR(50) DEFAULT 'shared', -- shared, world_readable

  -- MLS Integration
  mls_group_id VARCHAR(255),                      -- Link to MLS group
  mls_enabled BOOLEAN DEFAULT false,

  -- Statistics
  member_count INTEGER DEFAULT 0,
  message_count BIGINT DEFAULT 0,
  last_message_at TIMESTAMP WITH TIME ZONE,

  -- Status
  is_active BOOLEAN DEFAULT true,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_matrix_rooms_station ON matrix_rooms(station_id);
CREATE INDEX idx_matrix_rooms_room_id ON matrix_rooms(room_id);
CREATE INDEX idx_matrix_rooms_alias ON matrix_rooms(room_alias);
CREATE INDEX idx_matrix_rooms_active ON matrix_rooms(is_active);

-- ============================================================
-- Matrix: Users (TrainBlink User → Matrix User Mapping)
-- ============================================================

CREATE TABLE matrix_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  user_id UUID NOT NULL UNIQUE REFERENCES users(id),
  matrix_user_id VARCHAR(255) NOT NULL UNIQUE,   -- @trainblink_xxx:trainblink.org
  display_name VARCHAR(100),
  avatar_url TEXT,

  -- Authentication
  access_token TEXT NOT NULL,                     -- Encrypted Matrix access token
  device_id VARCHAR(255),

  -- Status
  is_active BOOLEAN DEFAULT true,
  last_seen_at TIMESTAMP WITH TIME ZONE,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '30 days'
);

CREATE INDEX idx_matrix_users_user_id ON matrix_users(user_id);
CREATE INDEX idx_matrix_users_matrix_id ON matrix_users(matrix_user_id);
CREATE INDEX idx_matrix_users_active ON matrix_users(is_active);

-- ============================================================
-- Matrix: Room Memberships
-- ============================================================

CREATE TABLE matrix_room_memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  room_id UUID NOT NULL REFERENCES matrix_rooms(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  membership VARCHAR(50) NOT NULL,                -- join, leave, invite, ban
  reason TEXT,

  joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  left_at TIMESTAMP WITH TIME ZONE,

  UNIQUE(room_id, user_id)
);

CREATE INDEX idx_matrix_memberships_room ON matrix_room_memberships(room_id);
CREATE INDEX idx_matrix_memberships_user ON matrix_room_memberships(user_id);
CREATE INDEX idx_matrix_memberships_status ON matrix_room_memberships(membership);

-- ============================================================
-- Matrix: Content Publications (Tracking)
-- ============================================================

CREATE TABLE matrix_content_publications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Source Content
  content_type VARCHAR(50) NOT NULL,              -- icebreaker_card, blog_post, announcement
  content_id UUID NOT NULL,

  -- Matrix Destination
  room_id UUID NOT NULL REFERENCES matrix_rooms(id),
  event_id VARCHAR(255) NOT NULL,                 -- Matrix event ID

  -- Engagement Metrics
  view_count INTEGER DEFAULT 0,
  reaction_count INTEGER DEFAULT 0,
  reply_count INTEGER DEFAULT 0,

  published_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_matrix_pubs_content ON matrix_content_publications(content_type, content_id);
CREATE INDEX idx_matrix_pubs_room ON matrix_content_publications(room_id);
CREATE INDEX idx_matrix_pubs_event ON matrix_content_publications(event_id);
```

---

## 🗺 Implementation Roadmap

### Phase 1: Database Foundation (Week 1)
**Goal**: Set up database schema and migrations

**Tasks:**
1. ✅ Create migration files for content tables
   - `005_create_content_tables.up.sql`
   - `006_create_matrix_tables.up.sql`

2. ✅ Run migrations and verify schema

3. ✅ Create GORM models for all tables
   - `internal/model/content.go`
   - Update `internal/model/matrix.go`

4. ✅ Update docker-compose.yml with Matrix Synapse

**Deliverables:**
- [ ] All database tables created
- [ ] Migrations tested (up/down)
- [ ] GORM models defined

---

### Phase 2: Matrix Infrastructure (Week 2)
**Goal**: Set up Matrix Synapse and database-backed Matrix services

**Tasks:**
1. ✅ Configure Matrix Synapse in docker-compose
   - Synapse homeserver.yaml
   - Application Service registration
   - PostgreSQL backend

2. ✅ Implement database-backed Room Manager
   - Replace in-memory with PostgreSQL
   - GetOrCreateStationRoom()
   - JoinUserToRoom() / RemoveUserFromRoom()

3. ✅ Implement database-backed User Manager
   - GetOrCreateMatrixUser()
   - Token management
   - User cleanup service

4. ✅ Test Matrix integration
   - Create rooms via API
   - Register users
   - Send test messages

**Deliverables:**
- [ ] Matrix Synapse running
- [ ] Database-backed Matrix services
- [ ] Integration tests passing

---

### Phase 3: Content Management Service (Week 3)
**Goal**: Build backend CRUD for content

**Tasks:**
1. ✅ Create Content Service
   - `internal/content/service.go`
   - CRUD operations for icebreaker cards
   - CRUD operations for blog posts
   - CRUD operations for announcements

2. ✅ Implement Content Publisher
   - `internal/content/publisher.go`
   - Format content as Matrix messages
   - Geo-targeting logic
   - Publish to station rooms
   - Track publications

3. ✅ Add validation & business logic
   - Content validation
   - Geo-tag validation
   - Publishing rules

**Deliverables:**
- [ ] Content service implemented
- [ ] Content can be published to Matrix
- [ ] Unit tests for content operations

---

### Phase 4: Geofence → Matrix Integration (Week 4)
**Goal**: Auto-join/leave rooms based on geofence

**Tasks:**
1. ✅ Update Geofence Handler
   - `internal/geofence/service.go`
   - Integrate with Matrix Bridge
   - OnStationEnter() → Join Matrix room
   - OnStationExit() → Leave Matrix room

2. ✅ Implement Geofence → Matrix Bridge
   - `internal/matrix/geofence_bridge.go`
   - HandleStationEnter()
   - HandleStationExit()
   - Cleanup inactive users

3. ✅ Test geofence flow
   - Simulate station entry
   - Verify Matrix room join
   - Simulate station exit
   - Verify Matrix room leave

**Deliverables:**
- [ ] Geofence triggers Matrix actions
- [ ] Auto-join/leave working
- [ ] Integration tests passing

---

### Phase 5: API Implementation (Week 5-6)
**Goal**: Complete REST API for all features

#### 5.1 Content API (Admin)
```
POST   /api/v1/admin/content/cards              # Create icebreaker card
GET    /api/v1/admin/content/cards              # List cards
GET    /api/v1/admin/content/cards/:id          # Get card
PUT    /api/v1/admin/content/cards/:id          # Update card
DELETE /api/v1/admin/content/cards/:id          # Delete card
POST   /api/v1/admin/content/cards/:id/publish  # Publish card

POST   /api/v1/admin/content/posts              # Create blog post
GET    /api/v1/admin/content/posts              # List posts
GET    /api/v1/admin/content/posts/:id          # Get post
PUT    /api/v1/admin/content/posts/:id          # Update post
DELETE /api/v1/admin/content/posts/:id          # Delete post
POST   /api/v1/admin/content/posts/:id/publish  # Publish post

POST   /api/v1/admin/announcements              # Create announcement
GET    /api/v1/admin/announcements              # List announcements
GET    /api/v1/admin/announcements/:id          # Get announcement
PUT    /api/v1/admin/announcements/:id          # Update announcement
DELETE /api/v1/admin/announcements/:id          # Delete announcement
POST   /api/v1/admin/announcements/:id/broadcast # Broadcast
```

#### 5.2 Matrix API (User)
```
POST   /api/v1/matrix/register                  # Get/create Matrix user
GET    /api/v1/matrix/rooms                     # Get user's rooms
POST   /api/v1/matrix/rooms/:id/join            # Join room
POST   /api/v1/matrix/rooms/:id/leave           # Leave room
GET    /api/v1/matrix/rooms/:id/messages        # Get room messages
```

#### 5.3 Geofence API (Internal)
```
POST   /api/v1/geofence/enter                   # Station entry
POST   /api/v1/geofence/exit                    # Station exit
GET    /api/v1/geofence/status                  # User geofence status
```

#### 5.4 Public Content API (Mobile App)
```
GET    /api/v1/content/cards                    # Get published cards
GET    /api/v1/content/cards/:id                # Get card details
GET    /api/v1/content/posts                    # Get published posts
GET    /api/v1/content/posts/:slug              # Get post by slug
GET    /api/v1/announcements                    # Get active announcements
```

**Tasks:**
1. ✅ Implement Content API handlers
   - `internal/api/content_handler.go`
   - Full CRUD operations
   - Publishing endpoints
   - Validation middleware

2. ✅ Implement Matrix API handlers
   - `internal/api/matrix_handler.go`
   - User registration
   - Room operations
   - Message retrieval

3. ✅ Implement Geofence API handlers
   - Update `internal/api/geofence_handler.go`
   - Station entry/exit
   - Status queries

4. ✅ Add authentication middleware
   - Admin auth for content APIs
   - User auth for Matrix APIs
   - Internal auth for geofence APIs

5. ✅ API documentation
   - OpenAPI/Swagger spec
   - Postman collection

**Deliverables:**
- [ ] All API endpoints implemented
- [ ] Authentication working
- [ ] API documentation complete
- [ ] Postman tests passing

---

### Phase 6: Testing & Documentation (Week 7)
**Goal**: Comprehensive testing and documentation

**Tasks:**
1. ✅ Unit tests
   - Content service tests
   - Matrix service tests
   - Geofence integration tests

2. ✅ Integration tests
   - Full content publishing flow
   - Geofence → Matrix flow
   - End-to-end scenarios

3. ✅ Documentation
   - API documentation
   - Deployment guide
   - Admin guide

**Deliverables:**
- [ ] 80%+ test coverage
- [ ] All integration tests passing
- [ ] Documentation complete

---

## 🔌 API Endpoints Specification

### Content Management API

#### Create Icebreaker Card
```http
POST /api/v1/admin/content/cards
Authorization: Bearer {admin_token}
Content-Type: application/json

Request:
{
  "title_en": "Best Ramen Spots Near Tokyo Station",
  "summary_en": "Discover authentic ramen within 5 minutes walk",
  "content_body": {
    "root": {
      "type": "IcebreakerCard",
      "props": { ... }  // Puck editor output
    }
  },
  "category": "FOOD",
  "tags": ["ramen", "tokyo", "food"],
  "geo_tags": ["station_tokyo_001"],
  "icebreaker_score": 8.5,
  "cover_image_url": "https://cdn.trainblink.org/images/ramen.jpg",
  "visibility": "draft"
}

Response (201):
{
  "status": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title_en": "Best Ramen Spots Near Tokyo Station",
    "visibility": "draft",
    "created_at": "2025-11-20T10:00:00Z"
  }
}
```

#### Publish Card
```http
POST /api/v1/admin/content/cards/:id/publish
Authorization: Bearer {admin_token}

Request:
{
  "schedule_at": "2025-11-21T09:00:00Z"  // Optional, immediate if not provided
}

Response (200):
{
  "status": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "visibility": "published",
    "published_at": "2025-11-20T10:00:00Z",
    "published_to_rooms": [
      {
        "room_id": "!tokyo001:trainblink.org",
        "room_alias": "#tokyo-station:trainblink.org",
        "event_id": "$event123:trainblink.org"
      }
    ]
  }
}
```

#### List Cards
```http
GET /api/v1/admin/content/cards?page=1&limit=20&visibility=published&category=FOOD
Authorization: Bearer {admin_token}

Response (200):
{
  "status": "success",
  "data": {
    "cards": [
      {
        "id": "...",
        "title_en": "...",
        "category": "FOOD",
        "visibility": "published",
        "icebreaker_score": 8.5,
        "view_count": 1234,
        "created_at": "...",
        "published_at": "..."
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

### Matrix User API

#### Register/Get Matrix User
```http
POST /api/v1/matrix/register
Authorization: Bearer {user_token}

Response (200):
{
  "status": "success",
  "data": {
    "matrix_user_id": "@trainblink_abc123:trainblink.org",
    "access_token": "syt_...",
    "device_id": "DEVICEABC",
    "homeserver_url": "https://matrix.trainblink.org",
    "expires_at": "2025-12-20T10:00:00Z"
  }
}
```

#### Get User's Rooms
```http
GET /api/v1/matrix/rooms
Authorization: Bearer {user_token}

Response (200):
{
  "status": "success",
  "data": {
    "rooms": [
      {
        "room_id": "!tokyo001:trainblink.org",
        "room_alias": "#tokyo-station:trainblink.org",
        "name": "Tokyo Station",
        "topic": "TrainBlink station room for Tokyo",
        "member_count": 42,
        "last_message_at": "2025-11-20T09:55:00Z",
        "is_encrypted": false,
        "mls_enabled": false
      }
    ],
    "total": 1
  }
}
```

### Geofence API

#### Station Entry
```http
POST /api/v1/geofence/enter
Authorization: Bearer {user_token}
Content-Type: application/json

Request:
{
  "station_id": "station_tokyo_001",
  "latitude": 35.6812,
  "longitude": 139.7671,
  "timestamp": "2025-11-20T10:00:00Z"
}

Response (200):
{
  "status": "success",
  "data": {
    "station": {
      "id": "station_tokyo_001",
      "name": "Tokyo Station"
    },
    "matrix": {
      "room_id": "!tokyo001:trainblink.org",
      "room_alias": "#tokyo-station:trainblink.org",
      "joined": true,
      "member_count": 43
    }
  }
}
```

### Public Content API

#### Get Published Cards
```http
GET /api/v1/content/cards?station_id=station_tokyo_001&limit=10
Authorization: Bearer {user_token}

Response (200):
{
  "status": "success",
  "data": {
    "cards": [
      {
        "id": "...",
        "title_en": "Best Ramen Spots Near Tokyo Station",
        "summary_en": "Discover authentic ramen...",
        "category": "FOOD",
        "cover_image_url": "...",
        "icebreaker_score": 8.5,
        "tags": ["ramen", "tokyo", "food"],
        "published_at": "2025-11-20T10:00:00Z"
      }
    ],
    "total": 5
  }
}
```

---

## 📅 Development Phases Summary

### Current Status: ✅ MLS E2EE Complete

**Completed:**
- MLS group management
- KeyPackage service
- Delivery service
- MLS API endpoints

**Next:** Database Foundation & Matrix Infrastructure

### Timeline

```
Week 1: Database Foundation
├── Migration files
├── GORM models
└── Docker compose update

Week 2: Matrix Infrastructure
├── Synapse setup
├── Database-backed Room Manager
├── Database-backed User Manager
└── Integration tests

Week 3: Content Management
├── Content service (CRUD)
├── Content publisher
└── Validation logic

Week 4: Geofence Integration
├── Update geofence service
├── Matrix bridge
└── Auto-join/leave

Week 5-6: API Implementation
├── Content API
├── Matrix API
├── Geofence API
├── Public API
└── Documentation

Week 7: Testing & Polish
├── Unit tests
├── Integration tests
└── Documentation

Week 8: Admin Dashboard (Puck)
├── React setup
├── Puck integration
├── Admin interface
└── Content editor
```

---

## 🚀 Next Phase: Mobile Application

**After backend API completion:**

1. **iOS Application**
   - Matrix Swift SDK integration
   - Real-time messaging
   - Push notifications (APNs)
   - Geofence monitoring
   - Content feed

2. **Android Application**
   - Matrix Android SDK integration
   - Real-time messaging
   - Push notifications (FCM)
   - Geofence monitoring
   - Content feed

3. **Features**
   - Anonymous chat in station rooms
   - Icebreaker cards feed
   - Notifications for new content
   - MLS E2EE (when entering encrypted rooms)
   - User profile management

---

## ✅ Success Criteria

### Backend API (This Phase)
- [ ] All database tables created and migrated
- [ ] Matrix Synapse running and integrated
- [ ] Content CRUD operations working
- [ ] Content publishing to Matrix working
- [ ] Geofence triggers Matrix auto-join/leave
- [ ] All API endpoints implemented and documented
- [ ] 80%+ test coverage
- [ ] API deployed and accessible

### Admin Dashboard (Puck)
- [ ] React app running
- [ ] Puck editor integrated
- [ ] Can create/edit icebreaker cards
- [ ] Can publish content to stations
- [ ] Analytics dashboard working

### Ready for Mobile Development
- [ ] API stable and documented
- [ ] Matrix rooms working correctly
- [ ] Geofence integration tested
- [ ] Performance benchmarks met
- [ ] Security review passed

---

**Document Maintainer**: TrainBlink Backend Team
**Last Updated**: 2025-11-20
**Version**: 2.0
