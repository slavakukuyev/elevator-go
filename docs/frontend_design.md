# Frontend Design Specification (Updated with Modern Recommendations)

## 🏗️ Overview

This document outlines the design for a modern, interactive frontend for the Elevator Control System that will be deployed as a lightweight static application on GitHub Pages. The frontend provides real-time visualization of elevator operations, intelligent touch display interfaces for floor selection, and comprehensive monitoring capabilities, all built with modern best practices for performance, accessibility, and developer experience.

## 🎯 Goals

- **Interactive Simulation**: Real-time visualization of elevator movement with smooth animations and micro-interactions
- **Modern UX**: Clean, minimalist interface with intuitive controls and consistent theming (light/dark mode)
- **Real-time Updates**: WebSocket integration for live status monitoring with robust reconnection
- **Lightweight Deployment**: Static site optimized for GitHub Pages hosting with PWA capabilities
- **Scalable Architecture**: Component-based design supporting multiple elevators with TypeScript
- **Accessibility**: WCAG 2.1 AA compliant interface with full keyboard navigation and screen reader support
- **Performance**: < 100KB initial bundle with optimized loading and animations
- **Developer Experience**: Type-safe development with comprehensive testing and modern tooling

## 🚀 Technology Stack

### Core Framework
- **Svelte/SvelteKit (with TypeScript)**: Modern, compile-time optimized framework with minimal runtime overhead
  - Generates highly efficient vanilla JavaScript with code splitting
  - Excellent performance characteristics for GitHub Pages
  - Small bundle size ideal for static hosting
  - Built-in reactivity and state management
  - TypeScript for type safety and enhanced maintainability

### Styling & Layout
- **Tailwind CSS**: Utility-first CSS framework for rapid, consistent styling
  - Purged and minified in production to keep CSS bundle tiny
  - Design system implemented via Tailwind config and CSS variables
- **CSS Grid/Flexbox**: For responsive elevator building layouts
- **CSS Custom Properties**: For dynamic theming and animation variables

### Animation & Interactivity
- **Svelte Transitions**: Built-in animation system for smooth state changes
- **CSS Transforms**: Hardware-accelerated elevator movement animations with easing
- **Micro-interactions**: Subtle hover and active effects for responsive feedback
- **Intersection Observer API**: Performance-optimized viewport monitoring

### Real-time Communication
- **WebSocket API**: Native browser WebSocket for real-time elevator status
- **Reconnection Logic**: Robust connection management with exponential backoff
- **Error Handling**: Graceful degradation and user feedback

### Build & Deployment
- **Vite**: Lightning-fast build tool with HMR for development
- **GitHub Actions**: Automated CI/CD pipeline for GitHub Pages deployment
- **Static Site Generation**: Pre-rendered pages for optimal loading performance
- **Service Worker**: Offline caching and PWA capabilities
- **Asset Optimization**: Image compression, lazy loading, and bundle optimization

## 📐 Architecture Design

### Component Hierarchy

```
App.svelte
├── Header.svelte (top navigation bar, status indicators, theme toggle)
├── ElevatorControlPanel.svelte (left side panel for controls & creation)
│   ├── CreateElevatorModal.svelte (modal dialog for adding elevator)
│   └── SystemStatus.svelte (summary of system health, connect status)
├── ElevatorBuildingGrid.svelte (main view: grid of building(s))
│   └── ElevatorBuilding.svelte (a single elevator building column)
│       ├── ElevatorShaft.svelte (visual shaft containing floors)
│       ├── ElevatorCar.svelte (the moving elevator car)
│       ├── FloorRow.svelte (represents one floor in the shaft)
│       ├── CallButton.svelte (up/down call buttons for a floor)
│       └── FloorTouchDisplay.svelte (floor's touch panel for selecting destination)
├── MonitoringDashboard.svelte (right side panel or separate view for monitoring)
│   ├── MetricsPanel.svelte (real-time charts and metrics)
│   └── HealthStatus.svelte (system & elevator health overview)
└── Footer.svelte (footer with branding or links)
```

