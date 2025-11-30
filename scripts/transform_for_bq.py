#!/usr/bin/env python3
"""
Transform archive JSONL files for BigQuery import by converting LocationGeo 
to WKT format for GEOGRAPHY type.
"""
import json
import sys

def transform_record(record):
    """Transform a single record for BigQuery import."""
    if 'LocationGeo' in record and record['LocationGeo']:
        lat = record['LocationGeo']['latitude']
        lon = record['LocationGeo']['longitude']
        # Convert to WKT format: "POINT(longitude latitude)"
        record['LocationGeo'] = f"POINT({lon} {lat})"
    return record

def main():
    """Read JSONL from stdin, transform, write to stdout."""
    for line in sys.stdin:
        if line.strip():
            record = json.loads(line)
            transformed = transform_record(record)
            print(json.dumps(transformed))

if __name__ == '__main__':
    main()
