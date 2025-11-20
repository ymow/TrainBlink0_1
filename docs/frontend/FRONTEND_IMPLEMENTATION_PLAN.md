# TrainBlink Frontend Implementation Plan

**Version**: 1.0.0
**Last Updated**: 2025-11-20
**Status**: Ready for Implementation

---

## 📋 Executive Summary

This document provides a comprehensive, production-ready plan for implementing a modern web frontend for TrainBlink using Base UI and React. The frontend consists of two main applications:

1. **Landing Page** - Modern product website with real-time statistics
2. **Admin Panel** - Full-featured dashboard for system monitoring and management

**Key Features**:
- Real data integration with existing Go backend (NO MOCKS)
- Modern, responsive design using Base UI components
- Real-time data updates via HTTP polling
- TypeScript for type safety
- Production-ready architecture

---

## 🎯 Project Overview

### Backend Status (Already Running)

**Server**: `http://localhost:8080`

**Available Endpoints**:
```
GET  /ping                      - Health check
GET  /health                    - Detailed health status
POST /api/v1/geofence/enter     - Enter station (Hybrid P2P + Matrix)
POST /api/v1/geofence/exit      - Exit station
GET  /api/v1/geofence/stats     - Real-time statistics
```

**Live Data**:
- 5 Stations: Tokyo, Taipei, Shibuya, Taichung, Kaohsiung
- Real-time session tracking
- Matrix Bridge integration
- P2P + Matrix dual-channel support

### Technical Stack

**Core**:
- React 18.3+ with TypeScript
- Vite 5.x (build tool)
- Base UI (@base-ui-components/react)
- React Router 6.x (navigation)

**State Management & API**:
- TanStack Query v5 (API state management)
- Axios (HTTP client)

**Styling**:
- Tailwind CSS 3.x
- Base UI Tailwind integration
- CSS Modules (for component-specific styles)

**Data Visualization**:
- Recharts 2.x (primary choice)
- Alternative: Chart.js with react-chartjs-2

**Development Tools**:
- TypeScript 5.x
- ESLint + Prettier
- Vite HMR (Hot Module Replacement)

---

## 📁 Project Structure

```
trainblink-frontend/
├── public/
│   ├── favicon.ico
│   ├── logo.svg
│   └── images/
│       ├── hero-bg.jpg
│       ├── features/
│       └── screenshots/
│
├── src/
│   ├── main.tsx                 # Entry point
│   ├── App.tsx                  # Root component with routing
│   ├── vite-env.d.ts           # Vite type definitions
│   │
│   ├── config/
│   │   ├── api.config.ts        # API base URL and endpoints
│   │   └── constants.ts         # App-wide constants
│   │
│   ├── types/
│   │   ├── api.types.ts         # API request/response types
│   │   ├── station.types.ts     # Station-related types
│   │   ├── session.types.ts     # Session-related types
│   │   └── index.ts             # Export all types
│   │
│   ├── services/
│   │   ├── api.service.ts       # Base API client (Axios instance)
│   │   ├── geofence.service.ts  # Geofence API calls
│   │   ├── stats.service.ts     # Statistics API calls
│   │   └── index.ts             # Export all services
│   │
│   ├── hooks/
│   │   ├── useStats.ts          # Real-time stats polling
│   │   ├── useStations.ts       # Station data management
│   │   ├── useSessions.ts       # Session tracking
│   │   └── index.ts             # Export all hooks
│   │
│   ├── components/
│   │   ├── common/              # Shared components
│   │   │   ├── Button/
│   │   │   │   ├── Button.tsx
│   │   │   │   └── Button.module.css
│   │   │   ├── Card/
│   │   │   ├── Badge/
│   │   │   ├── Modal/
│   │   │   ├── Loading/
│   │   │   └── ErrorBoundary/
│   │   │
│   │   ├── layout/              # Layout components
│   │   │   ├── Header/
│   │   │   ├── Footer/
│   │   │   ├── Sidebar/
│   │   │   └── AdminLayout/
│   │   │
│   │   └── features/            # Feature-specific components
│   │       ├── stats/
│   │       │   ├── StatsCard.tsx
│   │       │   ├── StatsChart.tsx
│   │       │   └── RealTimeCounter.tsx
│   │       ├── stations/
│   │       │   ├── StationCard.tsx
│   │       │   ├── StationList.tsx
│   │       │   ├── StationMap.tsx
│   │       │   └── StationDetail.tsx
│   │       └── sessions/
│   │           ├── SessionTable.tsx
│   │           ├── SessionDetail.tsx
│   │           └── ActiveSessionsMonitor.tsx
│   │
│   ├── pages/
│   │   ├── landing/             # Landing page sections
│   │   │   ├── LandingPage.tsx  # Main landing page
│   │   │   ├── HeroSection.tsx
│   │   │   ├── FeaturesSection.tsx
│   │   │   ├── StatsSection.tsx
│   │   │   ├── DownloadSection.tsx
│   │   │   └── styles.css
│   │   │
│   │   └── admin/               # Admin panel pages
│   │       ├── Dashboard.tsx    # Main dashboard
│   │       ├── Stations.tsx     # Station management
│   │       ├── Sessions.tsx     # Active sessions monitoring
│   │       ├── Analytics.tsx    # Data analytics
│   │       └── Settings.tsx     # System settings (Phase 2)
│   │
│   ├── utils/
│   │   ├── formatters.ts        # Date, number formatting
│   │   ├── validators.ts        # Input validation
│   │   └── helpers.ts           # Helper functions
│   │
│   └── styles/
│       ├── index.css            # Global styles + Tailwind imports
│       ├── variables.css        # CSS custom properties
│       └── base-ui-overrides.css # Base UI component overrides
│
├── .env.development             # Development environment variables
├── .env.production              # Production environment variables
├── .eslintrc.cjs               # ESLint configuration
├── .prettierrc                 # Prettier configuration
├── tailwind.config.js          # Tailwind CSS configuration
├── tsconfig.json               # TypeScript configuration
├── tsconfig.node.json          # TypeScript config for Node
├── vite.config.ts              # Vite configuration
├── package.json
└── README.md
```

