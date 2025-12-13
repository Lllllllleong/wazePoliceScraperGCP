# Variable values for production environment
# Note: Update container image tags to match your latest deployed versions

project_id         = "wazepolicescrapergcp"
region             = "us-central1"
environment        = "prod"
gcs_archive_bucket = "wazepolicescrapergcp-archive"

# Container images - update these with your current commit SHA
# You can find these in your Cloud Run services or GitHub Actions outputs
# TIP: Use 'latest' tag for easier management, or commit SHA for immutability
scraper_image = "us-central1-docker.pkg.dev/wazepolicescrapergcp/scraper-service/scraper-service:0fea809cafcadc74b505bd22d4d28c5ba465745b"
alerts_image  = "us-central1-docker.pkg.dev/wazepolicescrapergcp/alerts-service/alerts-service:0fea809cafcadc74b505bd22d4d28c5ba465745b"
archive_image = "us-central1-docker.pkg.dev/wazepolicescrapergcp/archive-service/archive-service:0fea809cafcadc74b505bd22d4d28c5ba465745b"

# Access Control
bigquery_data_owners = ["chanleongyin@gmail.com"]
