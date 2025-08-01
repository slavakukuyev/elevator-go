# Client-Side API Refactoring

## Overview
This document outlines the refactoring changes made to the client-side code to properly integrate with the backend API without modifying the backend.

## Issues Fixed

### 1. API Base URL Mismatch
**Problem**: Client was using `/api/v1` but server routes are at `/v1`
**Solution**: Updated `API_BASE` in `client/src/services/api.ts` from:
```typescript
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:6660/api/v1';
```
to:
```typescript
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:6660/v1';
```

### 2. Response Structure Mismatches
**Problem**: Backend and client had different field naming conventions and response structures
**Solution**: Added transformation functions to map backend responses to client types:

- `transformElevatorResponse()` - Maps `BackendElevatorCreateResponse` to `Elevator`
- `transformFloorRequestResponse()` - Maps `BackendFloorRequestResponse` to `FloorRequest`
- `transformHealthResponse()` - Maps `BackendHealthResponse` to `SystemStatus`
- `transformMetricsResponse()` - Maps `BackendMetricsResponse` to `MetricsData`

### 3. Non-Existent Endpoints
**Problem**: Client was calling endpoints that don't exist in the backend API
**Solution**: Updated or removed calls to non-existent endpoints:

#### Implemented Endpoints (Working):
- `POST /v1/elevators` - Create elevator ✅
- `POST /v1/floors/request` - Request floor service ✅
- `GET /v1/health` - Health check ✅
- `GET /v1/metrics` - System metrics ✅

#### Non-Implemented Endpoints (Handled):
- `GET /v1/elevators` - Get all elevators (returns empty array)
- `GET /v1/elevators/{name}` - Get specific elevator (throws error)
- `DELETE /v1/elevators/{name}` - Delete elevator (throws error)
- `GET /v1/floors/requests` - Get floor requests (returns empty array)
- `POST /v1/elevators/call` - Call elevator (uses requestFloor instead)
- `GET /v1/status` - Get system status (throws error)
- Emergency and maintenance endpoints (throw errors)

### 4. Data Structure Updates
**Problem**: Client types didn't match backend expectations
**Solution**: 
- Removed `capacity` field from `ElevatorConfig` (not supported by backend)
- Updated form components to not send unsupported fields
- Added proper field mapping between snake_case (backend) and camelCase (client)

## Files Modified

### Core API Service
- `client/src/services/api.ts` - Complete refactor with proper response handling

### Type Definitions
- `client/src/types/index.ts` - Removed unsupported `capacity` field

### Components
- `client/src/components/controls/CreateElevatorModal.svelte` - Removed capacity field
- `client/src/components/elevator/ElevatorBuilding.svelte` - Updated to use requestFloor
- `client/src/components/controls/ElevatorControlPanel.svelte` - Updated to use requestFloor
- `client/src/routes/+layout.svelte` - Removed getElevators call

### Configuration
- `client/.env.local` - Added proper API URL configuration

## Backend API Endpoints

### Available Endpoints
```
POST /v1/elevators
  Body: { "name": "string", "min_floor": number, "max_floor": number }
  Response: { "name": "string", "min_floor": number, "max_floor": number, "message": "string" }

POST /v1/floors/request
  Body: { "from": number, "to": number }
  Response: { "elevator_name": "string", "from_floor": number, "to_floor": number, "direction": "string", "message": "string" }

GET /v1/health
  Response: { "status": "string", "timestamp": "string", "checks": object }

GET /v1/metrics
  Response: { "timestamp": "string", "metrics": object }

GET /v1
  Response: API information and available endpoints
```

## Testing the Integration

1. **Start the backend server**:
   ```bash
   cd /path/to/elevator
   go run cmd/server/main.go
   ```

2. **Start the client**:
   ```bash
   cd client
   npm run dev
   ```

3. **Test elevator creation**:
   - Open the client in browser
   - Click "Add Elevator"
   - Fill in name, min floor, max floor
   - Submit - should work without 404 errors

4. **Test floor requests**:
   - Use the control panel to request floor service
   - Should work without errors

## Future Enhancements

If additional backend endpoints are needed, they should be implemented in the backend first, then the client can be updated to use them. The current refactoring provides a solid foundation for future API expansion. 