---

## 🎨 Design System with Base UI

### Core Base UI Components to Use

**Navigation & Layout**:
- `Menu` - Navigation menus (header, sidebar)
- `Tabs` - Content organization in admin panel
- `Breadcrumb` - Navigation breadcrumbs

**Data Display**:
- `Table` - Session tables, station lists
- `Badge` - Status indicators, counts
- `Tooltip` - Additional information on hover

**Feedback**:
- `Alert` - Success/error messages
- `Progress` - Loading states, data fetching
- `Skeleton` - Loading placeholders

**Inputs** (Phase 2):
- `Input` - Form fields
- `Select` - Dropdowns
- `Switch` - Toggle settings
- `Checkbox` - Multi-select options

**Overlay**:
- `Dialog` - Modals for confirmations, details
- `Popover` - Contextual information

### Tailwind Configuration

```javascript
// tailwind.config.js
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
        success: {
          500: '#10b981',
          600: '#059669',
        },
        warning: {
          500: '#f59e0b',
          600: '#d97706',
        },
        error: {
          500: '#ef4444',
          600: '#dc2626',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['Fira Code', 'monospace'],
      },
    },
  },
  plugins: [],
}
```

---

## 🌐 Landing Page Design

### Page Structure

```
┌─────────────────────────────────────────────┐
│              HEADER / NAV                    │
│  Logo    Home  Features  Download   Admin   │
└─────────────────────────────────────────────┘
│                                              │
│              HERO SECTION                    │
│   TrainBlink - Connect at Train Stations    │
│        [Download iOS]  [Download Android]    │
│                                              │
├─────────────────────────────────────────────┤
│              STATS SECTION                   │
│  [Active Users] [Total Stations] [Sessions] │
│           REAL-TIME DATA FROM API            │
├─────────────────────────────────────────────┤
│            FEATURES SECTION                  │
│   ┌──────┐  ┌──────┐  ┌──────┐             │
│   │ P2P  │  │Matrix│  │ MLS  │             │
│   └──────┘  └──────┘  └──────┘             │
├─────────────────────────────────────────────┤
│         HOW IT WORKS SECTION                │
│   1. Enter Station  →  2. Connect           │
│   3. Chat Safely    →  4. Leave             │
├─────────────────────────────────────────────┤
│           DOWNLOAD SECTION                  │
│     App Store       Google Play             │
│     [QR Code]       [QR Code]               │
├─────────────────────────────────────────────┤
│              FOOTER                         │
│   About | Privacy | Terms | Contact         │
└─────────────────────────────────────────────┘
```

### Component Breakdown

#### 1. Hero Section
```typescript
// src/pages/landing/HeroSection.tsx
import { Button } from '@base-ui-components/react/button';

interface HeroSectionProps {}

export function HeroSection() {
  return (
    <section className="hero-section bg-gradient-to-br from-blue-600 to-blue-800 text-white py-20">
      <div className="container mx-auto px-4 text-center">
        <h1 className="text-5xl font-bold mb-6">
          Connect with People at Train Stations
        </h1>
        <p className="text-xl mb-8 max-w-2xl mx-auto">
          TrainBlink uses P2P and Matrix protocols for secure, 
          ephemeral messaging when you're near train stations.
        </p>
        <div className="flex gap-4 justify-center">
          <Button 
            className="bg-white text-blue-600 px-8 py-4 rounded-lg text-lg font-semibold hover:bg-gray-100"
          >
            Download for iOS
          </Button>
          <Button 
            className="bg-blue-500 text-white px-8 py-4 rounded-lg text-lg font-semibold hover:bg-blue-400"
          >
            Download for Android
          </Button>
        </div>
      </div>
    </section>
  );
}
```

#### 2. Real-Time Stats Section
```typescript
// src/pages/landing/StatsSection.tsx
import { useQuery } from '@tanstack/react-query';
import { statsService } from '@/services';
import { StatsCard } from '@/components/features/stats';

export function StatsSection() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['stats'],
    queryFn: statsService.getStats,
    refetchInterval: 5000, // Poll every 5 seconds
  });

  if (isLoading) {
    return <div>Loading stats...</div>;
  }

  return (
    <section className="stats-section py-16 bg-gray-50">
      <div className="container mx-auto px-4">
        <h2 className="text-3xl font-bold text-center mb-12">
          Live Platform Statistics
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <StatsCard
            title="Active Users"
            value={stats?.active_sessions || 0}
            icon="👥"
            trend="+12%"
          />
          <StatsCard
            title="Total Stations"
            value={stats?.total_stations || 0}
            icon="🚉"
          />
          <StatsCard
            title="Total Sessions"
            value={stats?.total_sessions || 0}
            icon="💬"
            trend="+8%"
          />
        </div>
      </div>
    </section>
  );
}
```

#### 3. Features Section
```typescript
// src/pages/landing/FeaturesSection.tsx
import { Badge } from '@base-ui-components/react/badge';

const features = [
  {
    title: 'P2P Connection',
    description: 'Direct peer-to-peer messaging with WebRTC',
    icon: '🔗',
    badge: 'Core',
  },
  {
    title: 'Matrix Protocol',
    description: 'Decentralized messaging with Matrix bridge',
    icon: '🌐',
    badge: 'Hybrid',
  },
  {
    title: 'MLS Encryption',
    description: 'End-to-end encryption with Message Layer Security',
    icon: '🔒',
    badge: 'Secure',
  },
  {
    title: 'Geofencing',
    description: 'Automatic station detection via GPS',
    icon: '📍',
    badge: 'Smart',
  },
  {
    title: 'Ephemeral Messages',
    description: 'Messages auto-delete after leaving station',
    icon: '⏱️',
    badge: 'Privacy',
  },
  {
    title: 'Multi-Platform',
    description: 'iOS and Android support',
    icon: '📱',
    badge: 'Cross-platform',
  },
];

export function FeaturesSection() {
  return (
    <section className="features-section py-16">
      <div className="container mx-auto px-4">
        <h2 className="text-3xl font-bold text-center mb-12">
          Powerful Features
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature) => (
            <div
              key={feature.title}
              className="feature-card bg-white p-6 rounded-lg shadow-md hover:shadow-lg transition"
            >
              <div className="flex items-start justify-between mb-4">
                <span className="text-4xl">{feature.icon}</span>
                <Badge className="bg-blue-100 text-blue-800">
                  {feature.badge}
                </Badge>
              </div>
              <h3 className="text-xl font-semibold mb-2">{feature.title}</h3>
              <p className="text-gray-600">{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
```

