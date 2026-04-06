# Production environment main configuration
# This file will be expanded in Phase 2 to include module calls

provider "google" {
  project = var.project_id
  region  = var.region

  # Best practices
  user_project_override = true
  billing_project       = var.project_id

  # Default labels applied to all resources created by this provider
  default_labels = {
    managed-by  = "terraform"
    environment = var.environment
  }
}

# Common labels applied to all resources
locals {
  common_labels = {
    environment = var.environment
    managed-by  = "terraform"
    project     = "waze-police-scraper"
  }

  # Common environment variables for Cloud Run services
  common_env_vars = {
    GCP_PROJECT_ID       = var.project_id
    FIRESTORE_COLLECTION = var.firestore_collection
    GCS_BUCKET_NAME      = var.gcs_archive_bucket
  }
}

# Data source to verify project access
data "google_project" "current" {
  project_id = var.project_id
}

# =============================================================================
# BIGQUERY DATASET
# =============================================================================

module "bigquery_dataset" {
  source = "../../modules/bigquery"

  dataset_id            = var.bigquery_dataset_id
  location              = var.bigquery_location
  project_id            = var.project_id
  max_time_travel_hours = "168"

  description = "Dataset for police alert data from Waze"

  labels = local.common_labels

  # Grant your personal email OWNER access (matches exported config)
  owner_users = var.bigquery_data_owners
}

# =============================================================================
# STORAGE BUCKET
# =============================================================================

module "archive_bucket" {
  source = "../../modules/storage"

  bucket_name   = var.gcs_archive_bucket
  location      = "US"
  project_id    = var.project_id
  storage_class = "STANDARD"

  force_destroy               = false
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"

  # Soft delete with 7-day retention (604800 seconds)
  soft_delete_retention_seconds = 604800

  labels = local.common_labels
}

# =============================================================================
# CLOUD RUN SERVICES
# =============================================================================

# Alerts Service - API for frontend dashboard
module "alerts_service" {
  source = "../../modules/cloud-run"

  service_name          = "alerts-service"
  project_id            = var.project_id
  location              = var.region
  container_image       = var.alerts_image
  service_account_email = module.alerts_service_account.service_account_email

  cpu_limit    = "1"
  memory_limit = "512Mi"
  timeout      = "300s"

  max_instance_count               = 1
  max_instance_request_concurrency = 80

  env_vars = merge(
    local.common_env_vars,
    {
      CORS_ALLOWED_ORIGIN   = var.cors_allowed_origin
      RATE_LIMIT_PER_MINUTE = tostring(var.rate_limit_per_minute)
      FIREBASE_PROJECT_ID   = var.project_id
    }
  )

  labels = merge(
    local.common_labels,
    {
      service = "alerts"
    }
  )

  # Public API endpoint
  allow_unauthenticated = true
}

# =============================================================================
# ARTIFACT REGISTRY REPOSITORIES
# =============================================================================

# Alerts service repository
module "alerts_registry" {
  source = "../../modules/artifact-registry"

  project_id    = var.project_id
  location      = var.region
  repository_id = "alerts-service"
  description   = "Repository for alerts-service images"
  format        = "DOCKER"

  labels = merge(
    local.common_labels,
    {
      service = "alerts"
    }
  )

  cleanup_policy_dry_run    = false
  cleanup_older_than        = "2592000s"
  cleanup_keep_tag_prefixes = null
}

# =============================================================================
# SERVICE ACCOUNTS & IAM
# =============================================================================

# Alerts Service Account - Read from Firestore and GCS
module "alerts_service_account" {
  source = "../../modules/service-account"

  project_id             = var.project_id
  account_id             = "alerts-sa"
  display_name           = "Alerts Service"
  description            = "Service account for alerts-service with read access to Firestore and GCS"
  create_service_account = true

  project_roles = [
    "roles/datastore.viewer",    # Read police alerts from Firestore
    "roles/storage.objectViewer" # Read archive files from GCS bucket
  ]
}

# GitHub Actions runner service account
module "github_actions_sa" {
  source = "../../modules/service-account"

  project_id             = var.project_id
  account_id             = "github-actions-runner"
  display_name           = "GitHub Actions Runner"
  create_service_account = false # Already exists

  # Roles for CI/CD deployment (minimal permissions)
  project_roles = [
    "roles/artifactregistry.writer", # Push Docker images
    "roles/run.developer",           # Deploy Cloud Run services (less than admin)
    "roles/storage.objectAdmin"      # Access Terraform state bucket
  ]
}

# Grant GitHub Actions SA access to the Terraform state bucket
resource "google_storage_bucket_iam_member" "github_actions_state_access" {
  bucket = var.terraform_state_bucket
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${module.github_actions_sa.service_account_email}"
}

# Grant GitHub Actions SA ability to impersonate service accounts for deployments
resource "google_service_account_iam_member" "github_actions_impersonate_alerts" {
  service_account_id = module.alerts_service_account.service_account_name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.github_actions_sa.service_account_email}"
}

# Firebase Admin SDK service account
module "firebase_adminsdk_sa" {
  source = "../../modules/service-account"

  project_id             = var.project_id
  account_id             = "firebase-adminsdk-fbsvc"
  display_name           = "firebase-adminsdk"
  description            = "Firebase Admin SDK Service Agent"
  create_service_account = false # Already exists

  # Firebase-specific roles (removed firebaseauth.admin - not used)
  project_roles = [
    "roles/firebase.sdkAdminServiceAgent",
    "roles/iam.serviceAccountTokenCreator"
  ]
}

# =============================================================================
# FIRESTORE DATABASE
# =============================================================================

# Firestore database - Native mode
# NOTE: Terraform can manage the database resource but NOT collections/documents
# Collections are created dynamically by the application
module "firestore_database" {
  source = "../../modules/firestore"

  project_id                        = var.project_id
  database_name                     = "(default)"
  location_id                       = var.region
  database_type                     = "FIRESTORE_NATIVE"
  concurrency_mode                  = "PESSIMISTIC"
  app_engine_integration_mode       = "DISABLED"
  point_in_time_recovery_enablement = "POINT_IN_TIME_RECOVERY_DISABLED"
  delete_protection_state           = "DELETE_PROTECTION_DISABLED"
}

# =============================================================================
# AUDIT LOGGING CONFIGURATION
# =============================================================================

# Enable audit logging for Firestore data access
resource "google_project_iam_audit_config" "firestore_audit" {
  project = var.project_id
  service = "datastore.googleapis.com"

  audit_log_config {
    log_type = "DATA_READ"
  }

  audit_log_config {
    log_type = "DATA_WRITE"
  }
}

# Enable audit logging for Cloud Storage data access
resource "google_project_iam_audit_config" "storage_audit" {
  project = var.project_id
  service = "storage.googleapis.com"

  audit_log_config {
    log_type = "DATA_READ"
  }

  audit_log_config {
    log_type = "DATA_WRITE"
  }
}

# Enable audit logging for BigQuery data access
resource "google_project_iam_audit_config" "bigquery_audit" {
  project = var.project_id
  service = "bigquery.googleapis.com"

  audit_log_config {
    log_type = "DATA_READ"
  }

  audit_log_config {
    log_type = "DATA_WRITE"
  }
}
