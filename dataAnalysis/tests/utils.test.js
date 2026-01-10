import { describe, it, expect, beforeEach } from 'vitest';

/**
 * Utility functions extracted from app.js for testing
 * These functions are tested in isolation to ensure correct behavior
 */

// formatDateDDMMYYYY - formats dates as dd-mm-yyyy HH:MM:SS
function formatDateDDMMYYYY(date, includeTime = true) {
    // Handle null, undefined, and invalid values explicitly
    if (date === null || date === undefined || date === '') {
        return 'Invalid Date';
    }
    const d = new Date(date);
    if (isNaN(d.getTime())) return 'Invalid Date';

    const day = String(d.getDate()).padStart(2, '0');
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const year = d.getFullYear();

    if (!includeTime) {
        return `${day}-${month}-${year}`;
    }

    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    const seconds = String(d.getSeconds()).padStart(2, '0');

    return `${day}-${month}-${year} ${hours}:${minutes}:${seconds}`;
}

// parseTimestamp - parses various timestamp formats to milliseconds
function parseTimestamp(timestamp) {
    if (!timestamp) return 0;

    // If it's already a number (milliseconds), return it
    if (typeof timestamp === 'number') return timestamp;

    // If it's a string (ISO 8601 / RFC3339), parse it
    if (typeof timestamp === 'string') {
        const date = new Date(timestamp);
        return date.getTime();
    }

    // If it's a Date object
    if (timestamp instanceof Date) return timestamp.getTime();

    return 0;
}

describe('formatDateDDMMYYYY', () => {
    it('formats a date with time correctly', () => {
        const date = new Date('2024-01-15T14:30:45');
        const result = formatDateDDMMYYYY(date, true);
        expect(result).toBe('15-01-2024 14:30:45');
    });

    it('formats a date without time correctly', () => {
        const date = new Date('2024-01-15T14:30:45');
        const result = formatDateDDMMYYYY(date, false);
        expect(result).toBe('15-01-2024');
    });

    it('pads single digit day and month', () => {
        const date = new Date('2024-05-07T09:05:03');
        const result = formatDateDDMMYYYY(date, true);
        expect(result).toBe('07-05-2024 09:05:03');
    });

    it('handles midnight correctly', () => {
        const date = new Date('2024-12-31T00:00:00');
        const result = formatDateDDMMYYYY(date, true);
        expect(result).toBe('31-12-2024 00:00:00');
    });

    it('handles end of day correctly', () => {
        const date = new Date('2024-12-31T23:59:59');
        const result = formatDateDDMMYYYY(date, true);
        expect(result).toBe('31-12-2024 23:59:59');
    });

    it('returns Invalid Date for invalid input', () => {
        const result = formatDateDDMMYYYY('not a date');
        expect(result).toBe('Invalid Date');
    });

    it('returns Invalid Date for null', () => {
        const result = formatDateDDMMYYYY(null);
        expect(result).toBe('Invalid Date');
    });

    it('returns Invalid Date for undefined', () => {
        const result = formatDateDDMMYYYY(undefined);
        expect(result).toBe('Invalid Date');
    });

    it('handles timestamp in milliseconds', () => {
        const timestamp = 1704067200000; // 2024-01-01 00:00:00 UTC
        const date = new Date(timestamp);
        const result = formatDateDDMMYYYY(date, false);
        // Note: exact output depends on timezone, so we just check format
        expect(result).toMatch(/^\d{2}-\d{2}-\d{4}$/);
    });

    it('handles leap year date', () => {
        const date = new Date('2024-02-29T12:00:00');
        const result = formatDateDDMMYYYY(date, false);
        expect(result).toBe('29-02-2024');
    });

    it('defaults to including time when no second argument provided', () => {
        const date = new Date('2024-06-15T10:30:00');
        const result = formatDateDDMMYYYY(date);
        expect(result).toBe('15-06-2024 10:30:00');
    });
});

