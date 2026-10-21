import { describe, expect, it } from 'vitest';
import { compass, fmtDate, fmtDateRange, fmtDateTime, fmtDuration, fmtHour, fmtInt, fmtPercent } from './format.ts';

describe('format', () => {
  it('formats numbers', () => {
    expect(fmtInt(241717)).toBe('241,717');
    expect(fmtPercent(28981 / 241717)).toBe('12.0%');
  });

  it('formats hours on a 12-hour clock', () => {
    expect([0, 1, 11, 12, 15, 23].map(fmtHour)).toEqual(['12 am', '1 am', '11 am', '12 pm', '3 pm', '11 pm']);
  });

  it('formats durations', () => {
    expect(fmtDuration(30)).toBe('30 min');
    expect(fmtDuration(60)).toBe('1 h');
    expect(fmtDuration(176)).toBe('2 h 56 min');
  });

  it('names compass points', () => {
    expect(compass(0)).toBe('N');
    expect(compass(249)).toBe('WSW');
    expect(compass(359)).toBe('N');
    expect(compass(-90)).toBe('W');
  });

  it('formats dates and ranges', () => {
    expect(fmtDate('2025-12-19')).toBe('19 Dec 2025');
    expect(fmtDateRange('2025-12-19', '2026-01-04')).toBe('19 Dec 2025 – 4 Jan 2026');
    expect(fmtDateRange('2025-11-03', '2025-11-09')).toBe('3 Nov – 9 Nov 2025');
    expect(fmtDateRange('2025-11-03', '2025-11-03')).toBe('3 Nov 2025');
  });

  it('formats instants in Sydney time', () => {
    expect(fmtDateTime(Date.parse('2025-10-13T04:55:00Z'), 'Australia/Sydney')).toBe('Mon 13 Oct 2025, 3:55 pm');
  });
});
