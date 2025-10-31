# Backend Technical Highlights

**Professional Go Microservices Architecture on Google Cloud Platform**

This document highlights the key backend features and technical implementations that demonstrate professional software engineering capabilities, suitable for showcasing to senior developers and hiring managers.

---

## 🏗️ **Architecture & Design Patterns**

### Microservices Architecture
- **Three Independent Services**: Clean separation of concerns with `scraper-service`, `alerts-service`, and `archive-service`
- **Domain-Driven Design**: Each service owns its specific business logic and data operations
- **Containerized Deployment**: Docker multi-stage builds for minimal image sizes and production-ready containers
- **Cloud-Native**: Serverless deployment on Google Cloud Run with auto-scaling and scale-to-zero capabilities

### Modular Code Organization
```
internal/
├── models/      # Domain models with rich type definitions
├── storage/     # Data access layer with clean abstractions
└── waze/        # External API client encapsulation
```
- **Separation of Concerns**: Clear boundaries between API handlers, business logic, and data access
- **Dependency Injection**: Services receive configured clients rather than creating their own
- **Package Structure**: Follows Go best practices with `cmd/` for executables and `internal/` for libraries

---

## ⚡ **Performance & Concurrency**

### Concurrent Data Processing (`alerts-service`)
- **Worker Pool Pattern**: Implements a configurable worker pool (7 concurrent workers) to parallelize date-based queries
- **Channels for Communication**: Uses buffered channels for job distribution and data streaming
  ```go
  jobs := make(chan time.Time, len(dates))
  dataChan := make(chan []byte, 100)
  ```
- **WaitGroup Synchronization**: Properly coordinates goroutine lifecycle with `sync.WaitGroup`
- **Single Writer Pattern**: Dedicated writer goroutine prevents race conditions on HTTP response writer

### HTTP Streaming Response
- **JSONL Streaming**: Efficiently streams large datasets line-by-line instead of buffering entire response
- **Chunked Transfer Encoding**: Uses HTTP flusher for real-time data delivery
- **Memory Efficiency**: Processes and sends data incrementally, handling arbitrarily large datasets without memory bloat
- **Line-by-Line Parsing**: Custom buffer management ensures JSON objects aren't split mid-stream

### Intelligent Data Tiering
- **Hot/Cold Storage Strategy**: 
  - Fresh data served from Firestore (fast, expensive)
  - Archived data served from GCS (slow, cheap)
- **Automatic Fallback**: Seamlessly falls back to Firestore when GCS archives don't exist
- **Lazy Archive Creation**: Archives generated on-demand by scheduled jobs, not blocking scraper

---

## 🛡️ **Data Integrity & Reliability**

### Comprehensive Lifecycle Tracking
The `PoliceAlert` model tracks the complete lifecycle of each alert:
```go
type PoliceAlert struct {
    PublishTime  time.Time  // When Waze first published it
    ScrapeTime   time.Time  // When we first discovered it
    ExpireTime   time.Time  // When we last saw it (presumed expired after)
    ActiveMillis int64      // Total time alert was active
    
    // Dual raw data preservation
    RawDataInitial string   // First scrape JSON
    RawDataLast    string   // Most recent scrape JSON
}
```

### Idempotent Operations
- **Archive Service Idempotency**: Checks for existing archives before processing
  ```go
  _, err = obj.Attrs(ctx)
  if err == nil {
      log.Printf("Archive already exists. Skipping.")
      return
  }
  ```
- **Deduplication by UUID**: Prevents duplicate alerts across multiple geographic bounding boxes
- **Update-or-Create Pattern**: Safely handles both new and existing alerts in a single operation

### Error Handling & Resilience
- **Graceful Degradation**: Individual alert failures don't crash the entire scraping operation
- **Detailed Error Context**: Uses `fmt.Errorf` with `%w` for error wrapping and stack preservation
- **Comprehensive Logging**: Structured logging at INFO, WARN, and ERROR levels for observability
- **Resource Cleanup**: Proper `defer` usage for closing connections and releasing resources

---

## 📊 **Advanced Data Modeling**

### Rich Type Definitions
- **Type Safety**: Strongly typed models prevent runtime errors
- **Firestore Tags**: Custom struct tags for database serialization
  ```go
  UUID      string    `firestore:"uuid"`
  PublishTime time.Time `firestore:"publish_time"`
  ```
- **GeoPoint Support**: Native integration with Firestore's geospatial types using `latlng.LatLng`
- **Nullable Fields**: Proper use of pointers for optional data (`*time.Time`, `*int64`)

### Complex Query Patterns
- **Multi-Criteria Queries**: Compound Firestore queries with multiple where clauses
  ```go
  query := collection.
      Where("expire_time", ">=", startDate).
      Where("publish_time", "<=", endDate)
  ```
- **Composite Indexes**: Supports efficient queries requiring custom Firestore indexes (defined in `firestore.indexes.json`)
- **Client-Side Filtering**: Additional filtering (subtypes, streets) applied in-memory when database queries are insufficient

### Data Transformation Pipeline
1. **Ingestion**: Raw Waze JSON → `WazeAlert` struct
2. **Enrichment**: Extract verification data from nested comments array
3. **Persistence**: Transform to `PoliceAlert` with computed fields
4. **Archival**: Serialize to JSONL with full fidelity
5. **Retrieval**: Deserialize and stream to frontend

---

## 🔐 **Security & Best Practices**

### API Security
- **CORS Configuration**: Properly configured CORS headers for cross-origin requests
- **Method Validation**: HTTP method checking (GET, POST) with appropriate error responses
- **Input Validation**: Date format validation with clear error messages
- **No Direct Database Exposure**: Frontend never accesses Firestore directly

