import { describe, it, expect, beforeEach } from 'vitest';

/**
 * Filter logic tests
 * Tests the client-side filtering functionality used in the dashboard
 */

// Mock alert data for testing
const createMockAlert = (overrides = {}) => ({
    UUID: `alert-${Math.random().toString(36).substr(2, 9)}`,
    Type: 'POLICE',
    Subtype: 'POLICE_VISIBLE',
    Street: 'Test Street',
    City: 'Sydney',
    Country: 'AU',
    LocationGeo: { latitude: -33.8688, longitude: 151.2093 },
    Reliability: 8,
    Confidence: 7,
    PublishTime: Date.now() - 3600000, // 1 hour ago
    ExpireTime: Date.now() + 3600000, // 1 hour from now
    ScrapeTime: Date.now() - 1800000, // 30 minutes ago
    ActiveMillis: 3600000,
    NThumbsUpLast: 5,
    LastVerificationMillis: null,
    ...overrides,
});

// Simulated filter function (matches app.js logic)
function applyFilters(alerts, { subtypes = [], streets = [], verifiedOnly = false }) {
    let filtered = [...alerts];

    // Filter by selected subtypes (if any are selected)
    if (subtypes.length > 0) {
        filtered = filtered.filter(a => subtypes.includes(a.Subtype));
    }

    // Filter by selected streets (if any are selected)
    if (streets.length > 0) {
        filtered = filtered.filter(a => streets.includes(a.Street || ''));
    }

    // Filter by verified status (if checkbox is checked)
    if (verifiedOnly) {
        filtered = filtered.filter(a => {
            const thumbsUp = a.NThumbsUpLast || 0;
            return thumbsUp !== null && thumbsUp > 0;
        });
    }

    return filtered;
}

// Contains helper (matches app.js)
function contains(slice, value) {
    return slice.includes(value);
}

describe('applyFilters', () => {
    let mockAlerts;

    beforeEach(() => {
        mockAlerts = [
            createMockAlert({ UUID: 'alert-1', Subtype: 'POLICE_VISIBLE', Street: 'Hume Highway', NThumbsUpLast: 5 }),
            createMockAlert({ UUID: 'alert-2', Subtype: 'POLICE_WITH_MOBILE_CAMERA', Street: 'Federal Highway', NThumbsUpLast: 3 }),
            createMockAlert({ UUID: 'alert-3', Subtype: 'POLICE_VISIBLE', Street: 'Federal Highway', NThumbsUpLast: 0 }),
            createMockAlert({ UUID: 'alert-4', Subtype: 'POLICE_HIDING', Street: 'Hume Highway', NThumbsUpLast: 10 }),
            createMockAlert({ UUID: 'alert-5', Subtype: '', Street: '', NThumbsUpLast: 2 }),
        ];
    });

    describe('no filters applied', () => {
        it('returns all alerts when no filters are applied', () => {
            const result = applyFilters(mockAlerts, {});
            expect(result).toHaveLength(5);
        });

        it('returns empty array for empty input', () => {
            const result = applyFilters([], {});
            expect(result).toHaveLength(0);
        });
    });

    describe('subtype filtering', () => {
        it('filters by single subtype', () => {
            const result = applyFilters(mockAlerts, { subtypes: ['POLICE_VISIBLE'] });
            expect(result).toHaveLength(2);
            expect(result.every(a => a.Subtype === 'POLICE_VISIBLE')).toBe(true);
        });

        it('filters by multiple subtypes', () => {
            const result = applyFilters(mockAlerts, { 
                subtypes: ['POLICE_VISIBLE', 'POLICE_WITH_MOBILE_CAMERA'] 
            });
            expect(result).toHaveLength(3);
        });

        it('filters by empty subtype', () => {
            const result = applyFilters(mockAlerts, { subtypes: [''] });
            expect(result).toHaveLength(1);
            expect(result[0].UUID).toBe('alert-5');
        });

        it('returns empty array for non-matching subtype', () => {
            const result = applyFilters(mockAlerts, { subtypes: ['NONEXISTENT'] });
            expect(result).toHaveLength(0);
        });
    });

    describe('street filtering', () => {
        it('filters by single street', () => {
            const result = applyFilters(mockAlerts, { streets: ['Hume Highway'] });
            expect(result).toHaveLength(2);
            expect(result.every(a => a.Street === 'Hume Highway')).toBe(true);
        });

        it('filters by multiple streets', () => {
            const result = applyFilters(mockAlerts, { 
                streets: ['Hume Highway', 'Federal Highway'] 
            });
            expect(result).toHaveLength(4);
        });

        it('filters by empty street', () => {
            const result = applyFilters(mockAlerts, { streets: [''] });
            expect(result).toHaveLength(1);
            expect(result[0].UUID).toBe('alert-5');
        });

        it('returns empty array for non-matching street', () => {
            const result = applyFilters(mockAlerts, { streets: ['Nonexistent Street'] });
            expect(result).toHaveLength(0);
        });
    });

    describe('verified only filtering', () => {
        it('filters to only verified alerts (thumbs up > 0)', () => {
            const result = applyFilters(mockAlerts, { verifiedOnly: true });
            expect(result).toHaveLength(4);
            expect(result.every(a => a.NThumbsUpLast > 0)).toBe(true);
        });

        it('excludes alerts with zero thumbs up', () => {
            const result = applyFilters(mockAlerts, { verifiedOnly: true });
            expect(result.find(a => a.UUID === 'alert-3')).toBeUndefined();
        });

        it('includes all alerts when verifiedOnly is false', () => {
            const result = applyFilters(mockAlerts, { verifiedOnly: false });
            expect(result).toHaveLength(5);
        });
    });

    describe('combined filters', () => {
        it('applies subtype and street filters together', () => {
            const result = applyFilters(mockAlerts, { 
                subtypes: ['POLICE_VISIBLE'],
                streets: ['Hume Highway']
            });
            expect(result).toHaveLength(1);
            expect(result[0].UUID).toBe('alert-1');
        });

        it('applies subtype and verified filters together', () => {
            const result = applyFilters(mockAlerts, { 
                subtypes: ['POLICE_VISIBLE'],
                verifiedOnly: true
            });
            expect(result).toHaveLength(1);
            expect(result[0].UUID).toBe('alert-1');
        });

        it('applies all three filters together', () => {
            const result = applyFilters(mockAlerts, { 
                subtypes: ['POLICE_VISIBLE', 'POLICE_HIDING'],
                streets: ['Hume Highway'],
                verifiedOnly: true
            });
            expect(result).toHaveLength(2);
        });

        it('returns empty when combined filters match nothing', () => {
            const result = applyFilters(mockAlerts, { 
                subtypes: ['POLICE_VISIBLE'],
                streets: ['Nonexistent Street'],
                verifiedOnly: true
            });
            expect(result).toHaveLength(0);
        });
    });
});

