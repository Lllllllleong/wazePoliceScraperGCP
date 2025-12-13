# Cloud Run Service Module

This module creates a Google Cloud Run v2 service with sensible defaults for the Waze Police Scraper project.

## Features

- Configurable CPU and memory limits
- Automatic scaling configuration
- Environment variable support
- Custom labels
- IAM policy for public/private access
- Startup probes for health checking
- Service account assignment

## Usage

```hcl
module "scraper_service" {
  source = "../../modules/cloud-run"
  
  service_name    = "scraper-service"
  project_id      = var.project_id
  location        = var.region
  container_image = var.scraper_image
  
  cpu_limit    = "1"
  memory_limit = "512Mi"
  
  service_account_email = var.service_account_email
  
  env_vars = {
    GCP_PROJECT_ID       = var.project_id
    FIRESTORE_COLLECTION = var.firestore_collection
    GCS_BUCKET_NAME      = var.gcs_archive_bucket
  }
  
  labels = {
    environment = "prod"
    managed-by  = "terraform"
  }
}
```

## Requirements

- Cloud Run API enabled
- Service account with appropriate permissions
- Container image in Artifact Registry or Container Registry

## Inputs

See [variables.tf](variables.tf) for all available inputs.

## Outputs

See [outputs.tf](outputs.tf) for all available outputs.