### State Management

```typescript
// stores.ts - Svelte stores for global state with TypeScript
import { writable, derived } from 'svelte/store';
import type { Elevator, SystemStatus } from './types';

export const elevators = writable<Elevator[]>([]);
export const systemStatus = writable<SystemStatus>({ healthy: true, elevatorCount: 0 });
export const isConnected = writable(false);          // WebSocket connection status
export const currentFloor = writable<number>(0);     // Global current floor context
export const selectedElevator = writable<Elevator | null>(null);

// Derived store: elevators serving the currentFloor (for context-sensitive UI)
export const availableElevators = derived(
  [currentFloor, elevators],
  ([$currentFloor, $elevators]) =>
    $elevators.filter(elev => 
      $currentFloor >= elev.minFloor && $currentFloor <= elev.maxFloor
    )
);

// Other derived states for performance optimization
export const idleElevators = derived(elevators, $elevators => 
  $elevators.filter(elev => elev.status === 'idle')
);

export const errorElevators = derived(elevators, $elevators => 
  $elevators.filter(elev => elev.status === 'error')
);
```

### Data Types (TypeScript interfaces)

```typescript
// types.ts - Type definitions for type safety
interface Elevator {
  name: string;
  minFloor: number;
  maxFloor: number;
  currentFloor: number;
  status: 'idle' | 'moving' | 'error';
  direction: 'up' | 'down' | null;
  doorsOpen: boolean;
  hasPassenger: boolean;
  threshold?: number;
}

interface SystemStatus {
  healthy: boolean;
  elevatorCount: number;
  lastMaintenance?: Date;
  alerts?: Alert[];
}

interface Alert {
  id: string;
  type: 'warning' | 'error' | 'info';
  message: string;
  timestamp: Date;
}

interface FloorRequest {
  from: number;
  to: number;
  timestamp: Date;
  status: 'pending' | 'assigned' | 'completed' | 'failed';
}
```

## 🎨 User Interface Design

### 1. Header Section (Top Bar)
- **Brand Identity**: Elevator Control System logo and title on the left
- **Connection Status**: Small indicator showing WebSocket status (green dot + "Connected" or red "Disconnected")
- **Theme Toggle**: Button (🌞/🌜 icon) to switch between light and dark mode with aria-label
- **User Actions**: Optional menu with Help link, refresh button, or monitoring panel toggle
- **Layout**: Flex container with spaced-between alignment, responsive on mobile

### 2. Control Panel (Sidebar)
- **System Status Summary**: Small section showing overall health ("All systems operational" or warning icon)
- **Create Elevator Button**: Prominent "+ Add Elevator" CTA with keyboard shortcut support
- **Create Elevator Modal**: Accessible modal dialog with form validation:
  ```
  ┌─ Create New Elevator ───────────────┐
  │ Name:        [__________]          │
  │ Min Floor:   [__]   Max Floor: [__] │
  │ Capacity:    [__]   (optional)      │
  │                                     │
  │              [Cancel]  [Create]     │
  └─────────────────────────────────────┘
  ```
- **Real-time Validation**: Inline error messages with aria-live assertive
- **Success Feedback**: Toast notifications for successful operations
- **Mobile Adaptation**: Collapsible off-canvas sidebar on small screens

### 3. Building Visualization Grid (Main View)

#### Layout Strategy
- **CSS Grid**: Responsive grid accommodating multiple elevator buildings
- **Floor Alignment**: Common baseline (floor 0) across all buildings
- **Scalability**: Dynamic grid sizing based on elevator count
- **Responsive Design**: Mobile-first approach with breakpoint adaptations

