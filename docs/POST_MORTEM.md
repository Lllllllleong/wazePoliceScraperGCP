# Post-Mortem: Waze Police Scraper GCP

**Data collection:** September 26, 2025 – March 16, 2026

**Decommissioned:** April 6, 2026

**Written:** April 6, 2026

---

## What the Project Was

A pipeline that scraped police alert data from the Waze live traffic API across the Sydney–Canberra corridor (Hume Highway), stored it in Firestore, archived it daily to GCS, and served it via a JavaScript dashboard. Started from curiosity during regular drives between Sydney and Canberra.

The system ran three microservices on Cloud Run: a scraper (triggered every minute by Cloud Scheduler), an archive service (triggered daily), and an alerts API (public HTTPS endpoint for the frontend). Infrastructure was defined in Terraform; all services had CI/CD pipelines via GitHub Actions.

---

## Timeline

| Date | Event |
|---|---|
| Sep 26, 2025 | First data collected |
| Jan 9, 2026 | Last GCS archive created before the archive service bug |
| ~Jan 10, 2026 | Archive service begins failing with 500 on every run due to `IsObjectNotExist` bug (#18); scraper continues collecting full data into Firestore |
| Jan 10 – Mar 16, 2026 | ~66 days of Firestore data accumulates with no corresponding GCS archives |
| Mar 16, 2026 | Waze API begins returning 403 on all requests; final Firestore document timestamped 11:05 PM AEDT; data collection permanently stops |
| Mar–Apr 2026 | #18 fixed (PR #22); backfill script run to create missing archives for Jan 10 – Mar 16 (#20, PR #23) |
| Apr 5, 2026 | All historical data confirmed archived to GCS |
| Apr 6, 2026 | Scraper and archive services decommissioned (#25) |

---

## What Went Well

**Firestore as a safety net.** Even though the archive service was broken for ~66 days, all data remained safe in Firestore. The backfill script was able to recover the full Jan 10 – Mar 16 window before any teardown.

**GCS archival strategy.** Pre-computing daily archives in GCS reduced Firestore read costs and enabled the alerts service to stream JSONL responses with parallel worker goroutines rather than issuing repeated Firestore queries per request. The architecture paid off — once archives existed for all dates, the alerts service required no changes to serve the full historical dataset.

**Terraform for infrastructure.** Having all GCP resources defined in Terraform made the decommission clean — removing modules from `main.tf` was sufficient to express the desired end state. Dependency ordering (schedulers before services) was handled automatically.

**CI/CD pipelines.** The reusable workflow pattern kept the three service pipelines consistent. Coverage thresholds caught regressions early.

**Firebase Anonymous Authentication + rate limiting.** The alerts API was never abused despite being publicly discoverable. Anon auth with per-user token-bucket rate limiting held up fine at this scale.

**Serverless economics.** The entire system ran at near-zero cost during active collection (~$5–10/month). After decommission, the remaining infrastructure (alerts API on Cloud Run, Firebase Hosting, Firestore, GCS) costs effectively nothing.

---

## What Went Wrong

**Archive service `IsObjectNotExist` bug (#18).** The archive service's idempotency check used `==` pointer comparison and an exact string match to detect GCS "object not found" errors. With GCS SDK v1.57.0, these errors are wrapped, breaking both checks silently. The result was a 500 on every daily run for ~66 days, putting 66 days of Firestore data at risk of being lost before the bug was caught. The fix was a one-line change to `errors.Is`.

**No alerting on sustained service failures.** The archive service was returning 500 every day from January 10, with Cloud Scheduler logging retries each time — but no alert fired. The failure was only discovered during manual investigation. A simple alert on repeated Cloud Scheduler job failures would have surfaced this the same day it started.

**Dependency on an undocumented, unofficial API.** The Waze endpoint was never officially documented or offered as a public API. There was no terms of service, no versioning guarantee, and no deprecation notice. When Waze changed something server-side in March 2026, every request began returning 403 with no explanation and no recourse.

---

## Lessons Learned

**Wrap error comparisons with `errors.Is`, not `==`.** Sentinel error comparisons with `==` break silently when errors are wrapped, which is common across Go SDKs. `errors.Is` unwraps the chain correctly and should be the default.

**Alert on sustained job failures, not just individual ones.** A single failed run is noise. Daily consecutive failures for weeks is a critical signal. The logs and retry history were there — a monitoring alert would have closed the feedback loop from weeks to hours.

**Treat unofficial APIs as ephemeral from day one.** Any pipeline built on an undocumented endpoint should preserve whatever has already been collected and fail gracefully when the source disappears. Firestore holding the full dataset meant nothing was lost when the Waze API went down permanently — but that was a property of the system design, not something explicitly planned for this scenario.

---

## Outcome

The [dashboard](https://dashboard.whyhireleong.com) remains live, serving the full historical dataset (Sep 26, 2025 – Mar 16, 2026).
