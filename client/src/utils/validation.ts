// utils/validation.ts - Validation utilities
import type { ElevatorConfig } from '../types';

export interface ValidationError {
    field: string;
    message: string;
}

export interface ValidationResult {
    valid: boolean;
    errors: ValidationError[];
}

export class ValidationService {
    /**
     * Validate elevator configuration before creation.
     */
    validateElevatorConfig(config: ElevatorConfig): ValidationResult {
        const errors: ValidationError[] = [];

        // Name validation
        if (!config.name || config.name.trim().length === 0) {
            errors.push({ field: 'name', message: 'Elevator name is required' });
        } else if (config.name.length > 50) {
            errors.push({ field: 'name', message: 'Elevator name must be 50 characters or less' });
        } else if (!/^[a-zA-Z0-9\-_\s]+$/.test(config.name)) {
            errors.push({ field: 'name', message: 'Elevator name can only contain letters, numbers, hyphens, underscores, and spaces' });
        }

        // Floor range validation
        if (config.minFloor === undefined || config.minFloor === null) {
            errors.push({ field: 'minFloor', message: 'Minimum floor is required' });
        } else if (!Number.isInteger(config.minFloor)) {
            errors.push({ field: 'minFloor', message: 'Minimum floor must be an integer' });
        } else if (config.minFloor < -10 || config.minFloor > 100) {
            errors.push({ field: 'minFloor', message: 'Minimum floor must be between -10 and 100' });
        }

        if (config.maxFloor === undefined || config.maxFloor === null) {
            errors.push({ field: 'maxFloor', message: 'Maximum floor is required' });
        } else if (!Number.isInteger(config.maxFloor)) {
            errors.push({ field: 'maxFloor', message: 'Maximum floor must be an integer' });
        } else if (config.maxFloor < -10 || config.maxFloor > 100) {
            errors.push({ field: 'maxFloor', message: 'Maximum floor must be between -10 and 100' });
        }

        // Floor range logic validation
        if (config.minFloor !== undefined && config.maxFloor !== undefined) {
            if (config.minFloor >= config.maxFloor) {
                errors.push({ field: 'maxFloor', message: 'Maximum floor must be greater than minimum floor' });
            }

            const floorRange = config.maxFloor - config.minFloor;
            if (floorRange < 1) {
                errors.push({ field: 'maxFloor', message: 'Elevator must serve at least 2 floors' });
            } else if (floorRange > 50) {
                errors.push({ field: 'maxFloor', message: 'Elevator range cannot exceed 50 floors' });
            }
        }

        // Capacity validation (optional field)
        if (config.capacity !== undefined) {
            if (!Number.isInteger(config.capacity) || config.capacity <= 0) {
                errors.push({ field: 'capacity', message: 'Capacity must be a positive integer' });
            } else if (config.capacity > 50) {
                errors.push({ field: 'capacity', message: 'Capacity cannot exceed 50 people' });
            }
        }

        // Threshold validation (optional field)
        if (config.threshold !== undefined) {
            if (!Number.isInteger(config.threshold) || config.threshold <= 0) {
                errors.push({ field: 'threshold', message: 'Threshold must be a positive integer' });
            } else if (config.threshold > 100) {
                errors.push({ field: 'threshold', message: 'Threshold cannot exceed 100' });
            }
        }

        return {
            valid: errors.length === 0,
            errors
        };
    }

    /**
     * Validate floor request parameters.
     */
    validateFloorRequest(fromFloor: number, toFloor: number): ValidationResult {
        const errors: ValidationError[] = [];

        if (!Number.isInteger(fromFloor)) {
            errors.push({ field: 'fromFloor', message: 'From floor must be an integer' });
        }

        if (!Number.isInteger(toFloor)) {
            errors.push({ field: 'toFloor', message: 'To floor must be an integer' });
        }

        if (fromFloor === toFloor) {
            errors.push({ field: 'toFloor', message: 'Destination floor must be different from current floor' });
        }

        if (Math.abs(fromFloor - toFloor) > 50) {
            errors.push({ field: 'toFloor', message: 'Floor range too large (max 50 floors)' });
        }

        return {
            valid: errors.length === 0,
            errors
        };
    }

    /**
     * Validate elevator name uniqueness.
     */
    validateElevatorNameUniqueness(name: string, existingNames: string[]): ValidationResult {
        const errors: ValidationError[] = [];

        if (existingNames.includes(name.trim())) {
            errors.push({ field: 'name', message: 'An elevator with this name already exists' });
        }

        return {
            valid: errors.length === 0,
            errors
        };
    }

    /**
     * Sanitize input string to prevent XSS.
     */
    sanitizeInput(input: string): string {
        return input
            .replace(/[<>]/g, '') // Remove potential HTML tags
            .trim()
            .substring(0, 100); // Limit length
    }

    /**
     * Validate numeric input within range.
     */
    validateNumberInRange(value: number, min: number, max: number, fieldName: string): ValidationError | null {
        if (typeof value !== 'number' || isNaN(value)) {
            return { field: fieldName, message: `${fieldName} must be a valid number` };
        }

        if (value < min || value > max) {
            return { field: fieldName, message: `${fieldName} must be between ${min} and ${max}` };
        }

        return null;
    }

    /**
     * Validate required field.
     */
    validateRequired(value: any, fieldName: string): ValidationError | null {
        if (value === undefined || value === null || value === '') {
            return { field: fieldName, message: `${fieldName} is required` };
        }

        return null;
    }

    /**
     * Get error message for a specific field.
     */
    getFieldError(errors: ValidationError[], fieldName: string): string | null {
        const error = errors.find(e => e.field === fieldName);
        return error ? error.message : null;
    }

    /**
     * Check if a field has errors.
     */
    hasFieldError(errors: ValidationError[], fieldName: string): boolean {
        return errors.some(e => e.field === fieldName);
    }

    /**
     * Format validation errors for display.
     */
    formatErrors(errors: ValidationError[]): string[] {
        return errors.map(error => `${error.field}: ${error.message}`);
    }
}

export const validationService = new ValidationService(); 