#### Individual Building Structure
```
┌─ Elevator: "Main-A" ─┐
│  ┌────┐  Floor 10   │ ← Call Buttons
│  │ 🔺 │ [↑] [↓]     │
│  │ E  │  Floor 9    │
│  │ L  │ [↑] [↓]     │
│  │ E  │  Floor 8    │
│  │ V  │ [↑] [↓]     │
│  │[●]│  Floor 7    │ ← Elevator Car Position
│  │ T  │ [↑] [↓]     │
│  │ O  │  Floor 6    │
│  │ R  │ [↑] [↓]     │
│  │    │  Floor 5    │
│  │    │ [↑] [↓]     │
│  │    │  Floor 4    │
│  │    │ [↑] [↓]     │
│  │    │  Floor 3    │
│  │    │ [↑] [↓]     │
│  │    │  Floor 2    │
│  │    │ [↑] [↓]     │
│  │    │  Floor 1    │
│  │    │ [↑] [↓]     │
│  └────┘  Floor 0    │ ← Ground Floor Baseline
└───────────────────────┘
```

### 4. Interactive Elements & Feedback

#### Elevator Car Movement & Feedback
- **Smooth Motion**: CSS transitions with cubic-bezier easing for natural movement
- **State Colors**: 
  - Moving: blue background with subtle animation
  - Idle: green background
  - Error: red background with alert icon
- **Direction Indicator**: Arrow (↑/↓) showing current movement direction
- **Door Animation**: Sliding doors with transform animations
- **Arrival Feedback**: Visual and optional audio cues when elevator arrives

#### Call Buttons
- **Visual States**:
  - Idle: Default state, neutral color
  - Pressed/Pending: Active state with blue background and pulse animation
  - Serviced: Resets to idle when elevator arrives
- **Accessibility**: Proper aria-labels and keyboard support
- **Prevent Double Requests**: Ignore duplicate calls while pending

#### Floor Touch Display
- **Available Destinations**: Computed based on elevator ranges and current floor
- **Styling**: Special styling for basement (negative) and ground (0) floors
- **User Flow**: Immediate feedback on selection with error handling
- **Accessibility**: Keyboard navigation and screen reader support

### 5. Monitoring Dashboard (Sidebar/View)
- **Metrics Panel**: Performance metrics, request queue, system throughput
- **Health Status**: Connection status, elevator health, system alerts
- **Visualization**: Simple SVG charts for lightweight performance
- **Mobile Adaptation**: Hidden by default on small screens, accessible via menu

## 🔄 Real-time Data Flow

### WebSocket Integration

```typescript
// websocket.ts - WebSocket service with TypeScript
class ElevatorWebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  connect() {
    const url = import.meta.env.VITE_WS_URL || 'wss://example.com/ws/status';
    try {
      this.ws = new WebSocket(url);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        isConnected.set(true);
        this.reconnectAttempts = 0;
      };
      
      this.ws.onmessage = (event) => {
        const status = JSON.parse(event.data);
        this.handleStatusUpdate(status);
      };
      
      this.ws.onclose = () => {
        isConnected.set(false);
        console.warn('WebSocket closed');
        this.handleReconnection();
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
    } catch (error) {
      console.error('WebSocket connection failed:', error);
      this.handleReconnection();
    }
  }

  private handleStatusUpdate(status: any) {
    systemStatus.set(status.system || { 
      healthy: true, 
      elevatorCount: Object.keys(status.elevators || {}).length 
    });
    
    if (status.elevators) {
      elevators.update(currentElevators => {
        return currentElevators.map(elevator => {
          const update = status.elevators[elevator.name];
          return update ? { ...elevator, ...update } : elevator;
        });
      });
    }
  }

  private handleReconnection() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      const delay = this.reconnectDelay;
      this.reconnectAttempts++;
      this.reconnectDelay *= 2; // exponential backoff
      setTimeout(() => {
        console.log(`Reconnecting... attempt ${this.reconnectAttempts}`);
        this.connect();
      }, delay);
    } else {
      console.error('Max WebSocket reconnection attempts reached');
    }
  }
}

export const wsService = new ElevatorWebSocketService();
```

### API Integration