---

## 🎛️ Admin Panel Design

### Layout Structure

```
┌────────────────────────────────────────────────────────┐
│  HEADER                                    [User] [⚙️] │
│  TrainBlink Admin                                      │
├───────┬────────────────────────────────────────────────┤
│       │  DASHBOARD                                     │
│  📊   │  ┌─────────┐ ┌─────────┐ ┌─────────┐         │
│  📈   │  │ Active  │ │ Total   │ │ Sessions│         │
│  🚉   │  │ Users   │ │Stations │ │ Today   │         │
│  💬   │  └─────────┘ └─────────┘ └─────────┘         │
│  ⚙️   │  ┌──────────────────────────────────┐         │
│       │  │   Sessions Over Time (Chart)     │         │
│       │  └──────────────────────────────────┘         │
│       │  ┌──────────────────────────────────┐         │
│       │  │   Active Sessions Table          │         │
│       │  └──────────────────────────────────┘         │
└───────┴────────────────────────────────────────────────┘
```

### Admin Panel Pages

#### 1. Dashboard Page
```typescript
// src/pages/admin/Dashboard.tsx
import { useQuery } from '@tanstack/react-query';
import { statsService } from '@/services';
import { StatsCard, StatsChart, ActiveSessionsMonitor } from '@/components/features';

export function Dashboard() {
  const { data: stats, isLoading, error } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: statsService.getStats,
    refetchInterval: 3000, // Refresh every 3 seconds
  });

  if (error) {
    return (
      <Alert variant="error">
        Failed to load statistics. Please check if the server is running.
      </Alert>
    );
  }

  return (
    <div className="dashboard p-6">
      <h1 className="text-3xl font-bold mb-8">Dashboard</h1>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <StatsCard
          title="Active Sessions"
          value={stats?.active_sessions || 0}
          icon="👥"
          color="blue"
          isLoading={isLoading}
        />
        <StatsCard
          title="Total Stations"
          value={stats?.total_stations || 0}
          icon="🚉"
          color="green"
          isLoading={isLoading}
        />
        <StatsCard
          title="Total Sessions"
          value={stats?.total_sessions || 0}
          icon="💬"
          color="purple"
          isLoading={isLoading}
        />
        <StatsCard
          title="Matrix Rooms"
          value={stats?.matrix_stats?.total_rooms || 0}
          icon="🌐"
          color="orange"
          isLoading={isLoading}
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <StatsChart
          title="Sessions Over Time"
          data={stats?.timeline_data}
          type="line"
        />
        <StatsChart
          title="Stations by Activity"
          data={stats?.station_activity}
          type="bar"
        />
      </div>

      {/* Active Sessions Monitor */}
      <ActiveSessionsMonitor />
    </div>
  );
}
```

#### 2. Stations Page
```typescript
// src/pages/admin/Stations.tsx
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Table } from '@base-ui-components/react/table';
import { Badge } from '@base-ui-components/react/badge';
import { StationCard, StationDetail } from '@/components/features/stations';

export function Stations() {
  const [selectedStation, setSelectedStation] = useState<string | null>(null);

  const { data: stats } = useQuery({
    queryKey: ['stations-data'],
    queryFn: statsService.getStats,
    refetchInterval: 5000,
  });

  // In Phase 1, station list comes from stats API
  // In Phase 2, we'll add dedicated stations endpoint
  const stations = [
    { id: 'station_tokyo_001', name: 'Tokyo Station', city: 'Tokyo' },
    { id: 'station_taipei_001', name: 'Taipei Main Station', city: 'Taipei' },
    { id: 'station_shibuya_001', name: 'Shibuya Station', city: 'Tokyo' },
    { id: 'station_taichung_001', name: 'Taichung Station', city: 'Taichung' },
    { id: 'station_kaohsiung_001', name: 'Kaohsiung Station', city: 'Kaohsiung' },
  ];

  return (
    <div className="stations-page p-6">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Stations</h1>
        <Badge className="bg-green-100 text-green-800">
          {stats?.total_stations || 0} Active
        </Badge>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {stations.map((station) => (
          <StationCard
            key={station.id}
            station={station}
            onClick={() => setSelectedStation(station.id)}
          />
        ))}
      </div>

      {/* Station Detail Modal */}
      {selectedStation && (
        <StationDetail
          stationId={selectedStation}
          onClose={() => setSelectedStation(null)}
        />
      )}
    </div>
  );
}
```

#### 3. Sessions Monitoring Page
```typescript
// src/pages/admin/Sessions.tsx
import { useQuery } from '@tanstack/react-query';
import { Table } from '@base-ui-components/react/table';
import { Badge } from '@base-ui-components/react/badge';
import { formatDistanceToNow } from 'date-fns';

export function Sessions() {
  const { data: stats } = useQuery({
    queryKey: ['sessions'],
    queryFn: statsService.getStats,
    refetchInterval: 2000, // Refresh every 2 seconds for real-time
  });

  // Note: In Phase 1, we work with aggregated data
  // In Phase 2, we'll add dedicated sessions endpoint

  const columns = [
    { key: 'id', label: 'Session ID' },
    { key: 'user', label: 'User' },
    { key: 'station', label: 'Station' },
    { key: 'duration', label: 'Duration' },
    { key: 'type', label: 'Type' },
    { key: 'status', label: 'Status' },
  ];

  return (
    <div className="sessions-page p-6">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Active Sessions</h1>
        <div className="flex gap-4">
          <Badge className="bg-blue-100 text-blue-800">
            Active: {stats?.active_sessions || 0}
          </Badge>
          <Badge className="bg-gray-100 text-gray-800">
            Total: {stats?.total_sessions || 0}
          </Badge>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow">
        <div className="p-4 border-b">
          <input
            type="text"
            placeholder="Search sessions..."
            className="w-full px-4 py-2 border rounded-lg"
          />
        </div>

        <Table>
          <Table.Head>
            <Table.Row>
              {columns.map((col) => (
                <Table.ColumnHeader key={col.key}>
                  {col.label}
                </Table.ColumnHeader>
              ))}
            </Table.Row>
          </Table.Head>
          <Table.Body>
            {/* Session rows will be populated here */}
            <Table.Row>
              <Table.Cell colSpan={6} className="text-center text-gray-500 py-8">
                Real-time session data will appear here
              </Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table>
      </div>
    </div>
  );
}
```

