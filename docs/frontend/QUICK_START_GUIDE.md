# TrainBlink Frontend - Quick Start Guide

**Getting Started in 10 Minutes** ⚡

---

## 🚀 Quick Setup

```bash
# 1. Create project
npm create vite@latest trainblink-frontend -- --template react-ts
cd trainblink-frontend

# 2. Install all dependencies at once
npm install @base-ui-components/react react-router-dom @tanstack/react-query axios recharts date-fns clsx

# 3. Install dev dependencies
npm install -D tailwindcss postcss autoprefixer @types/node @tanstack/react-query-devtools

# 4. Initialize Tailwind
npx tailwindcss init -p

# 5. Create directory structure
mkdir -p src/{components/{common,layout,features/{stats,stations,sessions}},pages/{landing,admin},services,hooks,types,utils,styles,config}

# 6. Start backend (in separate terminal)
cd ../messenger_protocol_research
go run cmd/server/main_geofence_hybrid.go

# 7. Start frontend
npm run dev
```

---

## 📁 Essential Files to Create First

### 1. Tailwind Config
```javascript
// tailwind.config.js
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: { extend: {} },
  plugins: [],
}
```

### 2. Global Styles
```css
/* src/styles/index.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
```

### 3. API Config
```typescript
// src/config/api.config.ts
export const API_CONFIG = {
  BASE_URL: 'http://localhost:8080',
  TIMEOUT: 10000,
  POLLING: {
    STATS_LANDING: 5000,
    STATS_DASHBOARD: 3000,
  },
};

export const ENDPOINTS = {
  PING: '/ping',
  GEOFENCE_STATS: '/api/v1/geofence/stats',
};
```

### 4. Type Definitions
```typescript
// src/types/api.types.ts
export interface StatsResponse {
  status: string;
  data: {
    total_stations: number;
    total_sessions: number;
    active_sessions: number;
    matrix_stats: {
      total_rooms: number;
      active_rooms: number;
      total_users: number;
      active_users: number;
    };
  };
}
```

### 5. API Service
```typescript
// src/services/api.service.ts
import axios from 'axios';
import { API_CONFIG } from '@/config/api.config';

export const apiService = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
});
```

### 6. Stats Service
```typescript
// src/services/stats.service.ts
import { apiService } from './api.service';
import { StatsResponse } from '@/types/api.types';

export const statsService = {
  getStats: async (): Promise<StatsResponse> => {
    const response = await apiService.get('/api/v1/geofence/stats');
    return response.data;
  },
};
```

### 7. Stats Hook
```typescript
// src/hooks/useStats.ts
import { useQuery } from '@tanstack/react-query';
import { statsService } from '@/services/stats.service';

export function useStats(refetchInterval = 5000) {
  return useQuery({
    queryKey: ['stats'],
    queryFn: statsService.getStats,
    refetchInterval,
  });
}
```

### 8. Main Entry Point
```typescript
// src/main.tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from './App';
import './styles/index.css';

const queryClient = new QueryClient();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </React.StrictMode>
);
```

### 9. App Component (Landing Page)
```typescript
// src/App.tsx
import { useStats } from './hooks/useStats';

function App() {
  const { data: stats, isLoading, error } = useStats();

  if (error) return <div>Error: Server not running?</div>;
  if (isLoading) return <div>Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-blue-600 text-white py-20 text-center">
        <h1 className="text-5xl font-bold mb-4">TrainBlink</h1>
        <p className="text-xl">Connect at Train Stations</p>
      </header>

      <section className="py-16">
        <div className="container mx-auto px-4">
          <h2 className="text-3xl font-bold text-center mb-12">
            Live Statistics
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <StatsCard
              title="Active Users"
              value={stats?.data.active_sessions || 0}
              icon="👥"
            />
            <StatsCard
              title="Total Stations"
              value={stats?.data.total_stations || 0}
              icon="🚉"
            />
            <StatsCard
              title="Total Sessions"
              value={stats?.data.total_sessions || 0}
              icon="💬"
            />
          </div>
        </div>
      </section>
    </div>
  );
}

function StatsCard({ title, value, icon }: any) {
  return (
    <div className="bg-white rounded-lg shadow-md p-6 text-center">
      <div className="text-4xl mb-4">{icon}</div>
      <h3 className="text-gray-600 text-sm mb-2">{title}</h3>
      <p className="text-3xl font-bold">{value}</p>
    </div>
  );
}

export default App;
```

---

## 🧪 Testing the Setup

```bash
# 1. Check if backend is running
curl http://localhost:8080/ping
# Should return: {"message":"pong",...}

# 2. Check stats endpoint
curl http://localhost:8080/api/v1/geofence/stats
# Should return: {"status":"success","data":{...}}

# 3. Start frontend
npm run dev
# Open: http://localhost:5173

# 4. Verify real-time updates
# Stats should update every 5 seconds automatically
```

---

## 📦 Package.json

```json
{
  "name": "trainblink-frontend",
  "private": true,
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "@base-ui-components/react": "latest",
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
    "@vitejs/plugin-react": "^4.2.0",
    "autoprefixer": "^10.4.16",
    "postcss": "^8.4.32",
    "tailwindcss": "^3.3.6",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

---

## 🔧 Vite Config (Path Alias)

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
  },
});
```

```json
// tsconfig.json (add to compilerOptions)
{
  "compilerOptions": {
    // ... other options
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

---

## 🎯 Next Steps

### Phase 1: Landing Page (Week 1)
1. ✅ Setup project (Day 1)
2. Create Hero section (Day 2)
3. Add Features section (Day 3)
4. Add Download section (Day 4)
5. Polish & responsive (Day 5)

### Phase 2: Admin Dashboard (Week 2)
1. Create admin layout with sidebar
2. Add dashboard with charts
3. Add stations page
4. Add sessions monitoring
5. Polish & testing

---

## 🐛 Common Issues

### Issue: API calls failing
```bash
# Check if backend is running
curl http://localhost:8080/ping

# Check CORS headers
curl -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS http://localhost:8080/api/v1/geofence/stats
```

### Issue: Import aliases not working
```bash
# Make sure tsconfig.json has paths configured
# Restart TypeScript server in VS Code: Cmd+Shift+P -> "TypeScript: Restart TS Server"
```

### Issue: Tailwind classes not working
```bash
# Make sure tailwind.config.js content paths are correct
# Make sure index.css has @tailwind directives
# Restart Vite dev server
```

---

## 📚 Resources

- Full Implementation Plan: `FRONTEND_IMPLEMENTATION_PLAN.md`
- Backend API Docs: `../GEOFENCE_HYBRID_IMPLEMENTATION.md`
- Base UI Docs: https://base-ui.com/
- TanStack Query: https://tanstack.com/query/latest

---

## 🎉 Success!

If you can see stats updating every 5 seconds, you're ready to go!

**Next**: Follow Phase 1 in the main implementation plan.

