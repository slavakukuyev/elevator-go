# Elevator Control System - Client Side Documentation

## 🏗️ Overview

The client-side application is a modern, responsive web interface for the Elevator Control System built with SvelteKit, TypeScript, and Tailwind CSS. It provides real-time visualization of elevator operations, intuitive controls for floor selection, and comprehensive monitoring capabilities.

## Project custom rules
* always use the same port 5173 for web page appliaction.
* always run linters before running npm run dev. 

## 🚀 Technology Stack

### Core Framework
- **SvelteKit with TypeScript**: Modern, compile-time optimized framework with type safety
- **Tailwind CSS**: Utility-first CSS framework for rapid, consistent styling
- **Vite**: Lightning-fast build tool with hot module replacement

### Key Libraries
- **Svelte Stores**: Built-in state management for reactive data
- **Native WebSocket**: Real-time communication with the backend
- **Vitest**: Modern testing framework for unit tests

### Development Tools
- **ESLint**: Code linting with TypeScript and Svelte support
- **Prettier**: Code formatting with Svelte plugin
- **GitHub Actions**: Automated CI/CD pipeline for GitHub Pages deployment

## 📐 Architecture

### Component Hierarchy

```
App (routes/+layout.svelte)
├── Header (Header.svelte)
├── ElevatorControlPanel (ElevatorControlPanel.svelte)
│   └── CreateElevatorModal (CreateElevatorModal.svelte)
├── ElevatorBuildingGrid (ElevatorBuildingGrid.svelte)
│   └── ElevatorBuilding (ElevatorBuilding.svelte)
├── MonitoringDashboard (MonitoringDashboard.svelte)
├── Footer (Footer.svelte)
└── Toast (Toast.svelte)
```

### State Management

The application uses Svelte stores for centralized state management:

```typescript
// Primary stores
export const elevators = writable<Elevator[]>([]);
export const systemStatus = writable<SystemStatus>({ healthy: true, elevatorCount: 0 });
export const connectionStatus = writable<ConnectionStatus>({ connected: false, retryCount: 0 });

// Derived stores for performance
export const availableElevators = derived([currentFloor, elevators], ...);
export const idleElevators = derived(elevators, ...);
export const elevatorUtilization = derived(elevators, ...);
```

### Service Layer

- **API Service** (`services/api.ts`): HTTP client for REST API communication
- **WebSocket Service** (`services/websocket.ts`): Real-time data synchronization
- **Floor Selection Service** (`utils/floorSelection.ts`): Elevator assignment algorithms
- **Validation Service** (`utils/validation.ts`): Form and input validation

## 🎨 User Interface Features

### 1. Header Section
- **Brand Identity**: Elevator Control System logo and title
- **Connection Status**: Real-time WebSocket connection indicator
- **Theme Toggle**: Light/dark mode switcher with system preference detection
- **Monitoring Panel Toggle**: Show/hide system monitoring dashboard

### 2. Control Panel (Left Sidebar)
- **System Status Summary**: Overall health and elevator count
- **Create Elevator**: Modal form for adding new elevators with validation
- **Floor Controls**: Current floor selector and call elevator buttons
- **Quick Access**: Common floor shortcuts (Ground, 1, 2, 5, 10)

### 3. Elevator Visualization (Main Area)
- **Grid Layout**: Responsive grid showing all elevators
- **Individual Buildings**: Each elevator displayed as a vertical shaft
- **Real-time Animation**: Smooth elevator movement with CSS transitions
- **Status Indicators**: Color-coded status (idle, moving, error)
- **Call Buttons**: Up/down buttons for each floor
- **Door Animation**: Visual feedback for door open/close states

### 4. Monitoring Dashboard (Right Sidebar)
- **Connection Status**: WebSocket health and retry information
- **System Health**: Overall status and operational metrics
- **Utilization Charts**: Visual progress bars for elevator usage
- **Pending Requests**: Real-time queue of floor requests
- **Individual Details**: Per-elevator status and configuration

### 5. Interactive Elements
- **Smooth Animations**: CSS transitions with reduced motion support
- **Toast Notifications**: Success/error feedback with auto-dismiss
- **Responsive Design**: Mobile-first approach with breakpoint adaptations
- **Accessibility**: WCAG 2.1 AA compliant with keyboard navigation

## 🔄 Real-time Data Flow

### WebSocket Integration

The application maintains a persistent WebSocket connection for real-time updates:

