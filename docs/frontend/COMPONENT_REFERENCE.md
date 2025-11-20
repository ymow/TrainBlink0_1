# TrainBlink Frontend - Component Reference

Quick reference for all components with Base UI usage

---

## 🎨 Component Library

### Common Components

#### Button
```typescript
import { Button } from '@base-ui-components/react/button';

// Primary button
<Button className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700">
  Download App
</Button>

// Secondary button
<Button className="bg-gray-200 text-gray-800 px-6 py-3 rounded-lg hover:bg-gray-300">
  Learn More
</Button>
```

#### Badge
```typescript
import { Badge } from '@base-ui-components/react/badge';

// Status badge
<Badge className="bg-green-100 text-green-800">Active</Badge>
<Badge className="bg-red-100 text-red-800">Inactive</Badge>
<Badge className="bg-blue-100 text-blue-800">Online</Badge>
```

#### Alert
```typescript
import { Alert } from '@base-ui-components/react/alert';

<Alert variant="error">
  <Alert.Title>Error</Alert.Title>
  <Alert.Description>
    Failed to load data. Please try again.
  </Alert.Description>
</Alert>

<Alert variant="success">
  <Alert.Title>Success</Alert.Title>
  <Alert.Description>
    Settings saved successfully.
  </Alert.Description>
</Alert>
```

#### Dialog/Modal
```typescript
import { Dialog } from '@base-ui-components/react/dialog';

const [open, setOpen] = useState(false);

<Dialog open={open} onOpenChange={setOpen}>
  <Dialog.Trigger asChild>
    <Button>Open Details</Button>
  </Dialog.Trigger>
  
  <Dialog.Portal>
    <Dialog.Overlay className="fixed inset-0 bg-black/50" />
    <Dialog.Content className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-white rounded-lg p-6 w-full max-w-md">
      <Dialog.Title className="text-xl font-bold mb-4">
        Station Details
      </Dialog.Title>
      <Dialog.Description>
        View detailed information about this station.
      </Dialog.Description>
      <Dialog.Close asChild>
        <Button>Close</Button>
      </Dialog.Close>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog>
```

#### Table
```typescript
import { Table } from '@base-ui-components/react/table';

<Table>
  <Table.Head>
    <Table.Row>
      <Table.ColumnHeader>ID</Table.ColumnHeader>
      <Table.ColumnHeader>Station</Table.ColumnHeader>
      <Table.ColumnHeader>Status</Table.ColumnHeader>
    </Table.Row>
  </Table.Head>
  <Table.Body>
    <Table.Row>
      <Table.Cell>001</Table.Cell>
      <Table.Cell>Tokyo Station</Table.Cell>
      <Table.Cell>
        <Badge className="bg-green-100 text-green-800">Active</Badge>
      </Table.Cell>
    </Table.Row>
  </Table.Body>
</Table>
```

#### Menu (Navigation)
```typescript
import { Menu } from '@base-ui-components/react/menu';

<Menu>
  <Menu.Item>
    <NavLink to="/admin/dashboard">Dashboard</NavLink>
  </Menu.Item>
  <Menu.Item>
    <NavLink to="/admin/stations">Stations</NavLink>
  </Menu.Item>
  <Menu.Separator />
  <Menu.Item>
    <NavLink to="/admin/settings">Settings</NavLink>
  </Menu.Item>
</Menu>
```

#### Tooltip
```typescript
import { Tooltip } from '@base-ui-components/react/tooltip';

<Tooltip>
  <Tooltip.Trigger>
    <span className="cursor-help">ℹ️</span>
  </Tooltip.Trigger>
  <Tooltip.Content className="bg-gray-800 text-white px-3 py-2 rounded text-sm">
    Additional information here
  </Tooltip.Content>
</Tooltip>
```

#### Skeleton (Loading)
```typescript
import { Skeleton } from '@base-ui-components/react/skeleton';

// Loading card
<div className="bg-white p-6 rounded-lg">
  <Skeleton className="h-8 w-8 mb-4" />
  <Skeleton className="h-4 w-24 mb-2" />
  <Skeleton className="h-8 w-16" />
</div>
```

