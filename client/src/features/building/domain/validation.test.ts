import { describe, it, expect } from 'vitest';
import { validateConfig } from './validation';

describe('validateConfig', () => {
  it('accepts a valid config', () => {
    expect(validateConfig({ name: 'Main-A', minFloor: 0, maxFloor: 10 }, [])).toEqual({});
  });

  it('rejects empty name', () => {
    expect(validateConfig({ name: '  ', minFloor: 0, maxFloor: 10 }, []).name).toBeTruthy();
  });

  it('rejects duplicate name', () => {
    expect(validateConfig({ name: 'A', minFloor: 0, maxFloor: 10 }, ['A']).name).toMatch(/already exists/);
  });

  it('rejects invalid characters', () => {
    expect(validateConfig({ name: 'a@b', minFloor: 0, maxFloor: 10 }, []).name).toBeTruthy();
  });

  it('rejects min >= max', () => {
    expect(validateConfig({ name: 'A', minFloor: 5, maxFloor: 5 }, []).maxFloor).toBeTruthy();
  });

  it('rejects range over 50 floors', () => {
    expect(validateConfig({ name: 'A', minFloor: -10, maxFloor: 100 }, []).maxFloor).toMatch(/exceed 50/);
  });

  it('rejects out-of-bounds floors', () => {
    expect(validateConfig({ name: 'A', minFloor: -20, maxFloor: 10 }, []).minFloor).toBeTruthy();
  });
});