### Admin Layout Component

```typescript
// src/components/layout/AdminLayout/AdminLayout.tsx
import { Outlet, NavLink } from 'react-router-dom';
import { Menu } from '@base-ui-components/react/menu';

const menuItems = [
  { path: '/admin', label: 'Dashboard', icon: '📊' },
  { path: '/admin/stations', label: 'Stations', icon: '🚉' },
  { path: '/admin/sessions', label: 'Sessions', icon: '💬' },
  { path: '/admin/analytics', label: 'Analytics', icon: '📈' },
];

export function AdminLayout() {
  return (
    <div className="admin-layout min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="px-6 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-blue-600">
            TrainBlink Admin
          </h1>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">Admin User</span>
            <button className="p-2 hover:bg-gray-100 rounded-lg">⚙️</button>
          </div>
        </div>
      </header>

      <div className="flex">
        {/* Sidebar */}
        <aside className="w-64 bg-white border-r min-h-[calc(100vh-73px)]">
          <nav className="p-4">
            <Menu>
              {menuItems.map((item) => (
                <Menu.Item key={item.path}>
                  <NavLink
                    to={item.path}
                    className={({ isActive }) =>
                      `flex items-center gap-3 px-4 py-3 rounded-lg transition ${
                        isActive
                          ? 'bg-blue-50 text-blue-600 font-semibold'
                          : 'text-gray-700 hover:bg-gray-50'
                      }`
                    }
                  >
                    <span className="text-xl">{item.icon}</span>
                    {item.label}
                  </NavLink>
                </Menu.Item>
              ))}
            </Menu>
          </nav>
        </aside>

        {/* Main Content */}
        <main className="flex-1">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
```

---

## 🔌 API Integration Strategy

### Service Layer Architecture

#### Base API Client

```typescript
// src/services/api.service.ts
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { API_CONFIG } from '@/config/api.config';

class ApiService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_CONFIG.BASE_URL,
      timeout: API_CONFIG.TIMEOUT,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        // Add any auth headers here (Phase 2)
        console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`);
        return config;
      },
      (error) => {
        console.error('[API] Request error:', error);
        return Promise.reject(error);
      }
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => {
        console.log(`[API] Response from ${response.config.url}:`, response.status);
        return response;
      },
      (error) => {
        console.error('[API] Response error:', error.response?.status, error.message);
        
        // Handle common errors
        if (error.response?.status === 404) {
          console.error('[API] Resource not found');
        } else if (error.response?.status === 500) {
          console.error('[API] Server error');
        } else if (!error.response) {
          console.error('[API] Network error - is the server running?');
        }
        
        return Promise.reject(error);
      }
    );
  }

  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }
}

export const apiService = new ApiService();
```

#### Stats Service

```typescript
// src/services/stats.service.ts
import { apiService } from './api.service';
import { StatsResponse } from '@/types';

class StatsService {
  /**
   * Get real-time statistics
   * Polls: Every 3-5 seconds
   */
  async getStats(): Promise<StatsResponse> {
    return apiService.get<StatsResponse>('/api/v1/geofence/stats');
  }

  /**
   * Get health check
   */
  async getHealth(): Promise<HealthResponse> {
    return apiService.get<HealthResponse>('/health');
  }

  /**
   * Ping test
   */
  async ping(): Promise<PingResponse> {
    return apiService.get<PingResponse>('/ping');
  }
}

export const statsService = new StatsService();
```

#### Geofence Service

```typescript
// src/services/geofence.service.ts
import { apiService } from './api.service';
import {
  EnterStationRequest,
  EnterStationResponse,
  ExitStationRequest,
  ExitStationResponse,
} from '@/types';

class GeofenceService {
  /**
   * Enter a station
   */
  async enterStation(
    request: EnterStationRequest,
    userId: string,
    deviceId: string
  ): Promise<EnterStationResponse> {
    return apiService.post<EnterStationResponse>(
      '/api/v1/geofence/enter',
      request,
      {
        headers: {
          'X-User-ID': userId,
          'X-Device-ID': deviceId,
        },
      }
    );
  }

  /**
   * Exit a station
   */
  async exitStation(
    request: ExitStationRequest,
    userId: string
  ): Promise<ExitStationResponse> {
    return apiService.post<ExitStationResponse>(
      '/api/v1/geofence/exit',
      request,
      {
        headers: {
          'X-User-ID': userId,
        },
      }
    );
  }
}

export const geofenceService = new GeofenceService();
```

### Custom Hooks with TanStack Query

#### useStats Hook

```typescript
// src/hooks/useStats.ts
import { useQuery, UseQueryOptions } from '@tanstack/react-query';
import { statsService } from '@/services';
import { StatsResponse } from '@/types';

interface UseStatsOptions {
  refetchInterval?: number;
  enabled?: boolean;
}

export function useStats(options?: UseStatsOptions) {
  return useQuery<StatsResponse, Error>({
    queryKey: ['stats'],
    queryFn: () => statsService.getStats(),
    refetchInterval: options?.refetchInterval ?? 5000, // Default: 5s
    enabled: options?.enabled ?? true,
    staleTime: 2000, // Consider data stale after 2s
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });
}