describe('contains helper', () => {
    it('returns true when value exists in array', () => {
        expect(contains(['a', 'b', 'c'], 'b')).toBe(true);
    });

    it('returns false when value does not exist', () => {
        expect(contains(['a', 'b', 'c'], 'd')).toBe(false);
    });

    it('handles empty array', () => {
        expect(contains([], 'a')).toBe(false);
    });

    it('handles empty string value', () => {
        expect(contains(['', 'a', 'b'], '')).toBe(true);
    });

    it('is case sensitive', () => {
        expect(contains(['Apple', 'Banana'], 'apple')).toBe(false);
    });
});

describe('Hume Highway filter logic', () => {
    const humeAlerts = [
        createMockAlert({ Street: 'Hume Highway' }),
        createMockAlert({ Street: 'Federal Highway' }),
        createMockAlert({ Street: 'Pacific Highway' }),
        createMockAlert({ Street: 'Hume Freeway' }),
        createMockAlert({ Street: 'Federal Bypass' }),
        createMockAlert({ Street: 'Main Street' }),
    ];

    it('finds streets containing "Hume" (case-insensitive)', () => {
        const humeStreets = humeAlerts
            .map(a => a.Street || '')
            .filter(street => street.toLowerCase().includes('hume'));
        
        expect(humeStreets).toHaveLength(2);
        expect(humeStreets).toContain('Hume Highway');
        expect(humeStreets).toContain('Hume Freeway');
    });

    it('finds streets containing "Federal" (case-insensitive)', () => {
        const federalStreets = humeAlerts
            .map(a => a.Street || '')
            .filter(street => street.toLowerCase().includes('federal'));
        
        expect(federalStreets).toHaveLength(2);
        expect(federalStreets).toContain('Federal Highway');
        expect(federalStreets).toContain('Federal Bypass');
    });

    it('finds streets containing either "Hume" or "Federal"', () => {
        const matchingStreets = [...new Set(
            humeAlerts
                .map(a => a.Street || '')
                .filter(street => {
                    const lowerStreet = street.toLowerCase();
                    return lowerStreet.includes('hume') || lowerStreet.includes('federal');
                })
        )];
        
        expect(matchingStreets).toHaveLength(4);
    });
});