```typescript
// api.ts - HTTP API service with TypeScript
const API_BASE = import.meta.env.VITE_API_URL || 'https://example.com/api/v1';

export const elevatorAPI = {
  async createElevator(config: Partial<Elevator>): Promise<Elevator> {
    const response = await fetch(`${API_BASE}/elevators`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config)
    });
    
    if (!response.ok) {
      const msg = await response.text();
      throw new Error(`Failed to create elevator: ${msg}`);
    }
    
    const newElev = await response.json();
    elevators.update(list => [...list, newElev]);
    return newElev;
  },

  async requestFloor(fromFloor: number, toFloor: number): Promise<any> {
    const response = await fetch(`${API_BASE}/floors/request`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ from: fromFloor, to: toFloor })
    });
    
    if (!response.ok) {
      const msg = await response.text();
      throw new Error(`Floor request failed: ${msg}`);
    }
    
    return response.json();
  },

  async getHealthStatus(): Promise<SystemStatus | null> {
    const res = await fetch(`${API_BASE}/health`);
    return res.ok ? res.json() : null;
  },

  async getMetrics(): Promise<any> {
    const res = await fetch(`${API_BASE}/metrics`);
    return res.ok ? res.json() : null;
  }
};
```

### Floor Selection Algorithm

```typescript
// floorSelection.ts - Floor selection utilities with TypeScript
export class FloorSelectionService {
  /**
   * Calculate available destination floors from a given floor.
   */
  getAvailableDestinations(currentFloor: number, elevators: Elevator[]): number[] {
    const availableFloors = new Set<number>();
    
    elevators.filter(elev => 
      currentFloor >= elev.minFloor && currentFloor <= elev.maxFloor
    ).forEach(elev => {
      for (let f = elev.minFloor; f <= elev.maxFloor; f++) {
        if (f !== currentFloor) availableFloors.add(f);
      }
    });
    
    return Array.from(availableFloors).sort((a, b) => a - b);
  }

  /**
   * Find the optimal elevator to serve a requested trip.
   */
  findOptimalElevator(fromFloor: number, toFloor: number, elevators: Elevator[]): Elevator | null {
    const candidates = elevators.filter(elev =>
      fromFloor >= elev.minFloor && fromFloor <= elev.maxFloor &&
      toFloor >= elev.minFloor && toFloor <= elev.maxFloor
    );
    
    if (candidates.length === 0) return null;
    
    const direction = toFloor > fromFloor ? 'up' : 'down';
    
    // 1. Elevators already moving in the required direction
    const movingSameWay = candidates.filter(elev => 
      elev.status === 'moving' && elev.direction === direction
    );
    if (movingSameWay.length > 0) {
      return this.findClosestElevator(fromFloor, movingSameWay);
    }
    
    // 2. Idle elevators closest to the origin floor
    const idleElevators = candidates.filter(elev => elev.status === 'idle');
    if (idleElevators.length > 0) {
      return this.findClosestElevator(fromFloor, idleElevators);
    }
    
    // 3. Any available elevator closest to origin floor
    return this.findClosestElevator(fromFloor, candidates);
  }

  private findClosestElevator(targetFloor: number, elevators: Elevator[]): Elevator {
    return elevators.reduce((best, elev) => {
      if (!best) return elev;
      const dist = Math.abs(elev.currentFloor - targetFloor);
      const bestDist = Math.abs(best.currentFloor - targetFloor);
      return dist < bestDist ? elev : best;
    }, null as Elevator | null)!;
  }

  /**
   * Validate a floor request for feasibility.
   */
  validateFloorRequest(fromFloor: number, toFloor: number, elevators: Elevator[]): {
    valid: boolean;
    message?: string;
    elevator?: Elevator;
  } {
    if (fromFloor === toFloor) {
      return { valid: false, message: 'You are already on floor ' + fromFloor };
    }
    
    const elev = this.findOptimalElevator(fromFloor, toFloor, elevators);
    if (!elev) {
      return { valid: false, message: `No elevator can go from floor ${fromFloor} to ${toFloor}` };
    }
    
    return { valid: true, elevator: elev, message: `Elevator ${elev.name} will serve this request` };
  }
}
```