```typescript
// Auto-connect on service import
wsService.connect();

// Handle different message types
private handleMessage(message: WebSocketMessage) {
  switch (message.type) {
    case 'status': this.handleStatusUpdate(message.payload); break;
    case 'elevator_update': this.handleElevatorUpdate(message.payload); break;
    case 'floor_request': this.handleFloorRequest(message.payload); break;
    case 'system_alert': this.handleSystemAlert(message.payload); break;
  }
}
```

### State Synchronization

- **Automatic Reconnection**: Exponential backoff strategy for connection recovery
- **Ping/Pong**: Keep-alive mechanism to detect connection issues
- **Store Updates**: Real-time updates to Svelte stores trigger UI re-renders
- **Error Handling**: Graceful degradation when connection is lost

## 🛠️ Development Workflow

### Project Structure

```
client/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── common/         # Generic components (Button, Modal, Toast)
│   │   ├── layout/         # Layout components (Header, Footer)
│   │   ├── controls/       # Control panel components
│   │   ├── elevator/       # Elevator visualization components
│   │   └── monitoring/     # Monitoring dashboard components
│   ├── routes/             # SvelteKit pages and layouts
│   ├── stores/             # Svelte stores for state management
│   ├── services/           # API and WebSocket services
│   ├── utils/              # Utility functions and algorithms
│   ├── types/              # TypeScript type definitions
│   └── styles/             # Global CSS and Tailwind imports
├── static/                 # Static assets (favicon, manifest)
├── tests/                  # Test files and setup
└── docs/                   # Documentation files
```

### Development Commands

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run tests
npm run test

# Run linting
npm run lint

# Format code
npm run format

# Type check
npm run check
```

### Build Configuration

The application is configured for static site generation to support GitHub Pages deployment:

```javascript
// svelte.config.js
adapter: adapter({
  pages: 'dist',
  assets: 'dist',
  fallback: 'index.html',
  precompress: false,
  strict: true
})
```

## 🎯 Key Features

### 1. Elevator Management
- **Create Elevators**: Form-based creation with comprehensive validation
- **Real-time Status**: Live updates of elevator position, direction, and status
- **Visual Representation**: Intuitive shaft visualization with animated cars
- **Floor Range Support**: Configurable min/max floors including basement levels

### 2. Floor Control
- **Call Elevators**: Up/down call buttons for each floor
- **Smart Selection**: Optimal elevator assignment algorithm
- **Quick Access**: Shortcuts to common floors
- **Visual Feedback**: Clear indication of active requests and responses

### 3. System Monitoring
- **Health Dashboard**: Real-time system status and metrics
- **Connection Monitoring**: WebSocket status with retry information
- **Utilization Tracking**: Visual charts showing elevator usage
- **Request Queue**: Live view of pending floor requests

### 4. User Experience
- **Responsive Design**: Works seamlessly on desktop, tablet, and mobile
- **Theme Support**: Light/dark mode with system preference detection
- **Accessibility**: Full keyboard navigation and screen reader support
- **Progressive Enhancement**: Core functionality works without JavaScript

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the client directory:

```env
VITE_API_URL=http://localhost:6660/api/v1
VITE_WS_URL=ws://localhost:6660/ws/status
BASE_PATH=/elevator
```

### Tailwind Configuration

Custom design system configured in `tailwind.config.js`:

```javascript
theme: {
  extend: {
    colors: {
      primary: { /* Blue color palette */ },
      secondary: { /* Gray color palette */ }
    },
    animation: {
      'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite'
    }
  }
}
```

## 🧪 Testing

### Test Setup

- **Vitest**: Modern testing framework with TypeScript support
- **Testing Library**: Component testing utilities for Svelte
- **Jest DOM**: Additional DOM matchers for assertions
- **Mock Services**: WebSocket and API mocking for isolated tests

### Running Tests

```bash
# Run all tests
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

### Test Examples

```typescript
// Component test
test('renders elevator with correct floor position', () => {
  const { getByTestId } = render(ElevatorCar, { 
    props: { elevator: { currentFloor: 5, status: 'idle' } }
  });
  const car = getByTestId('elevator-car');
  expect(car).toHaveStyle('transform: translateY(-200px)');
});

// Service test
test('validates elevator configuration', () => {
  const result = validationService.validateElevatorConfig({
    name: 'Test', minFloor: 0, maxFloor: 10
  });
  expect(result.valid).toBe(true);
});
```

## 🚀 Deployment

### GitHub Pages Setup

The application is configured for automatic deployment to GitHub Pages:

