# TrainBlink Frontend - Architecture Overview

Visual guide to system architecture and data flow

---

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND                                 │
│                     (React + TypeScript)                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │                 LANDING PAGE                            │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐             │    │
│  │  │  Hero    │  │  Stats   │  │ Features │             │    │
│  │  │ Section  │  │ Section  │  │ Section  │             │    │
│  │  └──────────┘  └──────────┘  └──────────┘             │    │
│  │                    ↓ API Call (every 5s)               │    │
│  └────────────────────┼─────────────────────────────────┘    │
│                       │                                         │
│  ┌────────────────────┼─────────────────────────────────┐    │
│  │              ADMIN PANEL                              │    │
│  │  ┌──────────┐     ↓ API Call (every 3s)              │    │
│  │  │          │  ┌─────────────┐  ┌─────────────┐      │    │
│  │  │ Sidebar  │  │  Dashboard  │  │  Stations   │      │    │
│  │  │   Nav    │  │   + Charts  │  │  Management │      │    │
│  │  │          │  └─────────────┘  └─────────────┘      │    │
│  │  │ - Dash   │  ┌─────────────┐  ┌─────────────┐      │    │
│  │  │ - Stats  │  │  Sessions   │  │  Analytics  │      │    │
│  │  │ - Sess   │  │  Monitor    │  │  + Graphs   │      │    │
│  │  │ - Analyt │  └─────────────┘  └─────────────┘      │    │
│  │  └──────────┘        ↓ API Call (every 2s)           │    │
│  └────────────────────┼─────────────────────────────────┘    │
│                       │                                         │
└───────────────────────┼─────────────────────────────────────────┘
                        │
                        │ HTTP REST API
                        ↓
┌─────────────────────────────────────────────────────────────────┐
│                         BACKEND                                  │
│                      (Go + Gin HTTP)                             │
│                   http://localhost:8080                          │
├─────────────────────────────────────────────────────────────────┤
│  GET  /ping                        - Health check               │
│  GET  /health                      - Detailed health            │
│  POST /api/v1/geofence/enter       - Enter station              │
│  POST /api/v1/geofence/exit        - Exit station               │
│  GET  /api/v1/geofence/stats       - Real-time stats ⭐        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────┐    ┌─────────────────┐                     │
│  │   Geofence     │───→│  Matrix Bridge  │                     │
│  │   Service      │    │    Service      │                     │
│  └────────────────┘    └─────────────────┘                     │
│         │                      │                                │
│         ↓                      ↓                                │
│  ┌────────────────┐    ┌─────────────────┐                     │
│  │  5 Stations    │    │   Matrix Rooms  │                     │
│  │  (In-Memory)   │    │   (In-Memory)   │                     │
│  └────────────────┘    └─────────────────┘                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Data Flow Architecture

### Real-Time Polling Flow

```
Frontend Component
      ↓
   useQuery Hook (TanStack Query)
      ↓
   API Service Layer
      ↓
   Axios HTTP Client
      ↓
   Backend API Endpoint
      ↓
   Geofence/Stats Service
      ↓
   In-Memory Data Store
      ↓
   Response (JSON)
      ↓
   React State Update
      ↓
   UI Re-render
```

### Component → API Flow Example

```typescript
// 1. Component
function Dashboard() {
  // 2. Custom Hook
  const { data, isLoading } = useStats(3000); // Poll every 3s
  
  // 8. Render with data
  return <StatsCard value={data?.active_sessions} />;
}

// 2. Custom Hook
function useStats(refetchInterval: number) {
  // 3. TanStack Query
  return useQuery({
    queryKey: ['stats'],
    queryFn: statsService.getStats, // 4. Service call
    refetchInterval,
  });
}

// 4. Service Layer
const statsService = {
  async getStats() {
    // 5. HTTP Request
    const response = await apiService.get('/api/v1/geofence/stats');
    return response.data; // 7. Return data
  }
};

// 5. API Client
const apiService = axios.create({
  baseURL: 'http://localhost:8080',
});
```

