// Client-side validation for the create-elevator form. Mirrors the backend's
// accepted ranges so users get instant feedback before a round-trip.

import type { ElevatorConfig } from './types';

export type FieldErrors = Partial<Record<'name' | 'minFloor' | 'maxFloor', string>>;

export function validateConfig(config: ElevatorConfig, existingNames: string[]): FieldErrors {
  const errors: FieldErrors = {};
  const name = config.name.trim();

  if (!name) {
    errors.name = 'Name is required';
  } else if (name.length > 50) {
    errors.name = 'Name must be 50 characters or fewer';
  } else if (!/^[a-zA-Z0-9\-_\s]+$/.test(name)) {
    errors.name = 'Use only letters, numbers, hyphens, underscores, spaces';
  } else if (existingNames.includes(name)) {
    errors.name = 'An elevator with this name already exists';
  }

  if (!Number.isInteger(config.minFloor) || config.minFloor < -10 || config.minFloor > 100) {
    errors.minFloor = 'Min floor must be an integer between -10 and 100';
  }
  if (!Number.isInteger(config.maxFloor) || config.maxFloor < -10 || config.maxFloor > 100) {
    errors.maxFloor = 'Max floor must be an integer between -10 and 100';
  }

  if (!errors.minFloor && !errors.maxFloor) {
    if (config.minFloor >= config.maxFloor) {
      errors.maxFloor = 'Max floor must be greater than min floor';
    } else if (config.maxFloor - config.minFloor > 50) {
      errors.maxFloor = 'Range cannot exceed 50 floors';
    }
  }

  return errors;
}
