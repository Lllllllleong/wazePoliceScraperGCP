# Terraform Production Environment

This directory contains the Terraform configuration for the **production** environment of the Waze Police Scraper project.

## Overview

All GCP infrastructure is managed declaratively using Terraform, including:
- Cloud Run services (scraper, alerts, archive)
- Firestore database
- Cloud Storage buckets
- Cloud Scheduler jobs
- IAM service accounts and permissions
- BigQuery dataset

## Directory Structure

```
prod/
├── backend.tf           # Remote state configuration (GCS)
├── versions.tf          # Terraform and provider version constraints
├── main.tf              # Main infrastructure configuration
├── variables.tf         # Input variable declarations
├── terraform.tfvars     # Variable values (update image tags as needed)
└── outputs.tf           # Output value definitions
```

## Reusable Modules

The infrastructure uses modular Terraform configurations located in `../../modules/`:

- **cloud-run**: Cloud Run service configuration
- **firestore**: Firestore database setup
- **storage**: GCS bucket configuration
- **scheduler**: Cloud Scheduler job definitions
- **service-account**: IAM service account management
- **bigquery**: BigQuery dataset configuration
- **artifact-registry**: Container registry setup

## Usage

```bash
# Navigate to this directory
cd terraform/environments/prod

# Initialize Terraform (downloads providers, configures backend)
terraform init

# Preview infrastructure changes
terraform plan

# Apply changes
terraform apply
```

## Remote State

Terraform state is stored remotely in Google Cloud Storage:
- Bucket: `wazepolicescrapergcp-terraform-state`
- Path: `terraform/prod/state`
- Locking: Enabled via GCS

## Updating Container Images

When deploying new versions of the services, update the image tags in `terraform.tfvars`:

```hcl
scraper_image = "us-central1-docker.pkg.dev/PROJECT/scraper-service/scraper-service:COMMIT_SHA"
alerts_image  = "us-central1-docker.pkg.dev/PROJECT/alerts-service/alerts-service:COMMIT_SHA"
archive_image = "us-central1-docker.pkg.dev/PROJECT/archive-service/archive-service:COMMIT_SHA"
```

## CI/CD Integration

The Terraform workflow (`.github/workflows/terraform-ci-cd.yml`) automatically:
1. Runs `terraform plan` on pull requests
2. Posts plan output as PR comments
3. Runs `terraform apply` on merge to main

## Security Notes

- Service accounts follow the principle of least privilege
- Each service has dedicated IAM permissions
- No default service accounts are used
- Workload Identity Federation is used for GitHub Actions authentication
