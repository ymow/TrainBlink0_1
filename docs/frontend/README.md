# TrainBlink Frontend Documentation

**Modern React Frontend with Base UI and Real-Time Data Integration**

---

## 📚 Documentation Index

### Primary Documents

1. **[FRONTEND_IMPLEMENTATION_PLAN.md](./FRONTEND_IMPLEMENTATION_PLAN.md)** (Main Document)
   - Complete implementation plan
   - Project structure
   - Landing page design
   - Admin panel design
   - API integration strategy
   - Implementation phases (5 weeks)
   - Code examples
   - Testing strategy
   - Deployment guide

2. **[QUICK_START_GUIDE.md](./QUICK_START_GUIDE.md)** (Get Started in 10 Minutes)
   - Quick setup instructions
   - Essential files to create
   - Testing checklist
   - Common issues & solutions

3. **[COMPONENT_REFERENCE.md](./COMPONENT_REFERENCE.md)** (Component Library)
   - Base UI component examples
   - Feature component patterns
   - Tailwind utility classes
   - API integration patterns
   - Complete code examples

---

## 🎯 Project Overview

### What We're Building

1. **Landing Page** - Public-facing website
   - Hero section with call-to-action
   - Real-time statistics from API
   - Feature showcase
   - Download links (iOS/Android)
   - Responsive design

2. **Admin Panel** - System monitoring dashboard
   - Real-time dashboard with charts
   - Station management
   - Active sessions monitoring
   - Matrix room status
   - Data visualization

### Technology Stack

**Frontend Framework**:
- React 18.3+ with TypeScript
- Vite (build tool)
- React Router (navigation)

**UI Components**:
- Base UI (@base-ui-components/react)
- Tailwind CSS 3.x

**State Management**:
- TanStack Query v5 (API state)
- React Context (global state)

**Data Visualization**:
- Recharts 2.x

**Backend Integration**:
- Axios (HTTP client)
- Real Go backend at http://localhost:8080
- NO MOCK DATA - All real API calls

---

## 🚀 Quick Start

```bash
# 1. Create project
npm create vite@latest trainblink-frontend -- --template react-ts
cd trainblink-frontend

# 2. Install dependencies
npm install @base-ui-components/react react-router-dom @tanstack/react-query axios recharts date-fns clsx
npm install -D tailwindcss postcss autoprefixer @types/node

# 3. Initialize Tailwind
npx tailwindcss init -p

# 4. Create directory structure
mkdir -p src/{components,pages,services,hooks,types,utils,styles,config}

# 5. Start development
npm run dev
```

**See [QUICK_START_GUIDE.md](./QUICK_START_GUIDE.md) for detailed setup instructions.**

---

## 📁 Project Structure

```
trainblink-frontend/
├── src/
│   ├── components/          # Reusable components
│   │   ├── common/         # Base components
│   │   ├── layout/         # Layout components
│   │   └── features/       # Feature-specific components
│   │       ├── stats/      # Statistics components
│   │       ├── stations/   # Station components
│   │       └── sessions/   # Session components
│   │
│   ├── pages/              # Page components
│   │   ├── landing/        # Landing page
│   │   └── admin/          # Admin panel pages
│   │
│   ├── services/           # API services
│   │   ├── api.service.ts  # Base API client
│   │   ├── geofence.service.ts
│   │   └── stats.service.ts
│   │
│   ├── hooks/              # Custom React hooks
│   │   ├── useStats.ts
│   │   └── useServerHealth.ts
│   │
│   ├── types/              # TypeScript types
│   │   └── api.types.ts
│   │
│   ├── config/             # Configuration
│   │   └── api.config.ts
│   │
│   └── utils/              # Utility functions
│
├── public/                 # Static assets
├── .env.development       # Development config
└── package.json
```

---

## 🔌 Backend Integration

### Available API Endpoints

```bash
# Health Check
GET http://localhost:8080/ping
GET http://localhost:8080/health

# Geofence API
POST http://localhost:8080/api/v1/geofence/enter
POST http://localhost:8080/api/v1/geofence/exit
GET  http://localhost:8080/api/v1/geofence/stats
```

### Real-Time Data

The frontend uses TanStack Query for automatic polling:

- **Landing Page Stats**: Poll every 5 seconds
- **Admin Dashboard**: Poll every 3 seconds
- **Active Sessions**: Poll every 2 seconds

### Example API Call

```typescript
import { useQuery } from '@tanstack/react-query';
import { statsService } from '@/services';

function Dashboard() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['stats'],
    queryFn: statsService.getStats,
    refetchInterval: 3000, // Poll every 3 seconds
  });

  return <div>Active Users: {data?.active_sessions}</div>;
}
```

---

## 📊 Implementation Phases

### Phase 1: Foundation & Landing Page (Week 1)
- Setup project structure
- Implement type definitions & services
- Build landing page
- Integrate real-time stats
- Responsive design

**Deliverables**: Fully functional landing page with real-time data

### Phase 2: Admin Panel - Dashboard (Week 2)
- Create admin layout
- Build dashboard with stats cards
- Add data visualization charts
- Implement active sessions monitor

**Deliverables**: Admin dashboard with real-time monitoring

### Phase 3: Admin Panel - Stations & Sessions (Week 3)
- Build stations management page
- Create sessions monitoring page
- Add Matrix room status display
- Implement search/filter

**Deliverables**: Complete admin panel functionality