// Usage example:
// const { data: stats, isLoading, error } = useStats({ refetchInterval: 3000 });
```

#### useServerHealth Hook

```typescript
// src/hooks/useServerHealth.ts
import { useQuery } from '@tanstack/react-query';
import { statsService } from '@/services';

export function useServerHealth() {
  return useQuery({
    queryKey: ['server-health'],
    queryFn: () => statsService.getHealth(),
    refetchInterval: 10000, // Check every 10 seconds
    retry: 1,
  });
}
```

### Type Definitions

```typescript
// src/types/api.types.ts

export interface StatsResponse {
  status: string;
  data: {
    total_stations: number;
    total_sessions: number;
    active_sessions: number;
    matrix_stats: MatrixStats;
  };
}

export interface MatrixStats {
  total_rooms: number;
  active_rooms: number;
  total_users: number;
  active_users: number;
}

export interface HealthResponse {
  status: string;
  version: string;
  timestamp: string;
  features: string[];
}

export interface PingResponse {
  message: string;
  timestamp: string;
  server: string;
}

// Geofence types
export interface Coordinates {
  latitude: number;
  longitude: number;
  accuracy: number;
}

export interface Capabilities {
  p2p_enabled: boolean;
  matrix_enabled: boolean;
  mls_supported: boolean;
}

export interface EnterStationRequest {
  station_id: string;
  coordinates: Coordinates;
  timestamp?: string;
  client_version?: string;
  capabilities: Capabilities;
}

export interface EnterStationResponse {
  session_id: string;
  station: Station;
  p2p?: P2PResources;
  matrix?: MatrixResources;
  content_available: number;
  recommendations?: Recommendations;
}

export interface ExitStationRequest {
  session_id: string;
  station_id: string;
  timestamp?: string;
  duration_seconds: number;
  activity: ActivitySummary;
}

export interface ExitStationResponse {
  session_summary: ActivitySummary;
  cleanup_status: CleanupStatus;
}

export interface Station {
  id: string;
  place_id: string;
  name: string;
  name_en: string;
  lat: number;
  lng: number;
  type: string;
  radius: number;
  address: string;
  city: string;
  country: string;
  lines: string[];
  created_at: string;
  updated_at: string;
}

export interface P2PResources {
  active_peers_nearby: number;
  signaling_server: string;
  ice_servers: ICEServer[];
}

export interface ICEServer {
  urls: string[];
  username?: string;
  credential?: string;
}

export interface MatrixResources {
  room_id: string;
  room_alias: string;
  matrix_user_id: string;
  homeserver_url: string;
  access_token: string;
  encryption_enabled: boolean;
  mls_group_id?: string;
  member_count: number;
}

export interface ActivitySummary {
  p2p_chats_created: number;
  p2p_messages_sent: number;
  content_shared: number;
  matrix_messages_sent: number;
}

export interface CleanupStatus {
  p2p_closed: boolean;
  matrix_left: boolean;
  local_data_cleared: boolean;
}

export interface Recommendations {
  peers: string[];
  icebreaker_cards: string[];
}
```

---

## 📊 Real-Time Data Polling Strategy

### Polling Configuration

```typescript
// src/config/api.config.ts

export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  TIMEOUT: 10000, // 10 seconds

  // Polling intervals (in milliseconds)
  POLLING: {
    STATS_LANDING: 5000,      // Landing page stats: every 5s
    STATS_DASHBOARD: 3000,    // Admin dashboard: every 3s
    ACTIVE_SESSIONS: 2000,    // Active sessions monitor: every 2s
    HEALTH_CHECK: 10000,      // Server health: every 10s
  },

  // Retry configuration
  RETRY: {
    MAX_ATTEMPTS: 3,
    INITIAL_DELAY: 1000,      // 1 second
    MAX_DELAY: 30000,         // 30 seconds
  },
};

export const ENDPOINTS = {
  PING: '/ping',
  HEALTH: '/health',
  GEOFENCE_ENTER: '/api/v1/geofence/enter',
  GEOFENCE_EXIT: '/api/v1/geofence/exit',
  GEOFENCE_STATS: '/api/v1/geofence/stats',
};
```

### TanStack Query Configuration

```typescript
// src/main.tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import App from './App';
import './styles/index.css';

// Configure QueryClient
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 2000,           // Data is fresh for 2 seconds
      gcTime: 5 * 60 * 1000,     // Cache for 5 minutes (formerly cacheTime)
      refetchOnWindowFocus: true, // Refetch when window regains focus
      refetchOnReconnect: true,  // Refetch when reconnecting
      retry: 3,                  // Retry failed requests 3 times
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
      {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
    </QueryClientProvider>
  </React.StrictMode>
);
```

### Error Handling

```typescript
// src/components/common/ErrorBoundary/ErrorBoundary.tsx
import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Alert } from '@base-ui-components/react/alert';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error:', error, errorInfo);
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
          <div className="max-w-md w-full">
            <Alert variant="error">
              <Alert.Title>Something went wrong</Alert.Title>
              <Alert.Description>
                {this.state.error?.message || 'An unexpected error occurred'}
              </Alert.Description>
              <button
                onClick={() => window.location.reload()}
                className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
              >
                Reload Page
              </button>
            </Alert>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