## 🎭 Animation and Interaction Details

### Elevator Movement Animation

```css
/* animations.css */
.elevator-car {
  transition: transform 0.8s ease-in-out;  /* smooth movement between floors */
  will-change: transform;
}

.elevator-car.moving {
  transition-duration: 1.2s; /* slower for visible movement */
}

.doors {
  position: relative;
  overflow: hidden;
}

.door-left, .door-right {
  width: 50%; height: 100%;
  background: #666;
  transition: transform 0.3s ease-in-out;
  position: absolute; top: 0;
}

.door-left { left: 0; }
.door-right { right: 0; }

.doors.open .door-left {
  transform: translateX(-100%); /* slide left door out of view */
}

.doors.open .door-right {
  transform: translateX(100%);  /* slide right door out */
}

/* Call Button hover/active feedback */
.call-button {
  transition: transform 0.2s ease, box-shadow 0.2s;
}

.call-button:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.call-button:active, .call-button.active {
  transform: scale(0.98);
}

.call-button.active {
  background: var(--primary-500);
  color: white;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Respect user motion preferences */
@media (prefers-reduced-motion: reduce) {
  .elevator-car, .door-left, .door-right, .call-button {
    transition: none !important;
    animation: none !important;
  }
}
```

### Loading States and Transitions

```svelte
<!-- LoadingSpinner.svelte -->
<script lang="ts">
  export let size: 'small' | 'medium' | 'large' = 'medium';
  export let color: 'primary' | 'secondary' = 'primary';
</script>

<div class="loading-spinner {size} {color}">
  <div class="spinner"></div>
</div>

<style>
  .loading-spinner {
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .spinner {
    border: 4px solid #f3f3f3;
    border-top: 4px solid var(--primary-600);
    border-radius: 50%;
    width: 40px; height: 40px;
    animation: spin 1s linear infinite;
  }

  .small .spinner { width: 20px; height: 20px; border-width: 2px; }
  .large .spinner { width: 60px; height: 60px; border-width: 6px; }

  @keyframes spin { 
    0% { transform: rotate(0deg);} 
    100% { transform: rotate(360deg);} 
  }
</style>
```

## 📱 Responsive Design

### Breakpoint Strategy

```css
/* Mobile First Approach */
.elevator-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: 1fr;
  padding: var(--space-4);
}

/* Tablet */
@media (min-width: 768px) {
  .elevator-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-6);
    padding: var(--space-6);
  }
}

/* Desktop */
@media (min-width: 1024px) {
  .elevator-grid {
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: var(--space-8);
    padding: var(--space-8);
  }
}

/* Large Desktop */
@media (min-width: 1536px) {
  .elevator-grid {
    max-width: 1400px;
    margin: 0 auto;
  }
}
```

### Mobile Optimizations
- **Touch-friendly Controls**: Minimum 44px touch targets
- **Simplified Navigation**: Collapsible sidebar for controls
- **Optimized Animations**: Reduced motion for battery conservation
- **Progressive Enhancement**: Core functionality without JavaScript

## ♿ Accessibility Features

### ARIA Implementation

```svelte
<!-- ElevatorBuilding.svelte -->
<section 
  role="region"
  aria-labelledby="elevator-{elevator.name}-title"
  aria-describedby="elevator-{elevator.name}-status">
  
  <h3 id="elevator-{elevator.name}-title">
    Elevator {elevator.name}
  </h3>
  
  <div id="elevator-{elevator.name}-status" class="sr-only" aria-live="polite">
    Elevator {elevator.name} is on floor {elevator.currentFloor}, {elevator.status}.
  </div>
  
  <div class="elevator-shaft" role="img" aria-label="Elevator shaft visualization">
    <!-- Elevator content -->
  </div>
</section>
```

### Keyboard Navigation

