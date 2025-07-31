# Elevator Algorithm Documentation

## Overview

The elevator system implements a **SCAN/LOOK algorithm** optimized for real-world scenarios including high-rise buildings with underground parking. Each elevator operates independently with its own request management system.

## Core Algorithm: The `Run()` Function

The `Run()` function is the heart of the elevator's movement logic. It processes requests based on the current direction and implements intelligent direction switching to optimize passenger service.

### Algorithm Flow

```
1. Get current position and direction
2. Wait for floor transition duration
3. Process requests in current direction
4. Handle direction changes when needed
5. Manage idle state when no requests exist
```

### Detailed Logic Breakdown

#### Phase 1: Moving Up with Up Requests
**Condition**: `direction == DirectionUp && upRequests > 0`

**Behavior**:
- **Service Current Floor**: If passengers requested this floor, open doors, flush requests, close doors
- **Boundary Check**: If reached top floor, change direction to down (prevents infinite upward movement)
- **Continue Movement**: If more up requests exist above current floor, move up one floor

**Real-world Application**: Handles normal upward traffic efficiently, serving floors in sequence.

#### Phase 2: Moving Down with Down Requests  
**Condition**: `direction == DirectionDown && downRequests > 0`

**Behavior**:
- **Service Current Floor**: If passengers requested this floor, open doors, flush requests, close doors
- **Boundary Check**: If reached bottom floor, change direction to up
- **Continue Movement**: If more down requests exist below current floor, move down one floor

**Real-world Application**: Handles downward traffic, including underground parking scenarios.

#### Phase 3: Direction Change - Down to Up Transition
**Condition**: `direction == DirectionDown && upRequests > 0 && noDownRequests`

**Smart Logic**:
- **Find Lowest Up Request**: Locate the smallest floor number in up requests
- **Continue Down if Needed**: If that floor is below current position, continue moving down to reach it
- **Change Direction**: If already at the pickup floor, switch to up direction immediately

**Real-world Application**: Efficiently handles passengers waiting on lower floors when elevator is moving down.

#### Phase 4: Direction Change - Up to Down Transition
**Condition**: `direction == DirectionUp && downRequests > 0 && noUpRequests`

**Smart Logic**:
- **Find Highest Down Request**: Locate the largest floor number in down requests  
- **Continue Up if Needed**: If that floor is above current position, continue moving up to reach it
- **Change Direction**: If already at the pickup floor, switch to down direction immediately

**Real-world Application**: Handles passengers waiting on upper floors when elevator is moving up.

#### Phase 5: Overshot Recovery - Up Direction
**Condition**: `direction == DirectionUp && upRequests > 0 && allUpRequestsBelow`

**Recovery Logic**:
- **Detect Overshoot**: Elevator moved past all up requests
- **Change Direction**: Switch to down to return and serve missed requests

**Real-world Application**: Prevents passengers from being stranded when elevator overshoots their floor.

#### Phase 6: Overshot Recovery - Down Direction  
**Condition**: `direction == DirectionDown && downRequests > 0 && allDownRequestsAbove`

**Recovery Logic**:
- **Detect Overshoot**: Elevator moved past all down requests
- **Change Direction**: Switch to up to return and serve missed requests

**Real-world Application**: Handles cases where elevator moved too far down past passenger floors.

#### Phase 7: Idle State Management
**Condition**: `noUpRequests && noDownRequests`

**Behavior**:
- **Enter Idle State**: Set direction to `DirectionIdle`
- **Log Status**: Record current floor and idle state for monitoring

**Real-world Application**: Energy efficient - elevator stops moving when no requests exist.

## Key Algorithm Benefits

### 1. **LOOK Algorithm Implementation**
- Continues in one direction until all requests in that direction are served
- Reverses direction only when necessary
- More efficient than pure SCAN (doesn't go to building boundaries unnecessarily)

### 2. **Smart Direction Changes**
- **Predictive Movement**: Moves toward pickup floors even when changing direction
- **Overshoot Recovery**: Handles cases where elevator moves past requested floors
- **Boundary Awareness**: Respects building limits (top/bottom floors)

### 3. **Multi-Range Support**
- **Different Floor Ranges**: Supports elevators with different min/max floors
- **Underground Parking**: Handles negative floor numbers efficiently
- **Express Zones**: Can serve specific floor ranges (e.g., floors 20-40 only)

## Performance Characteristics

### Time Complexity
- **Per Request**: O(1) for processing each movement step
- **Direction Change**: O(n) where n is number of pending requests (for finding min/max floors)

### Space Complexity
- **Request Storage**: O(r) where r is number of active requests
- **State Management**: O(1) constant space for elevator state

## Integration with Building Systems

### High-Rise Buildings
- **Zone Separation**: Multiple elevators can serve different floor ranges
- **Peak Traffic**: Algorithm handles rush hour patterns efficiently
- **Express Service**: Can skip floors without requests

### Underground Parking
- **Negative Floors**: Supports basement levels (-1, -2, -3, etc.)
- **Ground Floor Hub**: Efficiently handles basement ↔ ground ↔ upper floor traffic
- **Parking Rush**: Optimized for common parking exit patterns

## Safety Features

### 1. **Boundary Protection**
- Prevents movement beyond building limits
- Automatic direction reversal at top/bottom floors

### 2. **Request Validation**
- All requests validated against elevator's floor range
- Invalid requests rejected before processing

### 3. **Context Cancellation**
- Respects context cancellation for graceful shutdown
- Prevents blocked operations during system shutdown

## Edge Cases

### Same-Floor Request Processing
**Scenario**: Idle elevator on floor N receives request from floor N to floor M

**Behavior**:
1. **Immediate Direction Change**: Elevator switches from `DirectionIdle` to appropriate direction (Up/Down)
2. **Instant Pickup**: Door operations execute immediately since elevator is already at pickup floor
3. **Request Storage**: `directionsManager.Append()` creates entry `up[N] = [M]` or `down[N] = [M]`
4. **Flush on Pickup**: `Flush()` processes pickup by:
   - Creating destination entry: `up[M] = []` (empty slice)
   - Removing pickup entry: `delete(up, N)`
5. **Movement to Destination**: Elevator travels to floor M using normal algorithm
6. **Completion**: Second flush at destination removes empty slice, returns to idle

**Technical Note**: The `Flush()` operation creates empty destination slices which don't count toward `DirectionsLength()`, ensuring correct idle state detection after request completion.

**Real-world Application**: Handles passengers boarding elevators that happen to be stopped at their current floor, common in lobby areas and parking garages.

## Monitoring and Observability

### Logging
- **Movement Tracking**: Logs every floor change and direction change
- **Request Processing**: Records door operations and request fulfillment
- **Performance Metrics**: Tracks idle time and utilization

### Health Monitoring
- **Circuit Breaker**: Protects against cascading failures
- **Timeout Management**: Prevents infinite blocking operations
- **Error Recovery**: Graceful handling of operational failures

## Future Optimizations

### Potential Improvements
1. **Predictive Dispatching**: Pre-position elevators based on usage patterns
2. **Load Balancing**: Dynamic request distribution among multiple elevators
3. **Priority Queuing**: VIP or emergency request prioritization
4. **Energy Optimization**: Minimize power consumption during low-traffic periods

### Advanced Features
1. **Destination Dispatch**: Passengers enter destination before boarding
2. **AI-Powered Patterns**: Machine learning for traffic prediction
3. **Dynamic Zoning**: Automatic floor range adjustment based on demand 