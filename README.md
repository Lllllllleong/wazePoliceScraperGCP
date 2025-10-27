# Waze Police Scraper GCP

A comprehensive cloud-based system for scraping, storing, and analyzing police alerts from Waze's live traffic data. The system collects police presence alerts along major highways (Sydney to Canberra - Hume Highway) and provides an interactive web interface for analyzing and visualizing the data.

## 🎯 Overview

This project consists of three main components:

1. **Scraper Service** - Cloud Run service that fetches police alerts from Waze API on a scheduled basis
2. **Data Analysis Dashboard** - Interactive web interface for visualizing and analyzing police alert patterns

All data is stored in Google Cloud Firestore, providing reliable, scalable storage with real-time capabilities.

## 🏗️ Architecture

```
┌─────────────────────┐
│  Cloud Scheduler    │──triggers──┐
└─────────────────────┘            │
                                   ▼
┌─────────────────────┐      ┌──────────────────────┐
│   Scraper Service   │──────▶│   Cloud Firestore   │
│   (Cloud Run)       │      │  (police_alerts)     │
└─────────────────────┘      └──────────────────────┘
                                        │
                                        │ queries
                                        ▼
┌─────────────────────┐      ┌──────────────────────┐
│   Web Dashboard      │─────▶│   Cloud Firestore   │
│  (React/Vanilla JS)  │      │  (police_alerts)     │
└─────────────────────┘      └──────────────────────┘
```

## 📦 Components

### 1. Scraper Service (`cmd/scraper/`)


- **Trigger**: Cloud Scheduler (configurable intervals)
- **Coverage**: Sydney to Canberra via Hume Highway (4 bounding boxes)

  - Multiple bounding box support
  - Duplicate alert detection
  - Automatic retry logic


### 3. Data Exporter (`cmd/exporter/`)

- **Purpose**: Command-line tool for exporting alerts to JSON/JSONL
- **Usage**: Export alerts for a specific date range
- **Output Formats**: JSON or JSONL (newline-delimited JSON)

### 4. Data Analysis Dashboard

Two implementations available:

#### Vanilla JavaScript Version (`dataAnalysis/public/`)
- Production-ready dashboard
- Interactive map with Leaflet
- Timeline controls for temporal visualization
- Multi-date selection (up to 14 dates)

- Statistics panel
- Export capabilities

#### React Prototype (`dataAnalysis/react-prototype/`)
- Modern React + TypeScript implementation
- Tailwind CSS styling
- Component-based architecture
- Dark theme with gradient accents
- **Status**: UI skeleton (non-functional prototype)

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Google Cloud Platform account
- GCP Project with enabled APIs:
  - Cloud Run API
  - Cloud Firestore API
  - Cloud Scheduler API
  - Cloud Build API

### Environment Variables

Create a `.env` file in the project root:

```bash
GCP_PROJECT_ID=your-project-id
FIRESTORE_COLLECTION=police_alerts
WAZE_BBOXES="150.388,-34.255,151.009,-33.938;149.589,-34.769,150.830,-34.139"
```

### Local Development

1. **Install dependencies**:
```bash
go mod download
```

2. **Run the scraper locally**:
```bash
export GCP_PROJECT_ID=your-project-id
export FIRESTORE_COLLECTION=police_alerts
go run cmd/scraper/main.go
```

4. **Export data**:
```bash
go run cmd/exporter/main.go \
  -project your-project-id \
  -collection police_alerts \
  -start 2025-10-01 \
  -end 2025-10-15 \
  -output alerts.jsonl \
  -format jsonl
```

### Deployment

#### Deploy Scraper Service

```bash
./scripts/deploy.sh
```



#### Set up Cloud Scheduler

```bash
./scripts/setup-scheduler.sh
```

This creates a Cloud Scheduler job that triggers the scraper every hour.

#### Deploy Firestore Security Rules

```bash
./scripts/deploy-firestore-security.sh
```

Or manually:
```bash
firebase deploy --only firestore:rules
```

### Deploying the Dashboard

The dashboard can be deployed to:
- Firebase Hosting
- Cloud Storage + Cloud CDN
- Any static hosting service