```

---

## 🚀 Implementation Phases

### Phase 1: Foundation & Landing Page (Week 1)

**Duration**: 5-7 days
**Priority**: P0 (Must Have)

**Tasks**:
1. **Project Setup** (Day 1)
   - Initialize Vite + React + TypeScript project
   - Install dependencies (Base UI, TanStack Query, Tailwind, etc.)
   - Configure Tailwind CSS
   - Setup project structure
   - Configure ESLint + Prettier
   - Create environment files

2. **Type Definitions & Services** (Day 2)
   - Define all TypeScript interfaces
   - Implement base API service
   - Implement stats service
   - Implement geofence service
   - Create custom hooks (useStats, useServerHealth)
   - Test API integration with backend

3. **Landing Page - Core** (Day 3-4)
   - Create page structure
   - Implement Header/Navigation
   - Implement Hero Section
   - Implement Real-Time Stats Section
   - Implement Features Section
   - Test real-time data updates

4. **Landing Page - Complete** (Day 5)
   - Implement How It Works section
   - Implement Download section
   - Implement Footer
   - Responsive design testing
   - Cross-browser testing

5. **Polish & Documentation** (Day 6-7)
   - Add loading states
   - Add error handling
   - Performance optimization
   - Write component documentation
   - Create README with setup instructions

**Deliverables**:
- ✅ Fully functional landing page
- ✅ Real-time stats integration
- ✅ Responsive design (mobile, tablet, desktop)
- ✅ Type-safe API integration
- ✅ Error handling & loading states

**Success Criteria**:
- Landing page loads in < 2 seconds
- Stats update every 5 seconds without flickering
- Works on Chrome, Firefox, Safari
- Mobile responsive (iPhone, Android)
- Zero TypeScript errors

---

### Phase 2: Admin Panel - Dashboard (Week 2)

**Duration**: 5-7 days
**Priority**: P0 (Must Have)

**Tasks**:
1. **Admin Layout** (Day 1)
   - Create AdminLayout component
   - Implement Sidebar navigation
   - Implement Header
   - Setup routing (React Router)
   - Create dashboard page structure

2. **Dashboard - Stats Cards** (Day 2)
   - Implement StatsCard component
   - Create 4 main stat cards
   - Add real-time updates
   - Add animations/transitions
   - Test with real data

3. **Dashboard - Charts** (Day 3-4)
   - Install and configure Recharts
   - Create StatsChart component
   - Implement line chart (sessions over time)
   - Implement bar chart (stations by activity)
   - Add chart tooltips and legends
   - Mock time-series data (will be real in Phase 4)

4. **Dashboard - Active Sessions** (Day 5)
   - Create ActiveSessionsMonitor component
   - Implement Base UI Table
   - Add real-time updates (2s polling)
   - Add session status badges
   - Add search/filter (basic)

5. **Testing & Polish** (Day 6-7)
   - Integration testing
   - Performance optimization
   - Add loading skeletons
   - Refine responsive design
   - Documentation

**Deliverables**:
- ✅ Admin dashboard with real-time stats
- ✅ Data visualization charts
- ✅ Active sessions monitoring
- ✅ Navigation and routing

**Success Criteria**:
- Dashboard updates every 3 seconds
- Charts render correctly with data
- Table shows session information
- Navigation works smoothly
- No performance issues

---

### Phase 3: Admin Panel - Stations & Sessions (Week 3)

**Duration**: 5-7 days
**Priority**: P1 (Should Have)

**Tasks**:
1. **Stations Page** (Day 1-2)
   - Create Stations page layout
   - Implement StationCard component
   - Display 5 stations in grid
   - Add station details modal
   - Show active users per station

2. **Station Detail View** (Day 3)
   - Create StationDetail component
   - Show station information
   - Show active sessions at station
   - Show Matrix room info
   - Add close/back navigation

3. **Sessions Page** (Day 4-5)
   - Create Sessions page layout
   - Implement sessions table with Base UI
   - Add session detail view
   - Implement search functionality
   - Add status filters (active/ended)

4. **Matrix Room Status** (Day 6)
   - Display Matrix room information
   - Show room member count
   - Show room activity
   - Add room details modal

5. **Testing & Polish** (Day 7)
   - Integration testing
   - Fix bugs
   - Performance optimization
   - Documentation

**Deliverables**:
- ✅ Complete stations management interface
- ✅ Sessions monitoring with details
- ✅ Matrix room status display
- ✅ Search and filter functionality

**Success Criteria**:
- All 5 stations displayed correctly
- Station details show real data
- Sessions table updates in real-time
- Search/filter works properly

---

### Phase 4: Analytics & Data Visualization (Week 4)

**Duration**: 5-7 days
**Priority**: P2 (Nice to Have)

**Tasks**:
1. **Analytics Page Setup** (Day 1)
   - Create Analytics page structure
   - Design layout
   - Plan charts and metrics

2. **Advanced Charts** (Day 2-3)
   - Sessions timeline chart
   - Station popularity chart
   - P2P vs Matrix usage chart
   - Peak hours heatmap

3. **Metrics & Insights** (Day 4-5)
   - Calculate average session duration
   - Calculate peak usage times
   - Show trends (day/week/month)
   - Add comparison views

4. **Export Functionality** (Day 6)
   - Export data to CSV
   - Export charts as images
   - Print-friendly view

5. **Testing & Polish** (Day 7)
   - Integration testing
   - Performance optimization
   - Documentation

**Deliverables**:
- ✅ Comprehensive analytics page
- ✅ Multiple visualization types
- ✅ Data export functionality
- ✅ Insights and trends

---

### Phase 5: Polish & Optimization (Week 5)

**Duration**: 3-5 days
**Priority**: P1 (Should Have)

**Tasks**:
1. **Performance Optimization** (Day 1-2)
   - Code splitting
   - Lazy loading
   - Image optimization
   - Bundle size optimization

2. **Accessibility** (Day 2-3)
   - ARIA labels
   - Keyboard navigation
   - Screen reader testing
   - Color contrast fixes

3. **Testing** (Day 3-4)
   - Unit tests for utilities
   - Integration tests for key flows
   - E2E tests (optional)
   - Cross-browser testing

4. **Documentation** (Day 4-5)
   - Component documentation
   - API integration guide
   - Deployment guide
   - User manual

5. **Deployment Preparation** (Day 5)
   - Production build testing
   - Environment configuration
   - Deploy to staging
   - Performance testing

**Deliverables**:
- ✅ Optimized production build
- ✅ Accessibility improvements
- ✅ Test coverage
- ✅ Complete documentation
- ✅ Deployment-ready application

---

## 📦 Installation & Setup Guide

### Prerequisites

```bash
Node.js >= 18.0.0
npm >= 9.0.0 or yarn >= 1.22.0
Git
```

### Initial Setup

```bash
# 1. Create project
npm create vite@latest trainblink-frontend -- --template react-ts
cd trainblink-frontend

# 2. Install dependencies
npm install

# 3. Install Base UI
npm install @base-ui-components/react