---

## 📂 Component Hierarchy

### Landing Page Structure

```
App.tsx (Landing)
├── Header
│   ├── Logo
│   └── Navigation
│       ├── NavLink (Home)
│       ├── NavLink (Features)
│       └── NavLink (Admin)
│
├── HeroSection
│   ├── Heading
│   ├── Description
│   └── CTAButtons
│       ├── Button (iOS)
│       └── Button (Android)
│
├── StatsSection [Real-Time: 5s]
│   ├── Heading
│   └── StatsGrid
│       ├── StatsCard (Active Users)    ← useStats()
│       ├── StatsCard (Total Stations)  ← useStats()
│       └── StatsCard (Total Sessions)  ← useStats()
│
├── FeaturesSection
│   ├── Heading
│   └── FeatureGrid
│       ├── FeatureCard (P2P)
│       ├── FeatureCard (Matrix)
│       ├── FeatureCard (MLS)
│       ├── FeatureCard (Geofence)
│       ├── FeatureCard (Ephemeral)
│       └── FeatureCard (Multi-platform)
│
├── DownloadSection
│   ├── Heading
│   └── DownloadButtons
│       ├── AppStoreButton
│       └── GooglePlayButton
│
└── Footer
    ├── Links
    └── Copyright
```

### Admin Panel Structure

```
App.tsx (Admin Router)
└── AdminLayout
    ├── Header
    │   ├── Logo
    │   └── UserMenu
    │
    ├── Sidebar
    │   └── Menu
    │       ├── MenuItem (Dashboard) ← Active
    │       ├── MenuItem (Stations)
    │       ├── MenuItem (Sessions)
    │       └── MenuItem (Analytics)
    │
    └── MainContent
        └── Dashboard [Real-Time: 3s]
            ├── PageHeader
            │
            ├── StatsCardsGrid [useStats()]
            │   ├── StatsCard (Active Sessions)
            │   ├── StatsCard (Total Stations)
            │   ├── StatsCard (Total Sessions)
            │   └── StatsCard (Matrix Rooms)
            │
            ├── ChartsGrid
            │   ├── StatsChart (Sessions Timeline)
            │   └── StatsChart (Station Activity)
            │
            └── ActiveSessionsMonitor [Real-Time: 2s]
                ├── SectionHeader
                └── Table [useStats()]
                    ├── TableHead
                    └── TableBody
                        └── SessionRows
```

---

## 🔌 API Integration Architecture

### Service Layer Pattern

```
┌─────────────────────────────────────────────┐
│          Component Layer                     │
│  (React Components using hooks)              │
└──────────────┬───────────────────────────────┘
               │ import { useStats }
               ↓
┌─────────────────────────────────────────────┐
│          Custom Hooks Layer                  │
│  useStats, useServerHealth, useStations      │
│  (TanStack Query wrapper hooks)              │
└──────────────┬───────────────────────────────┘
               │ queryFn: statsService.getStats
               ↓
┌─────────────────────────────────────────────┐
│          Service Layer                       │
│  statsService, geofenceService               │
│  (Business logic & API calls)                │
└──────────────┬───────────────────────────────┘
               │ apiService.get('/stats')
               ↓
┌─────────────────────────────────────────────┐
│          API Client Layer                    │
│  apiService (Axios instance)                 │
│  (HTTP client with interceptors)             │
└──────────────┬───────────────────────────────┘
               │ HTTP GET/POST
               ↓
┌─────────────────────────────────────────────┐
│          Backend API                         │
│  Go server at http://localhost:8080          │
└─────────────────────────────────────────────┘
```

### Type Flow

```typescript
// 1. API Response Type
interface StatsResponse {
  status: string;
  data: StatsData;
}

// 2. Service returns typed data
async getStats(): Promise<StatsResponse>

// 3. Hook returns typed data
useStats(): UseQueryResult<StatsResponse>

// 4. Component receives typed data
const { data }: { data?: StatsResponse } = useStats()

// 5. Component uses typed properties
<div>{data?.data.active_sessions}</div>
```