```typescript
// keyboard.ts - Keyboard navigation handler
export function handleKeyboardNavigation(event: KeyboardEvent, elevator: Elevator) {
  const { key, target } = event;
  
  switch (key) {
    case 'ArrowUp':
      event.preventDefault();
      if (elevator.currentFloor < elevator.maxFloor) {
        requestFloor(elevator.currentFloor, elevator.currentFloor + 1);
      }
      break;
      
    case 'ArrowDown':
      event.preventDefault();
      if (elevator.currentFloor > elevator.minFloor) {
        requestFloor(elevator.currentFloor, elevator.currentFloor - 1);
      }
      break;
      
    case 'Enter':
    case ' ':
      event.preventDefault();
      (target as HTMLElement).click();
      break;
  }
}
```

### Screen Reader Support
- **Live Regions**: Dynamic content announcements for elevator status changes
- **Descriptive Labels**: Comprehensive ARIA labels for all interactive elements
- **Focus Management**: Logical tab order and focus trapping for modals
- **Alternative Text**: Meaningful descriptions for visual elements

## 🚀 Performance Optimization

### Code Splitting & Lazy Loading

```typescript
// Lazy load heavy components
const MonitoringDashboard = await import('./components/monitoring/MonitoringDashboard.svelte');
const CreateElevatorModal = await import('./components/controls/CreateElevatorModal.svelte');
```

### Bundle Size Optimization
- **Target**: < 100KB initial bundle
- **Tree Shaking**: Eliminate unused code
- **Compression**: Gzip/Brotli compression
- **Asset Optimization**: Compressed images and fonts

### Runtime Performance
- **Virtual Scrolling**: For large numbers of floors
- **Debounced Updates**: Prevent excessive re-renders
- **Memoization**: Cache expensive calculations
- **Animation Optimization**: Use CSS transforms and will-change

### Service Worker Implementation

```typescript
// service-worker.ts
const CACHE_NAME = 'elevator-v1';
const urlsToCache = [
  '/',
  '/static/js/bundle.js',
  '/static/css/main.css',
  '/static/media/logo.svg'
];

self.addEventListener('install', (event: any) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(urlsToCache))
  );
});

self.addEventListener('fetch', (event: any) => {
  event.respondWith(
    caches.match(event.request)
      .then((response) => response || fetch(event.request))
  );
});
```

## 🧪 Testing Strategy

### Component Testing

```typescript
// tests/components/ElevatorCar.test.ts
import { render, fireEvent } from '@testing-library/svelte';
import ElevatorCar from '../src/components/elevator/ElevatorCar.svelte';

describe('ElevatorCar', () => {
  test('renders with correct floor position', () => {
    const props = {
      elevator: {
        name: 'Test-Elevator',
        currentFloor: 5,
        status: 'idle' as const,
        direction: null
      }
    };
    
    const { getByTestId } = render(ElevatorCar, { props });
    const car = getByTestId('elevator-car');
    
    expect(car).toHaveStyle('transform: translateY(-200px)');
  });

  test('shows direction indicator when moving', () => {
    const props = {
      elevator: {
        name: 'Test-Elevator',
        currentFloor: 3,
        status: 'moving' as const,
        direction: 'up' as const
      }
    };
    
    const { getByText } = render(ElevatorCar, { props });
    expect(getByText('↑')).toBeInTheDocument();
  });
});
```

### Integration Testing

```typescript
// tests/integration/websocket.test.ts
import { WebSocketService } from '../src/services/websocket';
import { elevators } from '../src/stores/elevators';

describe('WebSocket Integration', () => {
  test('updates elevator positions on status message', () => {
    const mockStatus = {
      elevators: {
        'Elevator-1': {
          currentFloor: 7,
          status: 'moving',
          direction: 'up'
        }
      }
    };
    
    const ws = new WebSocketService();
    ws.handleStatusUpdate(mockStatus);
    
    expect(get(elevators)).toContainEqual(
      expect.objectContaining({
        name: 'Elevator-1',
        currentFloor: 7,
        status: 'moving'
      })
    );
  });
});
```

### Accessibility Testing

