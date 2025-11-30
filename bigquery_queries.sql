-- BigQuery Setup and Query Examples for Police Alerts Archive
-- This file contains SQL commands for loading and querying the archive data

-- =============================================================================
-- 1. CREATE TABLE (if loading from scratch)
-- =============================================================================
-- Note: If using `bq load` with the schema.json, this is not needed
-- But this is useful for reference or manual table creation

CREATE TABLE IF NOT EXISTS `your_project.your_dataset.police_alerts` (
  UUID STRING NOT NULL,
  ID STRING,
  Type STRING,
  Subtype STRING,
  Street STRING,
  City STRING,
  Country STRING,
  LocationGeo STRUCT<latitude FLOAT64, longitude FLOAT64>,
  Reliability INT64,
  Confidence INT64,
  ReportRating INT64,
  PublishTime TIMESTAMP,
  ScrapeTime TIMESTAMP,
  ExpireTime TIMESTAMP,
  LastVerificationTime TIMESTAMP,
  ActiveMillis INT64,
  LastVerificationMillis INT64,
  NThumbsUpInitial INT64,
  NThumbsUpLast INT64,
  RawDataInitial STRING,
  RawDataLast STRING
);

-- =============================================================================
-- 2. ADD GEOGRAPHY COLUMN (Recommended!)
-- =============================================================================
-- After loading data, add a computed GEOGRAPHY column for geospatial queries
ALTER TABLE `your_project.your_dataset.police_alerts` 
ADD COLUMN IF NOT EXISTS location_point GEOGRAPHY AS (
  ST_GEOGPOINT(LocationGeo.longitude, LocationGeo.latitude)
);

-- =============================================================================
-- 3. LOAD DATA from Cloud Storage
-- =============================================================================
-- Command line example (run in terminal):
-- bq load --source_format=NEWLINE_DELIMITED_JSON \
--   your_project:your_dataset.police_alerts \
--   gs://your-bucket/2025-11-10.jsonl \
--   bigquery_schema.json

-- =============================================================================
-- 4. EXAMPLE QUERIES
-- =============================================================================

-- Basic query with geography
SELECT 
  UUID,
  Street,
  City,
  Subtype,
  ST_GEOGPOINT(LocationGeo.longitude, LocationGeo.latitude) as location,
  PublishTime,
  ExpireTime,
  ActiveMillis / 1000 / 60 as active_minutes,
  NThumbsUpLast
FROM `your_project.your_dataset.police_alerts`
WHERE DATE(PublishTime) = '2025-11-10'
ORDER BY PublishTime DESC
LIMIT 100;

-- Alerts by subtype with average duration
SELECT 
  Subtype,
  COUNT(*) as alert_count,
  AVG(ActiveMillis / 1000 / 60) as avg_active_minutes,
  AVG(NThumbsUpLast) as avg_verifications
FROM `your_project.your_dataset.police_alerts`
WHERE DATE(PublishTime) >= '2025-11-01'
GROUP BY Subtype
ORDER BY alert_count DESC;

-- Find alerts within 5km of a point (using computed geography column)
SELECT 
  UUID,
  Street,
  City,
  ST_DISTANCE(location_point, ST_GEOGPOINT(149.1300, -35.2809)) / 1000 as distance_km,
  PublishTime
FROM `your_project.your_dataset.police_alerts`
WHERE ST_DWITHIN(location_point, ST_GEOGPOINT(149.1300, -35.2809), 5000) -- 5km radius
  AND DATE(PublishTime) = '2025-11-10'
ORDER BY distance_km;

-- Hourly distribution of alerts
SELECT 
  EXTRACT(HOUR FROM PublishTime) as hour_of_day,
  COUNT(*) as alert_count
FROM `your_project.your_dataset.police_alerts`
WHERE DATE(PublishTime) >= '2025-11-01'
GROUP BY hour_of_day
ORDER BY hour_of_day;

-- Top streets for police presence
SELECT 
  Street,
  COUNT(*) as alert_count,
  COUNT(DISTINCT DATE(PublishTime)) as days_active,
  AVG(ActiveMillis / 1000 / 60) as avg_duration_minutes
FROM `your_project.your_dataset.police_alerts`
WHERE Street != ''
  AND DATE(PublishTime) >= '2025-11-01'
GROUP BY Street
HAVING alert_count >= 5
ORDER BY alert_count DESC
LIMIT 20;

-- Community engagement analysis
SELECT 
  CASE 
    WHEN ActiveMillis < 1800000 THEN '< 30 min'
    WHEN ActiveMillis < 3600000 THEN '30-60 min'
    WHEN ActiveMillis < 7200000 THEN '1-2 hours'
    WHEN ActiveMillis < 14400000 THEN '2-4 hours'
    ELSE '4+ hours'
  END as duration_bucket,
  AVG(NThumbsUpLast) as avg_verifications,
  COUNT(*) as alert_count
FROM `your_project.your_dataset.police_alerts`
WHERE DATE(PublishTime) >= '2025-11-01'
GROUP BY duration_bucket
ORDER BY 
  CASE duration_bucket
    WHEN '< 30 min' THEN 1
    WHEN '30-60 min' THEN 2
    WHEN '1-2 hours' THEN 3
    WHEN '2-4 hours' THEN 4
    ELSE 5
  END;