---

## 🎨 State Management Architecture

### TanStack Query Cache

```
┌─────────────────────────────────────────────┐
│         TanStack Query Cache                 │
├─────────────────────────────────────────────┤
│  Query Key: ['stats']                        │
│  Data: { active_sessions: 10, ... }         │
│  Status: success                             │
│  Last Updated: 2025-11-20 14:30:00          │
│  Stale Time: 2000ms                          │
│  Refetch Interval: 3000ms ← Auto refresh    │
├─────────────────────────────────────────────┤
│  Query Key: ['server-health']               │
│  Data: { status: 'healthy', ... }           │
│  Status: success                             │
│  Refetch Interval: 10000ms                  │
└─────────────────────────────────────────────┘
        ↓ Automatic Updates
┌─────────────────────────────────────────────┐
│         Components (Subscribers)             │
├─────────────────────────────────────────────┤
│  Dashboard       → useQuery(['stats'])      │
│  StatsSection    → useQuery(['stats'])      │
│  StatsCard       → Receives data via props  │
└─────────────────────────────────────────────┘
```

### No Global State Needed (Phase 1)

In Phase 1, we don't need Redux or Context for global state because:
- **TanStack Query** handles all API state
- **React Router** handles navigation state
- **Props** handle component communication
- **URL params** handle route state

---

## 📊 Real-Time Update Strategy

### Polling Configuration by Component

```typescript
// Landing Page Stats - Updates every 5 seconds
useQuery({
  queryKey: ['stats-landing'],
  queryFn: statsService.getStats,
  refetchInterval: 5000,
  staleTime: 2000,
})

// Admin Dashboard - Updates every 3 seconds
useQuery({
  queryKey: ['stats-dashboard'],
  queryFn: statsService.getStats,
  refetchInterval: 3000,
  staleTime: 1000,
})

// Active Sessions Monitor - Updates every 2 seconds
useQuery({
  queryKey: ['active-sessions'],
  queryFn: statsService.getStats,
  refetchInterval: 2000,
  staleTime: 1000,
})
```

### Smart Refetching

```typescript
// Refetch on window focus
refetchOnWindowFocus: true

// Refetch on network reconnect
refetchOnReconnect: true

// Retry failed requests
retry: 3
retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000)
```

---

## 🎯 Component Communication Patterns

### Parent → Child (Props)

```typescript
// Parent
<StatsCard
  title="Active Users"
  value={stats?.active_sessions || 0}
  icon="👥"
/>

// Child
interface StatsCardProps {
  title: string;
  value: number;
  icon: string;
}
```

### Child → Parent (Callbacks)

```typescript
// Parent
<StationCard
  station={station}
  onClick={(id) => handleStationClick(id)}
/>

// Child
interface StationCardProps {
  station: Station;
  onClick: (id: string) => void;
}
```

### Sibling Communication (Shared State)

```typescript
// Shared hook in both siblings
function ComponentA() {
  const { data } = useStats();
  return <div>{data?.active_sessions}</div>;
}

function ComponentB() {
  const { data } = useStats(); // Same cache key!
  return <div>{data?.total_stations}</div>;
}
```

---

## 🔄 Request/Response Flow

### Example: Loading Dashboard Stats

```
1. User navigates to /admin
   ↓
2. Dashboard component mounts
   ↓
3. useStats() hook initializes
   ↓
4. TanStack Query checks cache
   ↓
5. Cache miss → Trigger API call
   ↓
6. statsService.getStats() called
   ↓
7. apiService.get('/api/v1/geofence/stats')
   ↓
8. HTTP GET → Backend server
   ↓
9. Backend: geofenceHandler.GetStats()
   ↓
10. Backend: service.GetStats()
   ↓
11. Backend: Read from in-memory store
   ↓
12. Backend: Return JSON response
   ↓
13. Frontend: Axios receives response
   ↓
14. TanStack Query caches data
   ↓
15. React state updates
   ↓
16. Component re-renders with data
   ↓
17. StatsCards display values
   ↓
18. After 3s, auto-refetch (steps 6-17 repeat)
```