# 4. Install other dependencies
npm install react-router-dom @tanstack/react-query axios
npm install -D tailwindcss postcss autoprefixer
npm install recharts date-fns clsx

# 5. Initialize Tailwind CSS
npx tailwindcss init -p

# 6. Install dev dependencies
npm install -D @types/node eslint prettier eslint-config-prettier
npm install -D @tanstack/react-query-devtools

# 7. Create directory structure
mkdir -p src/{components,pages,services,hooks,types,utils,styles,config}
mkdir -p src/components/{common,layout,features}
mkdir -p src/pages/{landing,admin}
mkdir -p src/components/features/{stats,stations,sessions}
```

### Environment Configuration

```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=TrainBlink
VITE_APP_VERSION=1.0.0
VITE_ENABLE_DEV_TOOLS=true
VITE_POLLING_INTERVAL=5000

# .env.production
VITE_API_BASE_URL=https://api.trainblink.org
VITE_APP_NAME=TrainBlink
VITE_APP_VERSION=1.0.0
VITE_ENABLE_DEV_TOOLS=false
VITE_POLLING_INTERVAL=5000
```

### Package.json Scripts

```json
{
  "name": "trainblink-frontend",
  "private": true,
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "format": "prettier --write \"src/**/*.{ts,tsx,css}\"",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "@base-ui-components/react": "^0.0.1",
    "@tanstack/react-query": "^5.0.0",
    "axios": "^1.6.0",
    "clsx": "^2.0.0",
    "date-fns": "^3.0.0",
    "react": "^18.3.0",
    "react-dom": "^18.3.0",
    "react-router-dom": "^6.20.0",
    "recharts": "^2.10.0"
  },
  "devDependencies": {
    "@tanstack/react-query-devtools": "^5.0.0",
    "@types/node": "^20.10.0",
    "@types/react": "^18.3.0",
    "@types/react-dom": "^18.3.0",
    "@typescript-eslint/eslint-plugin": "^6.14.0",
    "@typescript-eslint/parser": "^6.14.0",
    "@vitejs/plugin-react": "^4.2.0",
    "autoprefixer": "^10.4.16",
    "eslint": "^8.55.0",
    "eslint-config-prettier": "^9.1.0",
    "postcss": "^8.4.32",
    "prettier": "^3.1.0",
    "tailwindcss": "^3.3.6",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

### Vite Configuration

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
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

### TypeScript Configuration

```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

---

## 🧪 Testing Strategy

### Unit Testing

```bash
# Install testing dependencies
npm install -D vitest @testing-library/react @testing-library/jest-dom
npm install -D @testing-library/user-event jsdom
```

```typescript
// src/components/common/Button/Button.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders with text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('calls onClick when clicked', async () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    const button = screen.getByText('Click me');
    await userEvent.click(button);
    
    expect(handleClick).toHaveBeenCalledTimes(1);
  });
});
```

### Integration Testing

```typescript
// src/pages/admin/Dashboard.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Dashboard } from './Dashboard';
import * as statsService from '@/services/stats.service';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
  },
});

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={queryClient}>
    {children}
  </QueryClientProvider>
);

