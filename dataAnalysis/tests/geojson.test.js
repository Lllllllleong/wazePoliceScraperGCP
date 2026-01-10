import { describe, it, expect, beforeEach } from 'vitest';

/**
 * GeoJSON transformation tests
 * Tests the conversion of alerts to GeoJSON features for map visualization
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
    PublishTime: Date.now() - 3600000,
    ExpireTime: Date.now() + 3600000,
    ScrapeTime: Date.now() - 1800000,
    ActiveMillis: 3600000,
    NThumbsUpLast: 5,
    LastVerificationMillis: null,
    ...overrides,
});

// Simplified GeoJSON creation function (matches app.js logic)
function createGeoJSONFromAlerts(alerts, normalizeDates = false) {
    const geojsonFeatures = [];

    let referenceDate = null;
    if (normalizeDates) {
        referenceDate = new Date();
        referenceDate.setHours(0, 0, 0, 0);
    }

    const normalizeTimestamp = (timestamp) => {
        if (!normalizeDates) return timestamp;

        const date = new Date(timestamp);
        const normalized = new Date(referenceDate);
        normalized.setHours(date.getHours(), date.getMinutes(), date.getSeconds(), date.getMilliseconds());
        return normalized.getTime();
    };

    alerts.forEach(alert => {
        const lat = alert.LocationGeo.latitude || alert.LocationGeo.y;
        const lng = alert.LocationGeo.longitude || alert.LocationGeo.x;

        if (!lat || !lng || isNaN(lat) || isNaN(lng)) {
            return; // Skip alerts without valid coordinates
        }

        const publishTime = normalizeTimestamp(alert.PublishTime);
        const expireTime = normalizeTimestamp(alert.ExpireTime);
        const lastVerificationMillis = alert.LastVerificationMillis
            ? normalizeTimestamp(alert.LastVerificationMillis)
            : null;
        const isVerified = lastVerificationMillis !== null && lastVerificationMillis !== undefined;

        if (isVerified) {
            // Create two features: verified period and unverified period
            geojsonFeatures.push({
                type: 'Feature',
                properties: {
                    start: publishTime,
                    end: lastVerificationMillis,
                    verified: true,
                    subtype: alert.Subtype,
                    uuid: alert.UUID,
                    street: alert.Street || 'Unknown Location',
                    city: alert.City || 'Unknown City',
                    reliability: alert.Reliability,
                    thumbsUp: alert.NThumbsUpLast,
                },
                geometry: {
                    type: 'Point',
                    coordinates: [lng, lat]
                }
            });

            geojsonFeatures.push({
                type: 'Feature',
                properties: {
                    start: lastVerificationMillis,
                    end: expireTime,
                    verified: false,
                    subtype: alert.Subtype,
                    uuid: alert.UUID,
                    street: alert.Street || 'Unknown Location',
                    city: alert.City || 'Unknown City',
                    reliability: alert.Reliability,
                    thumbsUp: alert.NThumbsUpLast,
                },
                geometry: {
                    type: 'Point',
                    coordinates: [lng, lat]
                }
            });
        } else {
            // Single feature: PublishTime to ExpireTime (unverified)
            geojsonFeatures.push({
                type: 'Feature',
                properties: {
                    start: publishTime,
                    end: expireTime,
                    verified: false,
                    subtype: alert.Subtype,
                    uuid: alert.UUID,
                    street: alert.Street || 'Unknown Location',
                    city: alert.City || 'Unknown City',
                    reliability: alert.Reliability,
                    thumbsUp: alert.NThumbsUpLast,
                },
                geometry: {
                    type: 'Point',
                    coordinates: [lng, lat]
                }
            });
        }
    });

    return geojsonFeatures;
}

describe('createGeoJSONFromAlerts', () => {
    describe('basic conversion', () => {
        it('returns empty array for empty input', () => {
            const result = createGeoJSONFromAlerts([]);
            expect(result).toHaveLength(0);
        });

        it('creates single feature for unverified alert', () => {
            const alerts = [createMockAlert({ LastVerificationMillis: null })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result).toHaveLength(1);
            expect(result[0].properties.verified).toBe(false);
        });

        it('creates two features for verified alert', () => {
            const alerts = [createMockAlert({ 
                LastVerificationMillis: Date.now() - 1800000 // 30 minutes ago
            })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result).toHaveLength(2);
            expect(result[0].properties.verified).toBe(true);
            expect(result[1].properties.verified).toBe(false);
        });
    });

    describe('GeoJSON structure', () => {
        it('has correct Feature type', () => {
            const alerts = [createMockAlert()];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result[0].type).toBe('Feature');
        });

        it('has correct geometry type', () => {
            const alerts = [createMockAlert()];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result[0].geometry.type).toBe('Point');
        });

        it('has coordinates in [longitude, latitude] order', () => {
            const alerts = [createMockAlert({
                LocationGeo: { latitude: -33.8688, longitude: 151.2093 }
            })];
            const result = createGeoJSONFromAlerts(alerts);
            const [lng, lat] = result[0].geometry.coordinates;
            expect(lng).toBe(151.2093);
            expect(lat).toBe(-33.8688);
        });

        it('includes required properties', () => {
            const alerts = [createMockAlert({
                UUID: 'test-uuid',
                Subtype: 'POLICE_VISIBLE',
                Street: 'Test Street',
                City: 'Sydney',
                Reliability: 8,
                NThumbsUpLast: 5,
            })];
            const result = createGeoJSONFromAlerts(alerts);
            const props = result[0].properties;

            expect(props.uuid).toBe('test-uuid');
            expect(props.subtype).toBe('POLICE_VISIBLE');
            expect(props.street).toBe('Test Street');
            expect(props.city).toBe('Sydney');
            expect(props.reliability).toBe(8);
            expect(props.thumbsUp).toBe(5);
            expect(props.start).toBeDefined();
            expect(props.end).toBeDefined();
        });
    });

    describe('coordinate handling', () => {
        it('skips alerts with missing latitude', () => {
            const alerts = [createMockAlert({
                LocationGeo: { latitude: null, longitude: 151.2093 }
            })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result).toHaveLength(0);
        });

        it('skips alerts with missing longitude', () => {
            const alerts = [createMockAlert({
                LocationGeo: { latitude: -33.8688, longitude: null }
            })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result).toHaveLength(0);
        });

        it('skips alerts with NaN coordinates', () => {
            const alerts = [createMockAlert({
                LocationGeo: { latitude: NaN, longitude: 151.2093 }
            })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result).toHaveLength(0);
        });

        it('handles alternative coordinate keys (x, y)', () => {
            const alerts = [createMockAlert({
                LocationGeo: { y: -33.8688, x: 151.2093 }
            })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result).toHaveLength(1);
            const [lng, lat] = result[0].geometry.coordinates;
            expect(lng).toBe(151.2093);
            expect(lat).toBe(-33.8688);
        });
    });

    describe('verified alert handling', () => {
        it('sets correct time ranges for verified period', () => {
            const publishTime = Date.now() - 7200000; // 2 hours ago
            const verificationTime = Date.now() - 3600000; // 1 hour ago
            const expireTime = Date.now() + 3600000; // 1 hour from now

            const alerts = [createMockAlert({
                PublishTime: publishTime,
                ExpireTime: expireTime,
                LastVerificationMillis: verificationTime,
            })];
            const result = createGeoJSONFromAlerts(alerts);

            // First feature: verified period
            expect(result[0].properties.start).toBe(publishTime);
            expect(result[0].properties.end).toBe(verificationTime);
            expect(result[0].properties.verified).toBe(true);

            // Second feature: unverified period
            expect(result[1].properties.start).toBe(verificationTime);
            expect(result[1].properties.end).toBe(expireTime);
            expect(result[1].properties.verified).toBe(false);
        });
    });

    describe('date normalization', () => {
        it('preserves original timestamps when normalizeDates is false', () => {
            const publishTime = 1704067200000; // Specific timestamp
            const alerts = [createMockAlert({ PublishTime: publishTime })];
            const result = createGeoJSONFromAlerts(alerts, false);
            expect(result[0].properties.start).toBe(publishTime);
        });

        it('normalizes timestamps to same day when normalizeDates is true', () => {
            // Create alerts from different days
            const day1 = new Date('2024-01-15T10:30:00').getTime();
            const day2 = new Date('2024-01-16T10:30:00').getTime();

            const alerts = [
                createMockAlert({ PublishTime: day1, ExpireTime: day1 + 3600000 }),
                createMockAlert({ PublishTime: day2, ExpireTime: day2 + 3600000 }),
            ];

            const result = createGeoJSONFromAlerts(alerts, true);

            // Both should have same date portion (today) but preserve the hour
            const start1 = new Date(result[0].properties.start);
            const start2 = new Date(result[1].properties.start);

            expect(start1.getHours()).toBe(10);
            expect(start2.getHours()).toBe(10);
            expect(start1.getDate()).toBe(start2.getDate());
        });
    });

    describe('default values', () => {
        it('uses "Unknown Location" for missing street', () => {
            const alerts = [createMockAlert({ Street: '' })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result[0].properties.street).toBe('Unknown Location');
        });

        it('uses "Unknown City" for missing city', () => {
            const alerts = [createMockAlert({ City: '' })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result[0].properties.city).toBe('Unknown City');
        });

        it('handles null street', () => {
            const alerts = [createMockAlert({ Street: null })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result[0].properties.street).toBe('Unknown Location');
        });
    });

    describe('multiple alerts', () => {
        it('processes multiple alerts correctly', () => {
            const alerts = [
                createMockAlert({ UUID: 'alert-1' }),
                createMockAlert({ UUID: 'alert-2' }),
                createMockAlert({ UUID: 'alert-3' }),
            ];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result).toHaveLength(3);
        });

        it('creates mixed verified and unverified features', () => {
            const alerts = [
                createMockAlert({ UUID: 'unverified-1', LastVerificationMillis: null }),
                createMockAlert({ UUID: 'verified-1', LastVerificationMillis: Date.now() }),
                createMockAlert({ UUID: 'unverified-2', LastVerificationMillis: null }),
            ];
            const result = createGeoJSONFromAlerts(alerts);
            // 1 feature for each unverified, 2 for verified = 4 total
            expect(result).toHaveLength(4);
        });
    });
});

describe('GeoJSON FeatureCollection creation', () => {
    it('creates valid FeatureCollection structure', () => {
        const alerts = [createMockAlert(), createMockAlert()];
        const features = createGeoJSONFromAlerts(alerts);

        const featureCollection = {
            type: 'FeatureCollection',
            features: features,
        };

        expect(featureCollection.type).toBe('FeatureCollection');
        expect(Array.isArray(featureCollection.features)).toBe(true);
        expect(featureCollection.features).toHaveLength(2);
    });
});

describe('Subtype color mapping', () => {
    // Test that different subtypes would get different visual treatment
    const subtypeTests = [
        'POLICE_VISIBLE',
        'POLICE_WITH_MOBILE_CAMERA',
        'POLICE_HIDING',
        'POLICE_ON_BRIDGE',
        'POLICE_MOTORCYCLIST',
        '', // General police alert
    ];

    subtypeTests.forEach(subtype => {
        it(`correctly preserves subtype: "${subtype || 'empty'}"`, () => {
            const alerts = [createMockAlert({ Subtype: subtype })];
            const result = createGeoJSONFromAlerts(alerts);
            expect(result[0].properties.subtype).toBe(subtype);
        });
    });
});

describe('Bounds calculation helper', () => {
    it('extracts coordinates for bounds calculation', () => {
        const alerts = [
            createMockAlert({ LocationGeo: { latitude: -33.0, longitude: 150.0 } }),
            createMockAlert({ LocationGeo: { latitude: -34.0, longitude: 151.0 } }),
            createMockAlert({ LocationGeo: { latitude: -35.0, longitude: 149.0 } }),
        ];

        const coordinates = alerts
            .filter(alert => {
                const lat = alert.LocationGeo.latitude;
                const lng = alert.LocationGeo.longitude;
                return lat && lng && !isNaN(lat) && !isNaN(lng);
            })
            .map(alert => [alert.LocationGeo.latitude, alert.LocationGeo.longitude]);

        expect(coordinates).toHaveLength(3);
        
        // Calculate bounds
        const lats = coordinates.map(c => c[0]);
        const lngs = coordinates.map(c => c[1]);
        
        expect(Math.min(...lats)).toBe(-35.0);
        expect(Math.max(...lats)).toBe(-33.0);
        expect(Math.min(...lngs)).toBe(149.0);
        expect(Math.max(...lngs)).toBe(151.0);
    });
});