---

## 🎯 Feature Components

### StatsCard
```typescript
// src/components/features/stats/StatsCard.tsx
interface StatsCardProps {
  title: string;
  value: number;
  icon: string;
  color?: 'blue' | 'green' | 'purple' | 'orange';
  trend?: string;
  isLoading?: boolean;
}

<StatsCard
  title="Active Users"
  value={10}
  icon="👥"
  color="blue"
  trend="+12%"
/>
```

### StatsChart
```typescript
// src/components/features/stats/StatsChart.tsx
interface StatsChartProps {
  title: string;
  data: any[];
  type: 'line' | 'bar';
}

<StatsChart
  title="Sessions Over Time"
  data={[
    { name: 'Mon', value: 10 },
    { name: 'Tue', value: 15 },
    { name: 'Wed', value: 12 },
  ]}
  type="line"
/>
```

### StationCard
```typescript
// src/components/features/stations/StationCard.tsx
interface StationCardProps {
  station: {
    id: string;
    name: string;
    city: string;
  };
  onClick: () => void;
}

<StationCard
  station={{
    id: 'station_tokyo_001',
    name: 'Tokyo Station',
    city: 'Tokyo',
  }}
  onClick={() => console.log('clicked')}
/>
```

### ActiveSessionsMonitor
```typescript
// src/components/features/sessions/ActiveSessionsMonitor.tsx
<ActiveSessionsMonitor />

// Automatically polls every 2 seconds
// Shows real-time active sessions
// Uses Base UI Table component
```

---

## 🎨 Layout Components

### AdminLayout
```typescript
// src/components/layout/AdminLayout/AdminLayout.tsx
<AdminLayout>
  {/* Sidebar with navigation */}
  {/* Main content area */}
  {/* Header with user info */}
</AdminLayout>
```

### Header (Landing)
```typescript
// src/components/layout/Header/Header.tsx
<Header>
  <Logo />
  <Nav>
    <NavLink to="/">Home</NavLink>
    <NavLink to="/features">Features</NavLink>
    <NavLink to="/admin">Admin</NavLink>
  </Nav>
</Header>
```

---

## 📊 Chart Components (Recharts)

### Line Chart
```typescript
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

<ResponsiveContainer width="100%" height={300}>
  <LineChart data={data}>
    <CartesianGrid strokeDasharray="3 3" />
    <XAxis dataKey="name" />
    <YAxis />
    <Tooltip />
    <Line type="monotone" dataKey="value" stroke="#3b82f6" strokeWidth={2} />
  </LineChart>
</ResponsiveContainer>
```

### Bar Chart
```typescript
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

<ResponsiveContainer width="100%" height={300}>
  <BarChart data={data}>
    <CartesianGrid strokeDasharray="3 3" />
    <XAxis dataKey="name" />
    <YAxis />
    <Tooltip />
    <Bar dataKey="value" fill="#3b82f6" />
  </BarChart>
</ResponsiveContainer>
```

---

## 🔧 Utility Functions

### Date Formatting
```typescript
import { format, formatDistanceToNow } from 'date-fns';

// Format date
format(new Date(), 'yyyy-MM-dd HH:mm:ss')
// "2025-11-20 14:30:00"

// Relative time
formatDistanceToNow(new Date(), { addSuffix: true })
// "2 minutes ago"
```

### Class Names (clsx)
```typescript
import clsx from 'clsx';

<div className={clsx(
  'base-class',
  isActive && 'active-class',
  isDisabled && 'disabled-class'
)}>
  Content
</div>
```

### Number Formatting
```typescript
// src/utils/formatters.ts
export function formatNumber(num: number): string {
  return num.toLocaleString();
}

formatNumber(1234567); // "1,234,567"
```

---

## 🎨 Tailwind Utility Classes

### Common Patterns

