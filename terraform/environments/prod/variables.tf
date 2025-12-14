# Input variables for the production environment

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "Default GCP region for resources"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

# Cloud Run service container images
variable "scraper_image" {
  description = "Container image for scraper-service"
  type        = string
}

variable "alerts_image" {
  description = "Container image for alerts-service"
  type        = string
}

variable "archive_image" {
  description = "Container image for archive-service"
  type        = string
}

# Service configuration
variable "firestore_collection" {
  description = "Firestore collection name for police alerts"
  type        = string
  default     = "police_alerts"
}

variable "gcs_archive_bucket" {
  description = "GCS bucket name for archiving data"
  type        = string
}

variable "cors_allowed_origin" {
  description = "CORS allowed origin for alerts-service"
  type        = string
  default     = "https://wazepolicescrapergcp.web.app"
}

variable "rate_limit_per_minute" {
  description = "Rate limit for alerts-service API"
  type        = number
  default     = 30
}

variable "terraform_state_bucket" {
  description = "GCS bucket name for Terraform state"
  type        = string
}

# BigQuery configuration
variable "bigquery_dataset_id" {
  description = "BigQuery dataset ID for police alerts"
  type        = string
  default     = "policeAlertDataset"
}

variable "bigquery_location" {
  description = "BigQuery dataset location"
  type        = string
  default     = "US"
}

variable "bigquery_data_owners" {
  description = "List of user emails to grant OWNER access to the BigQuery dataset"
  type        = list(string)
  default     = []
}
