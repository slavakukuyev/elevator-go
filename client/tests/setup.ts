import '@testing-library/jest-dom';

// Mock WebSocket for tests
global.WebSocket = class WebSocket {
    constructor() {
        // Mock implementation
    }
    send() { }
    close() { }
} as any;

// Mock localStorage
Object.defineProperty(window, 'localStorage', {
    value: {
        getItem: () => null,
        setItem: () => { },
        removeItem: () => { },
        clear: () => { },
    },
    writable: true,
});

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: () => ({
        matches: false,
        addListener: () => { },
        removeListener: () => { },
    }),
}); 