For Firebase Hosting:
```bash
cd dataAnalysis
firebase deploy --only hosting
```

## 📊 Data Model

### WazeAlert Structure

```go
{
  "uuid": "unique-alert-id",
  "type": "POLICE",

  "location": {
    "latitude": -34.123,
    "longitude": 151.456
  },
  "street": "Hume Highway",
  "city": "Sydney",
  "country": "AU",
  "pub_millis": 1698765432000,
  "reliability": 5,
  "confidence": 3,
  "report_rating": 0,
  "report_by": "username",
  "is_thumb_up": true,
  "scrape_time": "2025-10-21T10:00:00Z",
  "scrape_date": "2025-10-21"
}
```

### Firestore Collections

- **Collection**: `police_alerts`
- **Document ID**: `{scrape_date}_{uuid}`


## 🔧 Configuration

### Bounding Boxes

Configure geographic areas to scrape in `configs/bboxes.yaml`:

```yaml
bboxes:
  - name: "Hume Highway - Sydney"
    bbox: "150.388,-34.255,151.009,-33.938"
    description: "Sydney section of Hume Highway"
```

Format: `"west,south,east,north"` (longitude_min, latitude_min, longitude_max, latitude_max)

Use [BoundingBox Tool](https://boundingbox.klokantech.com/) to generate coordinates.

### Dashboard Configuration

Edit `dataAnalysis/public/config.js`:

```javascript
const CONFIG = {
    MIN_DATE: '2025-09-26',
    MAX_SELECTABLE_DATES: 14
};
```

## 🔒 Security

See `SECURITY_CONFIG.md` for comprehensive security guidelines including:

- Cloud Run instance limits and concurrency settings
- Rate limiting and authentication
- Firestore security rules
- Ingress controls
- Monitoring and alerts


### Firestore Security Rules

The project includes production-ready security rules:
- Read-only access for authenticated users
- No write access from clients
- Only server-side writes via service accounts

## 📈 Features

### Scraper
- ✅ Multiple bounding box support
- ✅ Automatic deduplication by UUID
- ✅ Configurable scrape intervals
- ✅ Error handling and retry logic
- ✅ Structured logging

### Dashboard
- ✅ Interactive Leaflet map
- ✅ Timeline visualization
- ✅ Multi-date selection (up to 14 dates)
- ✅ Real-time filtering
- ✅ Statistics dashboard
- ✅ Export to CSV/JSON
- ✅ Responsive design
- ✅ Disclaimer modal

## 🛠️ Development

### Project Structure

```
.
├── cmd/                          # Executable commands
│   ├── scraper/                  # Scraper service
│   └── exporter/                 # Data export tool
├── internal/                     # Internal packages
│   ├── models/                   # Data models
│   ├── storage/                  # Firestore client
│   └── waze/                     # Waze API client
├── dataAnalysis/                 # Frontend dashboard
│   ├── public/                   # Vanilla JS version
│   └── react-prototype/          # React version
├── configs/                      # Configuration files
├── scripts/                      # Deployment scripts
├── docs/                         # Documentation
├── Dockerfile                    # Scraper container
├── firestore.rules              # Firestore security rules
└── firestore.indexes.json       # Firestore indexes
```

### Testing

Test the scraper endpoint:
```bash
curl http://localhost:8080/
```

Test HTML files:


## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is for educational and research purposes. Please respect Waze's Terms of Service when using their API.

## 🙏 Acknowledgments

- Waze for providing real-time traffic data
- Google Cloud Platform for hosting infrastructure
- Leaflet.js for mapping capabilities
- The open-source community

## 📞 Support

For issues, questions, or contributions:
- Open an issue on GitHub
- Check existing documentation in the `docs/` folder
- Review security guidelines in `SECURITY_CONFIG.md`

## 🗺️ Roadmap

- [ ] Add more geographic regions
- [ ] Machine learning for pattern detection
- [ ] Mobile app development
- [ ] Real-time notifications
- [ ] Historical trend analysis
- [ ] Complete React prototype implementation
- [ ] Implement rate limiting
- [ ] Add unit tests
- [ ] CI/CD pipeline

---

**Made with ❤️ for safer roads**