### Phase 4: Analytics & Visualization (Week 4)
- Advanced charts
- Metrics & insights
- Export functionality
- Trend analysis

**Deliverables**: Comprehensive analytics page

### Phase 5: Polish & Optimization (Week 5)
- Performance optimization
- Accessibility improvements
- Testing
- Documentation
- Deployment preparation

**Deliverables**: Production-ready application

**Total Duration**: 5 weeks

---

## 🎨 Design System

### Base UI Components Used

- **Button** - Call-to-action buttons
- **Badge** - Status indicators
- **Table** - Session data display
- **Dialog** - Modal dialogs
- **Alert** - Error/success messages
- **Tooltip** - Contextual help
- **Menu** - Navigation menus
- **Skeleton** - Loading states

### Color Palette

```javascript
colors: {
  primary: {
    500: '#0ea5e9', // Blue
    600: '#0284c7',
  },
  success: { 500: '#10b981' },
  warning: { 500: '#f59e0b' },
  error: { 500: '#ef4444' },
}
```

### Typography

- Font Family: Inter (sans-serif)
- Headings: Bold, 2xl-5xl
- Body: Regular, base-lg
- Code: Fira Code (monospace)

---

## 🧪 Testing

### Testing Strategy

1. **Unit Tests** - Component testing with Vitest
2. **Integration Tests** - API integration testing
3. **E2E Tests** - Full user flow testing (Phase 5)

### Example Test

```typescript
import { render, screen } from '@testing-library/react';
import { StatsCard } from './StatsCard';

test('renders stats card with value', () => {
  render(
    <StatsCard
      title="Active Users"
      value={42}
      icon="👥"
    />
  );
  
  expect(screen.getByText('42')).toBeInTheDocument();
  expect(screen.getByText('Active Users')).toBeInTheDocument();
});
```

---

## 🚀 Deployment

### Production Build

```bash
npm run build
# Output: dist/ directory
```

### Deployment Options

1. **Vercel** (Recommended)
   ```bash
   vercel --prod
   ```

2. **Netlify**
   ```bash
   netlify deploy --prod
   ```

3. **Static Hosting** (Nginx)
   ```nginx
   server {
       listen 80;
       root /var/www/trainblink-frontend/dist;
       index index.html;
       
       location / {
           try_files $uri $uri/ /index.html;
       }
   }
   ```

---

## 📈 Performance Goals

- First Contentful Paint < 1.5s
- Time to Interactive < 3s
- Lighthouse Score > 90
- Bundle size < 500KB (gzipped)

---

## 🤝 Contributing

### Development Workflow

1. Create feature branch
2. Implement feature
3. Write tests
4. Run linter: `npm run lint`
5. Submit pull request

### Code Style

- Use TypeScript for all components
- Use functional components with hooks
- Follow React best practices
- Use ESLint and Prettier

---

## 📞 Support & Resources

### Documentation

- [Main Implementation Plan](./FRONTEND_IMPLEMENTATION_PLAN.md)
- [Quick Start Guide](./QUICK_START_GUIDE.md)
- [Component Reference](./COMPONENT_REFERENCE.md)

### External Resources

- [Base UI Documentation](https://base-ui.com/)
- [TanStack Query Docs](https://tanstack.com/query/latest)
- [Tailwind CSS Docs](https://tailwindcss.com/)
- [React Router Docs](https://reactrouter.com/)
- [Recharts Docs](https://recharts.org/)

### Backend Documentation

- Backend API: `../GEOFENCE_HYBRID_IMPLEMENTATION.md`
- Protocol Spec: `../PROTOCOL_SPECIFICATION.md`
- Integration Guide: `../CLIENT_INTEGRATION_GUIDE.md`

---

## 🎯 Success Criteria

### Functionality
- All API endpoints integrated
- Real-time updates working
- Error handling implemented
- Loading states implemented
- Responsive design working

### Performance
- Fast page loads (< 2s)
- Smooth animations
- No UI jank
- Efficient polling

### Code Quality
- Zero TypeScript errors
- ESLint passing
- Proper error boundaries
- Clean component architecture

---

## 📋 Checklist

### Phase 1 (Week 1)
- [ ] Project setup complete
- [ ] Type definitions created
- [ ] API services implemented
- [ ] Landing page built
- [ ] Real-time stats working
- [ ] Responsive design implemented

### Phase 2 (Week 2)
- [ ] Admin layout created
- [ ] Dashboard with charts
- [ ] Stats cards with real data
- [ ] Active sessions monitor

### Phase 3 (Week 3)
- [ ] Stations page complete
- [ ] Sessions monitoring page
- [ ] Matrix room status
- [ ] Search/filter functionality

### Phase 4 (Week 4)
- [ ] Analytics page
- [ ] Advanced charts
- [ ] Export functionality
- [ ] Trend analysis

### Phase 5 (Week 5)
- [ ] Performance optimized
- [ ] Accessibility improved
- [ ] Tests written
- [ ] Documentation complete
- [ ] Production deployment

---

## 🎉 Getting Started

Ready to build? Start with the **[QUICK_START_GUIDE.md](./QUICK_START_GUIDE.md)** to get your development environment set up in 10 minutes!

For the complete implementation plan, see **[FRONTEND_IMPLEMENTATION_PLAN.md](./FRONTEND_IMPLEMENTATION_PLAN.md)**.

For component examples and patterns, see **[COMPONENT_REFERENCE.md](./COMPONENT_REFERENCE.md)**.

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-20
**Status**: ✅ Ready for Implementation