```yaml
# .github/workflows/deploy.yml
- name: Build
  working-directory: ./client
  run: npm run build

- name: Deploy to GitHub Pages
  uses: peaceiris/actions-gh-pages@v3
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    publish_dir: ./client/dist
```

### Production Build

```bash
# Build for production
npm run build

# The dist/ directory contains the static files ready for deployment
```

## 🔍 Performance

### Optimization Features

- **Code Splitting**: Automatic chunking for optimal loading
- **Tree Shaking**: Eliminates unused code from the bundle
- **Asset Optimization**: Compressed images and fonts
- **Service Worker**: Caching for offline functionality (optional)

### Bundle Analysis

- **Target**: < 100KB initial bundle
- **Actual**: Check with `npm run build` and analyze the dist/ directory
- **Core Web Vitals**: Optimized for LCP, FID, and CLS metrics

## 📱 Accessibility

### WCAG 2.1 AA Compliance

- **Keyboard Navigation**: Full functionality without mouse
- **Screen Reader Support**: Semantic HTML and ARIA labels
- **Color Contrast**: Sufficient contrast ratios in all themes
- **Focus Management**: Logical tab order and visible focus indicators
- **Motion Preferences**: Respects `prefers-reduced-motion` setting

### Testing Accessibility

```bash
# Install and run axe-core for accessibility testing
npm install --save-dev @axe-core/cli
npx axe http://localhost:4173
```

## 🔮 Future Enhancements

### Planned Features

1. **Progressive Web App**: Offline functionality and app-like experience
2. **3D Visualization**: Optional immersive elevator view
3. **Historical Analytics**: Request patterns and performance metrics
4. **Multi-language Support**: Internationalization (i18n)
5. **Advanced Monitoring**: Real-time charts and alerts

### Technical Improvements

1. **Virtual Scrolling**: For buildings with many floors
2. **Enhanced Animations**: Physics-based elevator movement
3. **Improved Caching**: Better offline experience
4. **Performance Monitoring**: Real user metrics collection

## 📞 API Integration

### REST Endpoints

```typescript
// Create elevator
POST /api/v1/elevators
{ name: "Main-A", minFloor: 0, maxFloor: 10 }

// Request floor
POST /api/v1/floors/request
{ from: 0, to: 5 }

// Call elevator
POST /api/v1/elevators/call
{ floor: 3, direction: "up" }

// Get status
GET /api/v1/status
```

### WebSocket Messages

```typescript
// Status update
{ type: "status", payload: { elevators: {...}, system: {...} } }

// Elevator update
{ type: "elevator_update", payload: { name: "Main-A", currentFloor: 3 } }

// Floor request
{ type: "floor_request", payload: { from: 0, to: 5, status: "assigned" } }
```

## 💡 Best Practices

### Code Organization

- **Single Responsibility**: Each component has a clear, focused purpose
- **Type Safety**: Comprehensive TypeScript coverage
- **Error Boundaries**: Graceful error handling throughout the app
- **Performance**: Derived stores and reactive patterns for efficiency

### Styling Guidelines

- **Utility-First**: Prefer Tailwind utilities over custom CSS
- **Consistent Spacing**: Use design system spacing values
- **Responsive Design**: Mobile-first breakpoint strategy
- **Dark Mode**: Support for both light and dark themes

### Testing Strategy

- **Unit Tests**: Individual component and utility testing
- **Integration Tests**: Service and store interaction testing
- **Accessibility Tests**: Automated a11y validation
- **Visual Regression**: Screenshot comparison (future)

## 🐛 Troubleshooting

### Common Issues

1. **WebSocket Connection Failed**
   - Check if backend server is running
   - Verify WebSocket URL in environment variables
   - Check browser network tab for connection errors

2. **Build Errors**
   - Clear node_modules and reinstall dependencies
   - Check TypeScript errors with `npm run check`
   - Verify all imports are correctly typed

3. **Styling Issues**
   - Check Tailwind CSS is properly imported
   - Verify custom CSS doesn't conflict with utilities
   - Test in different browsers and devices

### Debug Tools

```typescript
// Enable debug logging for WebSocket
localStorage.setItem('debug', 'websocket');

// Check store states in browser console
import { get } from 'svelte/store';
import { elevators } from './stores/elevators';
console.log(get(elevators));
```

## 📄 License

This project is part of the Elevator Control System and follows the same licensing terms as the main project.

---

For development instructions and advanced configuration, see [client_side_development.md](./client_side_development.md). 