describe('parseTimestamp', () => {
    it('returns 0 for null input', () => {
        expect(parseTimestamp(null)).toBe(0);
    });

    it('returns 0 for undefined input', () => {
        expect(parseTimestamp(undefined)).toBe(0);
    });

    it('returns 0 for empty string', () => {
        expect(parseTimestamp('')).toBe(0);
    });

    it('returns 0 for zero', () => {
        expect(parseTimestamp(0)).toBe(0);
    });

    it('returns the number directly if input is a number', () => {
        const timestamp = 1704067200000;
        expect(parseTimestamp(timestamp)).toBe(timestamp);
    });

    it('parses ISO 8601 string correctly', () => {
        const isoString = '2024-01-01T00:00:00.000Z';
        const result = parseTimestamp(isoString);
        expect(result).toBe(1704067200000);
    });

    it('parses ISO 8601 string with timezone', () => {
        const isoString = '2024-01-01T10:00:00+10:00';
        const result = parseTimestamp(isoString);
        // 10:00 +10:00 is 00:00 UTC
        expect(result).toBe(1704067200000);
    });

    it('handles Date object input', () => {
        const date = new Date('2024-01-01T00:00:00.000Z');
        const result = parseTimestamp(date);
        expect(result).toBe(1704067200000);
    });

    it('returns 0 for invalid object types', () => {
        expect(parseTimestamp({})).toBe(0);
        expect(parseTimestamp([])).toBe(0);
    });

    it('parses date-only string', () => {
        const dateString = '2024-01-15';
        const result = parseTimestamp(dateString);
        expect(result).toBeGreaterThan(0);
    });

    it('handles negative timestamps', () => {
        const negativeTimestamp = -1000;
        expect(parseTimestamp(negativeTimestamp)).toBe(-1000);
    });

    it('handles very large timestamps', () => {
        const futureTimestamp = 2524608000000; // Year 2050
        expect(parseTimestamp(futureTimestamp)).toBe(futureTimestamp);
    });
});

describe('Date boundary calculations', () => {
    it('calculates start of day correctly', () => {
        const date = new Date('2024-01-15T14:30:00');
        const startOfDay = new Date(date);
        startOfDay.setHours(0, 0, 0, 0);
        
        expect(startOfDay.getHours()).toBe(0);
        expect(startOfDay.getMinutes()).toBe(0);
        expect(startOfDay.getSeconds()).toBe(0);
        expect(startOfDay.getMilliseconds()).toBe(0);
    });

    it('calculates end of day correctly', () => {
        const date = new Date('2024-01-15T00:00:00');
        const endOfDay = new Date(date);
        endOfDay.setHours(23, 59, 59, 999);
        
        expect(endOfDay.getHours()).toBe(23);
        expect(endOfDay.getMinutes()).toBe(59);
        expect(endOfDay.getSeconds()).toBe(59);
        expect(endOfDay.getMilliseconds()).toBe(999);
    });

    it('handles timezone-aware date boundaries', () => {
        // This tests that we can work with dates in a specific timezone
        const dateString = '2024-01-15';
        const date = new Date(dateString);
        expect(date.getFullYear()).toBe(2024);
        expect(date.getMonth()).toBe(0); // January is 0
        expect(date.getDate()).toBe(15);
    });
});

describe('Date formatting edge cases', () => {
    it('handles year transition correctly', () => {
        const newYearsEve = new Date('2024-12-31T23:59:59');
        const newYearsDay = new Date('2025-01-01T00:00:00');
        
        expect(formatDateDDMMYYYY(newYearsEve, false)).toBe('31-12-2024');
        expect(formatDateDDMMYYYY(newYearsDay, false)).toBe('01-01-2025');
    });

    it('handles month transition correctly', () => {
        const endOfJan = new Date('2024-01-31T12:00:00');
        const startOfFeb = new Date('2024-02-01T12:00:00');
        
        expect(formatDateDDMMYYYY(endOfJan, false)).toBe('31-01-2024');
        expect(formatDateDDMMYYYY(startOfFeb, false)).toBe('01-02-2024');
    });

    it('handles February 28 in non-leap year', () => {
        const date = new Date('2023-02-28T12:00:00');
        expect(formatDateDDMMYYYY(date, false)).toBe('28-02-2023');
    });

    it('handles February 29 in leap year', () => {
        const date = new Date('2024-02-29T12:00:00');
        expect(formatDateDDMMYYYY(date, false)).toBe('29-02-2024');
    });
});


