import { describe, expect, it } from 'vitest';
import { displayRoadName, groupRoads, roadKey } from './roads.ts';

describe('roadKey', () => {
  it('merges motorway and highway spellings of the same route', () => {
    expect(roadKey('M31 - Hume Mwy')).toBe(roadKey('M31 - Hume Hwy'));
  });

  it('drops sign text after " to "', () => {
    expect(roadKey('B69 - Narellan Rd to Campbelltown / University / Hospital')).toBe(roadKey('B69 - Narellan Rd'));
    expect(roadKey('M31 - Hume Hwy to Yass / Canberra')).toBe(roadKey('M31 - Hume Mwy'));
  });

  it('keeps different routes apart even with the same name', () => {
    expect(roadKey('A1 - Princes Hwy')).not.toBe(roadKey('M1 - Princes Mwy'));
    expect(roadKey('A9 - Narellan Rd')).not.toBe(roadKey('B69 - Narellan Rd'));
  });

  it('keeps different streets on one route apart', () => {
    expect(roadKey('B73 - Old Hume Hwy')).not.toBe(roadKey('B73 - Kangaloon Rd'));
    expect(roadKey('M23 - Federal Hwy')).not.toBe(roadKey('M23 - Majura Pwy'));
  });

  it('leaves names without a route number alone', () => {
    expect(roadKey('Camden Valley Way')).toBe('Camden Valley Way');
    expect(roadKey('to M31 - Hume Mwy to Sydney')).toBe('to M31 - Hume Mwy to Sydney');
    expect(roadKey('')).toBe('');
  });
});

describe('displayRoadName', () => {
  it('uses an en dash after the route number', () => {
    expect(displayRoadName('M31 - Hume Mwy')).toBe('M31 – Hume Mwy');
    expect(displayRoadName('Main St')).toBe('Main St');
    expect(displayRoadName('')).toBe('No road name');
  });
});

describe('groupRoads', () => {
  it('orders groups by alert count and names them after the commonest variant', () => {
    const streets = ['Main St', 'M31 - Hume Hwy', 'M31 - Hume Mwy'];
    const { roads, streetToRoad } = groupRoads(streets, [50, 30, 40]);

    expect(roads).toEqual([
      { name: 'M31 – Hume Mwy', variants: ['M31 - Hume Mwy', 'M31 - Hume Hwy'] },
      { name: 'Main St', variants: ['Main St'] },
    ]);
    expect([...streetToRoad]).toEqual([1, 0, 0]);
  });
});