```typescript
// tests/accessibility/a11y.test.ts
import { axe, toHaveNoViolations } from 'jest-axe';
import { render } from '@testing-library/svelte';
import App from '../src/App.svelte';

expect.extend(toHaveNoViolations);

describe('Accessibility', () => {
  test('should not have accessibility violations', async () => {
    const { container } = render(App);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });
});
```

## 🔧 Developer Experience

### TypeScript Configuration

```json
// tsconfig.json
{
  "extends": "@tsconfig/svelte/tsconfig.json",
  "compilerOptions": {
    "target": "ESNext",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ESNext", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src/**/*.d.ts", "src/**/*.ts", "src/**/*.js", "src/**/*.svelte"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

### ESLint Configuration

```javascript
// .eslintrc.js
module.exports = {
  root: true,
  extends: [
    'eslint:recommended',
    '@typescript-eslint/recommended',
    'plugin:svelte/recommended'
  ],
  parser: '@typescript-eslint/parser',
  plugins: ['@typescript-eslint'],
  parserOptions: {
    sourceType: 'module',
    ecmaVersion: 2020,
    extraFileExtensions: ['.svelte']
  },
  env: {
    browser: true,
    es2017: true,
    node: true
  },
  overrides: [
    {
      files: ['*.svelte'],
      parser: 'svelte-eslint-parser',
      parserOptions: {
        parser: '@typescript-eslint/parser'
      }
    }
  ]
};
```

### Prettier Configuration

```json
// .prettierrc
{
  "useTabs": true,
  "singleQuote": true,
  "trailingComma": "es5",
  "printWidth": 100,
  "plugins": ["prettier-plugin-svelte"],
  "overrides": [
    {
      "files": "*.svelte",
      "options": {
        "parser": "svelte"
      }
    }
  ]
}
```

## 🚀 Deployment Strategy

### GitHub Pages Setup

```yaml
# .github/workflows/deploy.yml
name: Deploy to GitHub Pages

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout
      uses: actions/checkout@v3
      
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: 'client/package-lock.json'
        
    - name: Install dependencies
      working-directory: ./client
      run: npm ci
      
    - name: Run tests
      working-directory: ./client
      run: npm run test
      
    - name: Run linting
      working-directory: ./client
      run: npm run lint
      
    - name: Build
      working-directory: ./client
      run: npm run build
      
    - name: Deploy to GitHub Pages
      uses: peaceiris/actions-gh-pages@v3
      if: github.ref == 'refs/heads/main'
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: ./client/dist
```

### Build Configuration

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  base: '/elevator/', // GitHub Pages subdirectory
  build: {
    outDir: 'dist',
    minify: 'esbuild',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['svelte']
        }
      }
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:6660',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
});
```

## 🎯 Optional Feature Suggestions

### Progressive Web App (PWA) Support
- **Web App Manifest**: App name, icons, theme colors
- **Service Worker**: Offline caching and app-like experience
- **Install Prompt**: "Add to Home Screen" functionality
- **Offline Mode**: Demo mode when no connection available

### Lightweight Analytics
- **Privacy-conscious**: Minimal data collection
- **Usage Tracking**: Feature interaction analytics
- **Performance Monitoring**: Core Web Vitals tracking
- **Error Tracking**: Runtime error collection

### 3D Visualization Mode
- **Optional Feature**: Toggle for immersive view
- **CSS 3D Transforms**: Simple 3D perspective
- **Lazy Loading**: Only load when activated
- **Fallback Support**: Graceful degradation to 2D

### Enhanced Monitoring & Stats
- **Interactive Charts**: SVG-based performance charts
- **Historical Data**: Request replay functionality
- **Real-time Metrics**: Live system performance
- **Export Functionality**: Data export capabilities

### Offline Simulation / Demo Mode
- **Client-side Simulation**: Elevator movement simulation
- **Preset Configurations**: Sample elevator setups
- **Random Movements**: Dynamic demo behavior
- **Clear Indicators**: "Demo mode" notifications

### Multilingual Support
- **i18n Implementation**: Translation system
- **Language Detection**: Browser language detection
- **RTL Support**: Right-to-left language support
- **Cultural Adaptation**: Localized formatting