### Resource Management
- **Context Propagation**: Proper use of `context.Context` throughout the call chain
- **Connection Pooling**: Firestore and GCS clients properly initialized and reused
- **Defer Cleanup**: All resources (HTTP bodies, database connections) properly closed
  ```go
  defer firestoreClient.Close()
  defer storageClient.Close()
  defer resp.Body.Close()
  ```

### Environment Configuration
- **12-Factor App Principles**: All configuration via environment variables
- **Sensible Defaults**: Fallback values when optional config is missing
- **No Hardcoded Secrets**: Credentials managed via Google Cloud IAM, never in code

---

## 🚀 **CI/CD & DevOps**

### GitHub Actions Workflows
- **Automated Testing**: `go test` runs on every commit
- **Linting**: `golangci-lint` enforces code quality standards
- **Multi-Stage Pipeline**: Build → Test → Lint → Containerize → Deploy
- **Path-Based Triggers**: Services only rebuild when their code changes
  ```yaml
  paths:
    - 'cmd/scraper-service/**'
    - 'internal/**'
    - 'go.mod'
  ```

### Docker Best Practices
- **Multi-Stage Builds**: Separate builder and runtime stages minimize final image size
- **Alpine Base Image**: Minimal attack surface with `alpine:latest`
- **Non-Root User**: Runs as unprivileged user in production
- **Build Caching**: Leverages Docker layer caching and GitHub Actions cache

### Infrastructure as Code
- **Declarative Deployment**: Cloud Run deployments fully defined in YAML
- **Resource Constraints**: Memory and CPU limits explicitly set
  ```yaml
  flags: '--max-instances=1 --min-instances=0 --memory=512Mi --cpu=1'
  ```
- **Workload Identity Federation**: Secure, keyless authentication from GitHub to GCP

---

## 📈 **Scalability Considerations**

### Horizontal Scaling
- **Stateless Services**: All services are stateless and can scale horizontally
- **Concurrent Request Handling**: Go's goroutines enable high concurrent request throughput
- **Auto-Scaling**: Cloud Run automatically scales based on request volume

### Cost Optimization
- **Scale-to-Zero**: Services incur zero cost when not in use
- **Storage Tiering**: Old data moved from expensive Firestore to cheap GCS
- **Efficient Queries**: Indexed queries prevent full collection scans
- **Streaming Responses**: Reduces memory overhead and connection time

### Observability
- **Structured Logging**: Consistent log format for parsing and alerting
- **Metrics Tracking**: `ScrapingStats` struct tracks system health
- **Request Tracing**: Context-based request tracking through the entire stack

---

## 🧪 **Code Quality & Maintainability**

### Clean Code Principles
- **Self-Documenting Code**: Clear variable and function names
- **Single Responsibility**: Each function does one thing well
- **DRY (Don't Repeat Yourself)**: Shared logic extracted into reusable functions
- **Comments Where Needed**: Complex algorithms explained with inline comments

### Go Idioms & Best Practices
- **Error Handling**: Every error checked and handled appropriately
- **Defer for Cleanup**: Resources always released, even on error paths
- **Exported vs Unexported**: Proper use of capitalization for API design
- **Receiver Methods**: Type-based methods for clean API (`fc *FirestoreClient`)

### Testing Readiness
- **Testable Architecture**: Clean dependency injection enables easy mocking
- **Interface Segregation**: Small, focused interfaces (though not explicitly shown, structure supports it)
- **Modular Functions**: Small, pure functions are easier to unit test

---

## 🎯 **Domain Expertise Demonstrated**

### External API Integration
- **Resilient HTTP Clients**: Timeouts, retries, multiple URL fallback strategies
- **User-Agent Spoofing**: Proper headers to mimic browser requests
- **Rate Limiting Awareness**: Designed to respect external API constraints

### Time & Timezone Handling
- **UTC-First Design**: All internal operations in UTC to prevent timezone bugs
- **Explicit Timezone Conversion**: `time.LoadLocation()` for region-specific display
- **Millisecond Precision**: Maintains Waze's millisecond timestamps throughout

### Geographic Data Processing
- **Bounding Box Calculations**: Handles geographic coordinate systems
- **Deduplication Across Regions**: Intelligent handling of overlapping bounding boxes
- **GeoPoint Integration**: Firestore native geospatial queries for location-based searches

---

## 📚 **Documentation & Professional Presentation**

### Comprehensive Documentation
- **`ARCHITECTURE.md`**: System design, data flow, component interaction
- **`DECISIONS.md`**: Architectural Decision Records (ADRs) explaining key choices
- **`README.md`**: Quick start, deployment instructions, tech stack overview
- **Code Comments**: Complex algorithms and business logic well-documented

### Portfolio Quality
- **Live Demo**: Deployed and publicly accessible application
- **Professional README**: Badges, diagrams (Mermaid), clear sections
- **Open Source Ready**: MIT license, contribution-friendly structure

---

## 💡 **Key Takeaways for Hiring Managers**

This backend demonstrates:

✅ **Production-Ready Code**: Not a toy project—this is deployable, scalable, and maintainable  
✅ **Modern Go Development**: Proper use of goroutines, channels, contexts, and error handling  
✅ **Cloud-Native Architecture**: Serverless, containerized, auto-scaling microservices  
✅ **Data Engineering Skills**: ETL pipelines, storage optimization, lifecycle management  
✅ **DevOps Proficiency**: CI/CD, Docker, Infrastructure as Code, automated deployments  
✅ **System Design Thinking**: Performance optimization, cost management, reliability patterns  
✅ **Professional Communication**: Clear documentation, thoughtful architectural decisions  

**This is not just code—this is software craftsmanship.**
