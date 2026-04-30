/**
 * Property-based tests for CSV export round-trip correctness.
 *
 * **Validates: Requirements 16.2**
 */
import { describe, it, expect } from 'vitest';
import fc from 'fast-check';
import { generateCSVContent, parseCSV } from '../export';

describe('Property 14: CSV 导出往返正确性', () => {
  /**
   * Generates a valid CSV header string: non-empty, no commas, quotes, or newlines.
   * Excludes JavaScript reserved property names that interfere with object key access.
   */
  const reservedKeys = new Set(['__proto__', 'constructor', 'toString', 'valueOf', 'hasOwnProperty', 'toLocaleString', 'isPrototypeOf', 'propertyIsEnumerable']);
  const safeHeaderArb = fc.string({ minLength: 1, maxLength: 20 }).filter(
    (s) => !s.includes(',') && !s.includes('"') && !s.includes('\n') && !s.includes('\r') && s.trim().length > 0 && !reservedKeys.has(s)
  );

  /**
   * Generates an array of unique headers.
   */
  const headersArb = fc.uniqueArray(safeHeaderArb, { minLength: 1, maxLength: 8 });

  it('round-trips simple string values correctly', () => {
    fc.assert(
      fc.property(
        headersArb.chain((headers) => {
          // Generate records with simple safe values (no special chars)
          const safeValueArb = fc.string({ minLength: 0, maxLength: 30 }).filter(
            (s) => !s.includes(',') && !s.includes('"') && !s.includes('\n') && !s.includes('\r')
          );
          const recordArb = fc.record(
            Object.fromEntries(headers.map((h) => [h, safeValueArb]))
          );
          const recordsArb = fc.array(recordArb, { minLength: 0, maxLength: 10 });
          return recordsArb.map((records) => ({ headers, records }));
        }),
        ({ headers, records }) => {
          const csv = generateCSVContent(headers, records);
          const parsed = parseCSV(csv);

          // Headers should match
          expect(parsed.headers).toEqual(headers);

          // Records should match
          expect(parsed.records.length).toBe(records.length);
          for (let i = 0; i < records.length; i++) {
            for (const header of headers) {
              const originalValue = records[i][header] as string;
              expect(parsed.records[i][header]).toBe(originalValue);
            }
          }
        }
      ),
      { numRuns: 100 }
    );
  });

  it('round-trips values containing commas, quotes, and newlines correctly', () => {
    fc.assert(
      fc.property(
        headersArb.chain((headers) => {
          // Generate values that may contain special CSV characters
          const specialValueArb = fc.oneof(
            fc.string({ minLength: 0, maxLength: 20 }),
            fc.constant('value,with,commas'),
            fc.constant('value "with" quotes'),
            fc.constant('value\nwith\nnewlines'),
            fc.constant('mixed,"\nspecial'),
            fc.constant(''),
            fc.array(
              fc.oneof(
                fc.constant('a'),
                fc.constant('b'),
                fc.constant(' '),
                fc.constant(','),
                fc.constant('"'),
                fc.constant('\n')
              ),
              { minLength: 1, maxLength: 15 }
            ).map((chars) => chars.join(''))
          );
          const recordArb = fc.record(
            Object.fromEntries(headers.map((h) => [h, specialValueArb]))
          );
          const recordsArb = fc.array(recordArb, { minLength: 1, maxLength: 5 });
          return recordsArb.map((records) => ({ headers, records }));
        }),
        ({ headers, records }) => {
          const csv = generateCSVContent(headers, records);
          const parsed = parseCSV(csv);

          // Headers should match
          expect(parsed.headers).toEqual(headers);

          // Records count should match
          expect(parsed.records.length).toBe(records.length);

          // Each field value should round-trip correctly
          for (let i = 0; i < records.length; i++) {
            for (const header of headers) {
              const originalValue = String(records[i][header] ?? '');
              expect(parsed.records[i][header]).toBe(originalValue);
            }
          }
        }
      ),
      { numRuns: 100 }
    );
  });

  it('round-trips numeric values as strings correctly', () => {
    fc.assert(
      fc.property(
        headersArb.chain((headers) => {
          // Mix of string and numeric values
          const valueArb = fc.oneof(
            fc.string({ minLength: 0, maxLength: 15 }).filter(
              (s) => !s.includes(',') && !s.includes('"') && !s.includes('\n') && !s.includes('\r')
            ),
            fc.integer().map(String),
            fc.float({ noNaN: true, noDefaultInfinity: true }).map(String)
          );
          const recordArb = fc.record(
            Object.fromEntries(headers.map((h) => [h, valueArb]))
          );
          const recordsArb = fc.array(recordArb, { minLength: 0, maxLength: 8 });
          return recordsArb.map((records) => ({ headers, records }));
        }),
        ({ headers, records }) => {
          const csv = generateCSVContent(headers, records);
          const parsed = parseCSV(csv);

          expect(parsed.headers).toEqual(headers);
          expect(parsed.records.length).toBe(records.length);

          for (let i = 0; i < records.length; i++) {
            for (const header of headers) {
              const originalValue = String(records[i][header] ?? '');
              expect(parsed.records[i][header]).toBe(originalValue);
            }
          }
        }
      ),
      { numRuns: 100 }
    );
  });
});