## 📁 Project Structure

```
client/
├── public/
│   ├── index.html
│   ├── favicon.ico
│   ├── manifest.json
│   └── robots.txt
├── src/
│   ├── components/
│   │   ├── layout/
│   │   │   ├── Header.svelte
│   │   │   ├── Footer.svelte
│   │   │   └── Sidebar.svelte
│   │   ├── elevator/
│   │   │   ├── ElevatorBuilding.svelte
│   │   │   ├── ElevatorCar.svelte
│   │   │   ├── ElevatorShaft.svelte
│   │   │   ├── FloorRow.svelte
│   │   │   ├── CallButton.svelte
│   │   │   └── FloorTouchDisplay.svelte
│   │   ├── controls/
│   │   │   ├── CreateElevatorModal.svelte
│   │   │   ├── ElevatorControlPanel.svelte
│   │   │   └── SystemControls.svelte
│   │   ├── monitoring/
│   │   │   ├── MetricsPanel.svelte
│   │   │   ├── HealthStatus.svelte
│   │   │   └── ConnectionStatus.svelte
│   │   └── common/
│   │       ├── Button.svelte
│   │       ├── Modal.svelte
│   │       ├── LoadingSpinner.svelte
│   │       └── Toast.svelte
│   ├── stores/
│   │   ├── elevators.ts
│   │   ├── websocket.ts
│   │   └── ui.ts
│   ├── services/
│   │   ├── api.ts
│   │   ├── websocket.ts
│   │   └── validation.ts
│   ├── utils/
│   │   ├── animations.ts
│   │   ├── calculations.ts
│   │   ├── accessibility.ts
│   │   └── floorSelection.ts
│   ├── styles/
│   │   ├── global.css
│   │   ├── components.css
│   │   └── animations.css
│   ├── types/
│   │   └── index.ts
│   ├── App.svelte
│   └── main.ts
├── tests/
│   ├── components/
│   ├── integration/
│   ├── accessibility/
│   └── utils/
├── package.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json
├── .eslintrc.js
├── .prettierrc
└── README.md
```

## 🚀 Implementation Roadmap

### Phase 1: Foundation (Weeks 1-2)
- [ ] Project setup with Svelte/SvelteKit and TypeScript
- [ ] Basic component structure and routing
- [ ] Tailwind CSS integration and design system
- [ ] WebSocket service implementation
- [ ] Basic elevator visualization

### Phase 2: Core Features (Weeks 3-4)
- [ ] Elevator creation workflow with validation
- [ ] Real-time status updates and animations
- [ ] Floor request functionality and smart selection
- [ ] Touch display system implementation
- [ ] Responsive design and mobile optimization

### Phase 3: Enhancement (Weeks 5-6)
- [ ] Monitoring dashboard and metrics
- [ ] Accessibility features and testing
- [ ] Error handling and user feedback
- [ ] Performance optimization and testing
- [ ] PWA implementation

### Phase 4: Deployment (Week 7)
- [ ] GitHub Actions CI/CD setup
- [ ] Production build optimization
- [ ] Documentation completion
- [ ] Final testing and bug fixes
- [ ] GitHub Pages deployment

## 📚 Documentation Plan

### User Documentation
- **Getting Started Guide**: Quick setup and basic usage
- **Feature Documentation**: Comprehensive feature explanations
- **API Reference**: Frontend API and configuration options
- **Troubleshooting Guide**: Common issues and solutions

### Developer Documentation
- **Architecture Guide**: Detailed system architecture
- **Component API**: Component props and events documentation
- **Contributing Guide**: Development setup and contribution guidelines
- **Deployment Guide**: Production deployment instructions

---

This updated frontend design specification incorporates modern best practices for UX design, performance optimization, accessibility compliance, developer experience, and optional enhancements. The design provides a comprehensive blueprint for building a production-ready, accessible, and performant elevator control system frontend that meets all specified requirements while maintaining excellent user experience and developer productivity. 