**Container**
```html
<div className="container mx-auto px-4">
  Content
</div>
```

**Grid Layout**
```html
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
  <!-- Cards -->
</div>
```

**Flex Layout**
```html
<div className="flex items-center justify-between">
  <span>Left</span>
  <span>Right</span>
</div>
```

**Card**
```html
<div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition">
  Card content
</div>
```

**Button**
```html
<button className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition">
  Click me
</button>
```

---

## 🔌 API Integration Patterns

### Using useQuery
```typescript
import { useQuery } from '@tanstack/react-query';
import { statsService } from '@/services';

function MyComponent() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['stats'],
    queryFn: statsService.getStats,
    refetchInterval: 5000,
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return <div>Data: {JSON.stringify(data)}</div>;
}
```

### Using useMutation (Phase 2)
```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { geofenceService } from '@/services';

function EnterStationButton() {
  const queryClient = useQueryClient();
  
  const mutation = useMutation({
    mutationFn: (data) => geofenceService.enterStation(data, 'user-123', 'device-456'),
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: ['stats'] });
    },
  });

  return (
    <button onClick={() => mutation.mutate(stationData)}>
      Enter Station
    </button>
  );
}
```

---

## 🎯 Complete Component Example

### Full Stats Card with Base UI

```typescript
// src/components/features/stats/StatsCard.tsx
import { Badge } from '@base-ui-components/react/badge';
import { Skeleton } from '@base-ui-components/react/skeleton';
import { Tooltip } from '@base-ui-components/react/tooltip';
import clsx from 'clsx';

interface StatsCardProps {
  title: string;
  value: number;
  icon: string;
  color?: 'blue' | 'green' | 'purple' | 'orange';
  trend?: string;
  isLoading?: boolean;
  tooltip?: string;
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
  tooltip,
}: StatsCardProps) {
  if (isLoading) {
    return (
      <div className="stats-card bg-white rounded-lg shadow-md p-6 border-2">
        <Skeleton className="h-8 w-8 mb-4 rounded-full" />
        <Skeleton className="h-4 w-24 mb-2" />
        <Skeleton className="h-8 w-16" />
      </div>
    );
  }

  return (
    <div
      className={clsx(
        'stats-card rounded-lg shadow-md p-6 border-2 transition-all hover:shadow-lg cursor-pointer',
        colorClasses[color]
      )}
    >
      <div className="flex items-start justify-between mb-4">
        <span className="text-4xl">{icon}</span>
        <div className="flex gap-2">
          {trend && (
            <Badge className="bg-white text-green-600 font-semibold">
              {trend}
            </Badge>
          )}
          {tooltip && (
            <Tooltip>
              <Tooltip.Trigger>
                <span className="cursor-help text-gray-400">ℹ️</span>
              </Tooltip.Trigger>
              <Tooltip.Content className="bg-gray-800 text-white px-3 py-2 rounded text-sm max-w-xs">
                {tooltip}
              </Tooltip.Content>
            </Tooltip>
          )}
        </div>
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

// Usage:
<StatsCard
  title="Active Users"
  value={42}
  icon="👥"
  color="blue"
  trend="+12%"
  tooltip="Number of users currently in stations"
  isLoading={false}
/>
```

---

## 📚 Import Patterns

### Absolute Imports (using @ alias)
```typescript
import { apiService } from '@/services/api.service';
import { useStats } from '@/hooks/useStats';
import { StatsCard } from '@/components/features/stats';
import { API_CONFIG } from '@/config/api.config';
import type { StatsResponse } from '@/types/api.types';
```

### Index Files for Clean Imports
```typescript
// src/components/features/stats/index.ts
export { StatsCard } from './StatsCard';
export { StatsChart } from './StatsChart';
export { RealTimeCounter } from './RealTimeCounter';

// Usage:
import { StatsCard, StatsChart } from '@/components/features/stats';
```

---

This reference provides all the building blocks needed for the TrainBlink frontend. Refer to the main implementation plan for the complete architecture and development workflow.