describe('Dashboard', () => {
  it('displays loading state initially', () => {
    render(<Dashboard />, { wrapper });
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('displays stats after loading', async () => {
    vi.spyOn(statsService, 'getStats').mockResolvedValue({
      status: 'success',
      data: {
        total_stations: 5,
        total_sessions: 100,
        active_sessions: 10,
        matrix_stats: {
          total_rooms: 5,
          active_rooms: 3,
          total_users: 20,
          active_users: 10,
        },
      },
    });

    render(<Dashboard />, { wrapper });

    await waitFor(() => {
      expect(screen.getByText('10')).toBeInTheDocument(); // Active sessions
      expect(screen.getByText('5')).toBeInTheDocument();  // Total stations
    });
  });
});
```

---

## 📊 Component Examples

### StatsCard Component

```typescript
// src/components/features/stats/StatsCard.tsx
import { Badge } from '@base-ui-components/react/badge';
import { Skeleton } from '@base-ui-components/react/skeleton';
import clsx from 'clsx';

interface StatsCardProps {
  title: string;
  value: number;
  icon: string;
  color?: 'blue' | 'green' | 'purple' | 'orange';
  trend?: string;
  isLoading?: boolean;
}

const colorClasses = {
  blue: 'bg-blue-50 border-blue-200 text-blue-600',
  green: 'bg-green-50 border-green-200 text-green-600',
  purple: 'bg-purple-50 border-purple-200 text-purple-600',
  orange: 'bg-orange-50 border-orange-200 text-orange-600',
};

export function StatsCard({
  title,
  value,
  icon,
  color = 'blue',
  trend,
  isLoading,
}: StatsCardProps) {
  if (isLoading) {
    return (
      <div className="stats-card bg-white rounded-lg shadow-md p-6 border-2">
        <Skeleton className="h-8 w-8 mb-4" />
        <Skeleton className="h-4 w-24 mb-2" />
        <Skeleton className="h-8 w-16" />
      </div>
    );
  }

  return (
    <div
      className={clsx(
        'stats-card rounded-lg shadow-md p-6 border-2 transition-all hover:shadow-lg',
        colorClasses[color]
      )}
    >
      <div className="flex items-start justify-between mb-4">
        <span className="text-4xl">{icon}</span>
        {trend && (
          <Badge className="bg-white text-green-600">
            {trend}
          </Badge>
        )}
      </div>
      
      <h3 className="text-sm font-medium text-gray-600 mb-2">
        {title}
      </h3>
      
      <p className="text-3xl font-bold">
        {value.toLocaleString()}
      </p>
    </div>
  );
}
```

### StatsChart Component

```typescript
// src/components/features/stats/StatsChart.tsx
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';

interface StatsChartProps {
  title: string;
  data: any[];
  type: 'line' | 'bar';
  xKey?: string;
  yKey?: string;
}

export function StatsChart({
  title,
  data,
  type,
  xKey = 'name',
  yKey = 'value',
}: StatsChartProps) {
  const ChartComponent = type === 'line' ? LineChart : BarChart;
  const DataComponent = type === 'line' ? Line : Bar;

  return (
    <div className="stats-chart bg-white rounded-lg shadow-md p-6">
      <h3 className="text-lg font-semibold mb-4">{title}</h3>
      
      <ResponsiveContainer width="100%" height={300}>
        <ChartComponent data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey={xKey} />
          <YAxis />
          <Tooltip />
          <Legend />
          <DataComponent
            dataKey={yKey}
            fill="#3b82f6"
            stroke="#3b82f6"
            strokeWidth={2}
          />
        </ChartComponent>
      </ResponsiveContainer>
    </div>
  );
}
```

### ActiveSessionsMonitor Component

```typescript
// src/components/features/sessions/ActiveSessionsMonitor.tsx
import { useQuery } from '@tanstack/react-query';
import { Table } from '@base-ui-components/react/table';
import { Badge } from '@base-ui-components/react/badge';
import { statsService } from '@/services';
import { formatDistanceToNow } from 'date-fns';

export function ActiveSessionsMonitor() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['active-sessions-monitor'],
    queryFn: statsService.getStats,
    refetchInterval: 2000, // Refresh every 2 seconds
  });

  if (isLoading) {
    return <div>Loading sessions...</div>;
  }

  // Note: This is Phase 1 implementation with aggregated data
  // Phase 2 will add dedicated sessions endpoint

  return (
    <div className="active-sessions-monitor bg-white rounded-lg shadow-md">
      <div className="p-6 border-b">
        <div className="flex justify-between items-center">
          <h3 className="text-lg font-semibold">Active Sessions</h3>
          <Badge className="bg-blue-100 text-blue-800">
            {stats?.data.active_sessions || 0} Active
          </Badge>
        </div>
      </div>

      <div className="p-6">
        <Table>
          <Table.Head>
            <Table.Row>
              <Table.ColumnHeader>Session ID</Table.ColumnHeader>
              <Table.ColumnHeader>Station</Table.ColumnHeader>
              <Table.ColumnHeader>Duration</Table.ColumnHeader>
              <Table.ColumnHeader>Type</Table.ColumnHeader>
              <Table.ColumnHeader>Status</Table.ColumnHeader>
            </Table.Row>
          </Table.Head>
          <Table.Body>
            {stats?.data.active_sessions === 0 ? (
              <Table.Row>
                <Table.Cell colSpan={5} className="text-center text-gray-500 py-8">
                  No active sessions
                </Table.Cell>
              </Table.Row>
            ) : (
              <Table.Row>
                <Table.Cell colSpan={5} className="text-center text-gray-500 py-8">
                  {stats?.data.active_sessions} active session(s) - 
                  Full details in Phase 2
                </Table.Cell>
              </Table.Row>
            )}
          </Table.Body>
        </Table>
      </div>
    </div>
  );
}
```

---

## 🚀 Deployment

### Production Build

```bash
# Build for production
npm run build

# Preview production build locally
npm run preview

# Output will be in dist/ directory
```

### Environment Variables (Production)

```bash
# .env.production
VITE_API_BASE_URL=https://api.trainblink.org
VITE_APP_NAME=TrainBlink
VITE_APP_VERSION=1.0.0
VITE_ENABLE_DEV_TOOLS=false
```

### Deployment Options

**Option 1: Vercel**
```bash
npm install -g vercel
vercel --prod
```

**Option 2: Netlify**
```bash
npm install -g netlify-cli
netlify deploy --prod
```

**Option 3: Static Hosting (Nginx)**
```nginx
server {
    listen 80;
    server_name trainblink.org;
    root /var/www/trainblink-frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

---

## 📝 Development Guidelines

### Code Style

- Use TypeScript for all components
- Use functional components with hooks
- Use Base UI components where possible
- Follow React best practices
- Use ESLint and Prettier

### Naming Conventions

- Components: PascalCase (e.g., `StatsCard.tsx`)
- Hooks: camelCase with "use" prefix (e.g., `useStats.ts`)
- Services: camelCase with "Service" suffix (e.g., `statsService.ts`)
- Types: PascalCase (e.g., `StatsResponse`)
- Constants: UPPER_SNAKE_CASE (e.g., `API_BASE_URL`)

### File Organization

- Group by feature, not by type
- Keep related files together
- Use index.ts for clean imports
- Separate concerns (UI, logic, types)

---

## 🎯 Success Metrics

### Performance

- First Contentful Paint < 1.5s
- Time to Interactive < 3s
- Lighthouse Score > 90
- Bundle size < 500KB (gzipped)

### Functionality

- All API endpoints integrated
- Real-time updates working
- Error handling implemented
- Loading states implemented
- Responsive design working

### Code Quality

- Zero TypeScript errors
- ESLint passing
- 80%+ test coverage (Phase 5)
- No console errors
- Proper error boundaries

---

## 📚 Resources

### Documentation

- [Base UI Docs](https://base-ui.com/)
- [TanStack Query Docs](https://tanstack.com/query/latest)
- [React Router Docs](https://reactrouter.com/)
- [Tailwind CSS Docs](https://tailwindcss.com/)
- [Recharts Docs](https://recharts.org/)

### Design References

- [Base UI Dashboard Example](https://base-ui-tailwindcss-dashboard.vercel.app/ui-kit)
- [Tailwind UI Components](https://tailwindui.com/)
- [Headless UI Examples](https://headlessui.com/)

---

## 🤝 Contributing

### Development Workflow

1. Create feature branch
2. Implement feature
3. Write tests
4. Run linter and type-check
5. Submit pull request

### Code Review Checklist

- [ ] TypeScript types defined
- [ ] Error handling implemented
- [ ] Loading states added
- [ ] Responsive design tested
- [ ] Real data integration verified
- [ ] Documentation updated

---

## 📞 Support

For questions or issues:
- Backend API: See `docs/` in messenger_protocol_research
- Frontend Issues: Create GitHub issue
- General Questions: Contact development team

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-20
**Status**: ✅ Ready for Implementation
