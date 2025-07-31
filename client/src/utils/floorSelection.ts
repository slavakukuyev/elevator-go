// utils/floorSelection.ts - Floor selection utilities with TypeScript
import type { Elevator, ValidationResult } from '../types';

export class FloorSelectionService {
    /**
     * Calculate available destination floors from a given floor.
     * Only returns floors that can actually be reached by an elevator from the current floor.
     */
    getAvailableDestinations(currentFloor: number, elevators: Elevator[]): number[] {
        const availableFloors = new Set<number>();

        // Find elevators that can serve the current floor
        const servingElevators = elevators.filter(elev =>
            currentFloor >= elev.minFloor && currentFloor <= elev.maxFloor &&
            elev.status !== 'error'
        );

        // For each serving elevator, check which floors it can actually reach from current floor
        servingElevators.forEach(elev => {
            // Only add floors that are within this elevator's range
            for (let f = elev.minFloor; f <= elev.maxFloor; f++) {
                if (f !== currentFloor) {
                    // Additional validation: ensure there's a valid path
                    // This prevents showing floors that can't be reached due to elevator constraints
                    if (this.isValidDestination(currentFloor, f, elev)) {
                        availableFloors.add(f);
                    }
                }
            }
        });

        return Array.from(availableFloors).sort((a, b) => a - b);
    }

    /**
     * Check if a destination floor is valid for a specific elevator from the current floor.
     */
    private isValidDestination(fromFloor: number, toFloor: number, elevator: Elevator): boolean {
        // Basic range check
        if (toFloor < elevator.minFloor || toFloor > elevator.maxFloor) {
            return false;
        }

        // Check if elevator can physically reach the destination from current position
        // This prevents showing floors that are outside the elevator's operational range
        if (fromFloor < elevator.minFloor || fromFloor > elevator.maxFloor) {
            return false;
        }

        // For parking elevators or elevators with specific constraints,
        // we might need additional validation here
        return true;
    }

    /**
     * Find the optimal elevator to serve a requested trip.
     */
    findOptimalElevator(fromFloor: number, toFloor: number, elevators: Elevator[]): Elevator | null {
        const candidates = elevators.filter(elev =>
            fromFloor >= elev.minFloor && fromFloor <= elev.maxFloor &&
            toFloor >= elev.minFloor && toFloor <= elev.maxFloor &&
            elev.status !== 'error'
        );

        if (candidates.length === 0) return null;

        const direction = toFloor > fromFloor ? 'up' : 'down';

        // 1. Elevators already moving in the required direction and can serve the request
        const movingSameWay = candidates.filter(elev =>
            elev.status === 'moving' && elev.direction === direction &&
            this.canServeRequest(elev, fromFloor, toFloor)
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

    /**
     * Check if an elevator can serve a request based on its current state and position.
     */
    private canServeRequest(elevator: Elevator, fromFloor: number, toFloor: number): boolean {
        if (elevator.status === 'error') return false;

        // Check if elevator is within range
        if (fromFloor < elevator.minFloor || fromFloor > elevator.maxFloor ||
            toFloor < elevator.minFloor || toFloor > elevator.maxFloor) {
            return false;
        }

        // If elevator is moving, check if it can pick up the request on its way
        if (elevator.status === 'moving' && elevator.direction) {
            const requestDirection = toFloor > fromFloor ? 'up' : 'down';

            // Same direction and elevator hasn't passed the pickup floor
            if (elevator.direction === requestDirection) {
                if (elevator.direction === 'up' && elevator.currentFloor <= fromFloor) {
                    return true;
                }
                if (elevator.direction === 'down' && elevator.currentFloor >= fromFloor) {
                    return true;
                }
            }
        }

        return true;
    }

    /**
     * Find the closest elevator to a target floor.
     */
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
    validateFloorRequest(fromFloor: number, toFloor: number, elevators: Elevator[]): ValidationResult {
        if (fromFloor === toFloor) {
            return { valid: false, message: `You are already on floor ${fromFloor}` };
        }

        const elev = this.findOptimalElevator(fromFloor, toFloor, elevators);
        if (!elev) {
            return {
                valid: false,
                message: `No elevator can serve request from floor ${fromFloor} to ${toFloor}`
            };
        }

        return {
            valid: true,
            elevator: elev,
            message: `Elevator ${elev.name} will serve this request`
        };
    }

    /**
     * Calculate estimated wait time for a floor request.
     */
    calculateWaitTime(fromFloor: number, toFloor: number, elevators: Elevator[]): number {
        const elevator = this.findOptimalElevator(fromFloor, toFloor, elevators);
        if (!elevator) return -1;

        const distance = Math.abs(elevator.currentFloor - fromFloor);
        const baseTime = distance * 3; // 3 seconds per floor

        // Add delay if elevator is moving in opposite direction
        if (elevator.status === 'moving' && elevator.direction) {
            const requestDirection = toFloor > fromFloor ? 'up' : 'down';
            if (elevator.direction !== requestDirection) {
                return baseTime + 10; // Additional 10 seconds for direction change
            }
        }

        return baseTime;
    }

    /**
     * Get all floors that have call buttons (excluding current floor).
     * Uses the same logic as getAvailableDestinations for consistency.
     */
    getCallableFloors(currentFloor: number, elevators: Elevator[]): number[] {
        return this.getAvailableDestinations(currentFloor, elevators);
    }

    /**
     * Format floor display (handle basement floors, ground floor, etc.).
     */
    formatFloorDisplay(floor: number | undefined): string {
        if (floor === undefined || floor === null) return '?';
        if (floor === 0) return 'G'; // Ground floor
        if (floor < 0) return `B${Math.abs(floor)}`; // Basement
        return floor.toString();
    }

    /**
     * Get the direction needed to go from one floor to another.
     */
    getDirection(fromFloor: number, toFloor: number): 'up' | 'down' | null {
        if (fromFloor === toFloor) return null;
        return toFloor > fromFloor ? 'up' : 'down';
    }

    /**
     * Check if a floor is valid for a given elevator.
     */
    isFloorValid(floor: number, elevator: Elevator): boolean {
        return floor >= elevator.minFloor && floor <= elevator.maxFloor;
    }

    /**
     * Get recommended floors based on usage patterns (mock implementation).
     */
    getRecommendedFloors(currentFloor: number, elevators: Elevator[]): number[] {
        // Mock implementation - in a real app, this would use historical data
        const availableFloors = this.getAvailableDestinations(currentFloor, elevators);

        // Common floors (ground, typical office floors)
        const commonFloors = [0, 1, 2, 5, 10].filter(f =>
            availableFloors.includes(f)
        );

        return commonFloors.slice(0, 3); // Return top 3 recommendations
    }
}

export const floorSelectionService = new FloorSelectionService(); 