describe('Alert deduplication', () => {
    it('deduplicates alerts by UUID', () => {
        const duplicateAlerts = [
            createMockAlert({ UUID: 'dup-1', Street: 'Street A' }),
            createMockAlert({ UUID: 'dup-1', Street: 'Street A' }), // Duplicate
            createMockAlert({ UUID: 'dup-2', Street: 'Street B' }),
            createMockAlert({ UUID: 'dup-1', Street: 'Street A' }), // Another duplicate
        ];

        const alertsMap = new Map();
        for (const alert of duplicateAlerts) {
            if (!alertsMap.has(alert.UUID)) {
                alertsMap.set(alert.UUID, alert);
            }
        }

        expect(alertsMap.size).toBe(2);
        expect(alertsMap.has('dup-1')).toBe(true);
        expect(alertsMap.has('dup-2')).toBe(true);
    });

    it('keeps most recent alert when deduplicating by ExpireTime', () => {
        const duplicateAlerts = [
            createMockAlert({ UUID: 'dup-1', ExpireTime: 1000 }),
            createMockAlert({ UUID: 'dup-1', ExpireTime: 3000 }), // Latest
            createMockAlert({ UUID: 'dup-1', ExpireTime: 2000 }),
        ];

        const alertsMap = new Map();
        for (const alert of duplicateAlerts) {
            const existing = alertsMap.get(alert.UUID);
            if (!existing || alert.ExpireTime > existing.ExpireTime) {
                alertsMap.set(alert.UUID, alert);
            }
        }

        expect(alertsMap.get('dup-1').ExpireTime).toBe(3000);
    });
});

describe('Alert sorting', () => {
    it('sorts alerts by PublishTime ascending', () => {
        const alerts = [
            createMockAlert({ UUID: 'a3', PublishTime: 3000 }),
            createMockAlert({ UUID: 'a1', PublishTime: 1000 }),
            createMockAlert({ UUID: 'a2', PublishTime: 2000 }),
        ];

        const sorted = [...alerts].sort((a, b) => a.PublishTime - b.PublishTime);

        expect(sorted[0].UUID).toBe('a1');
        expect(sorted[1].UUID).toBe('a2');
        expect(sorted[2].UUID).toBe('a3');
    });

    it('sorts alerts by PublishTime descending', () => {
        const alerts = [
            createMockAlert({ UUID: 'a3', PublishTime: 3000 }),
            createMockAlert({ UUID: 'a1', PublishTime: 1000 }),
            createMockAlert({ UUID: 'a2', PublishTime: 2000 }),
        ];

        const sorted = [...alerts].sort((a, b) => b.PublishTime - a.PublishTime);

        expect(sorted[0].UUID).toBe('a3');
        expect(sorted[1].UUID).toBe('a2');
        expect(sorted[2].UUID).toBe('a1');
    });
});

describe('Statistics calculation', () => {
    const statsAlerts = [
        createMockAlert({ Reliability: 8, Confidence: 7, City: 'Sydney' }),
        createMockAlert({ Reliability: 9, Confidence: 8, City: 'Sydney' }),
        createMockAlert({ Reliability: 7, Confidence: 6, City: 'Canberra' }),
        createMockAlert({ Reliability: 8, Confidence: 7, City: 'Sydney' }),
    ];

    it('calculates average reliability correctly', () => {
        const avgReliability = statsAlerts.reduce((sum, a) => sum + a.Reliability, 0) / statsAlerts.length;
        expect(avgReliability).toBe(8);
    });

    it('calculates average confidence correctly', () => {
        const avgConfidence = statsAlerts.reduce((sum, a) => sum + a.Confidence, 0) / statsAlerts.length;
        expect(avgConfidence).toBe(7);
    });

    it('finds top city correctly', () => {
        const cityCounts = {};
        statsAlerts.forEach(a => {
            if (a.City) {
                cityCounts[a.City] = (cityCounts[a.City] || 0) + 1;
            }
        });
        
        const topCity = Object.entries(cityCounts).sort((a, b) => b[1] - a[1])[0];
        expect(topCity[0]).toBe('Sydney');
        expect(topCity[1]).toBe(3);
    });

    it('handles empty alerts array', () => {
        const emptyAlerts = [];
        const avgReliability = emptyAlerts.length > 0 
            ? emptyAlerts.reduce((sum, a) => sum + a.Reliability, 0) / emptyAlerts.length
            : 0;
        expect(avgReliability).toBe(0);
    });
});