---

## 🚀 Build & Deploy Architecture

### Development Mode

```
┌─────────────────────────────────────────┐
│  Developer's Machine                     │
├─────────────────────────────────────────┤
│  Terminal 1: Go backend                 │
│  $ go run cmd/server/main_geofence.go   │
│  Running at: http://localhost:8080      │
├─────────────────────────────────────────┤
│  Terminal 2: Vite dev server            │
│  $ npm run dev                          │
│  Running at: http://localhost:3000      │
│  With HMR (Hot Module Replacement)      │
└─────────────────────────────────────────┘
```

### Production Build

```
┌─────────────────────────────────────────┐
│  Build Process                           │
├─────────────────────────────────────────┤
│  1. npm run build                       │
│  2. TypeScript compilation              │
│  3. Vite bundling                       │
│  4. Tailwind CSS processing             │
│  5. Asset optimization                  │
│  6. Code splitting                      │
│  7. Output → dist/ directory            │
└─────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│  Deployment                              │
├─────────────────────────────────────────┤
│  Option 1: Vercel/Netlify               │
│  - Automatic builds from git            │
│  - CDN distribution                     │
│  - SSL certificate                      │
│                                          │
│  Option 2: Static hosting (Nginx)       │
│  - Upload dist/ to server               │
│  - Configure reverse proxy              │
│  - Serve static files                   │
└─────────────────────────────────────────┘
```

---

## 🧩 Module Dependencies

### Dependency Graph

```
main.tsx
  ├── App.tsx
  │   ├── react-router-dom (routing)
  │   └── @tanstack/react-query (data fetching)
  │
  ├── components/
  │   ├── common/
  │   │   └── @base-ui-components/react
  │   ├── features/
  │   │   └── recharts (charts)
  │   └── layout/
  │
  ├── services/
  │   └── axios (HTTP client)
  │
  ├── hooks/
  │   └── @tanstack/react-query
  │
  └── styles/
      └── tailwindcss
```

### Bundle Size Optimization

```typescript
// vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'query-vendor': ['@tanstack/react-query'],
          'chart-vendor': ['recharts'],
        },
      },
    },
  },
});
```

Result:
- Main bundle: ~150KB
- React vendor: ~120KB
- Query vendor: ~50KB
- Chart vendor: ~180KB
- **Total: ~500KB** (gzipped: ~150KB)

---

## 🎯 Key Architectural Decisions

### Why TanStack Query?
- **Automatic caching** - No manual cache management
- **Background refetching** - Real-time updates made easy
- **Retry logic** - Built-in error recovery
- **Devtools** - Debug queries visually

### Why Base UI?
- **Headless components** - Full styling control
- **Accessibility** - ARIA support built-in
- **Lightweight** - Smaller bundle size
- **Tailwind integration** - Perfect match

### Why Vite?
- **Fast HMR** - Instant updates during development
- **ES modules** - Modern build system
- **Plugin ecosystem** - Easy extensibility
- **TypeScript support** - First-class support

### Why No Redux? (Phase 1)
- **TanStack Query handles API state** - 90% of state
- **React Router handles navigation** - URL state
- **Props sufficient for UI state** - Component communication
- **Simpler architecture** - Less boilerplate

---

## 📚 Related Documentation

- [Main Implementation Plan](./FRONTEND_IMPLEMENTATION_PLAN.md)
- [Quick Start Guide](./QUICK_START_GUIDE.md)
- [Component Reference](./COMPONENT_REFERENCE.md)
- [Backend API Docs](../GEOFENCE_HYBRID_IMPLEMENTATION.md)

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-20
**Status**: ✅ Complete

