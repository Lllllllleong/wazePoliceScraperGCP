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
  public_access_prevention    = "inherited"

  # Soft delete with 7-day retention (604800 seconds)
  soft_delete_retention_seconds = 604800

  labels = local.common_labels
}

# =============================================================================
# CLOUD RUN SERVICES
# =============================================================================

# Scraper Service - Fetches data from Waze API
module "scraper_service" {
  source = "../../modules/cloud-run"

  service_name          = "scraper-service"
  project_id            = var.project_id
  location              = var.region
  container_image       = var.scraper_image
  service_account_email = var.service_account_email

  # Resource limits
  cpu_limit    = "1"
  memory_limit = "512Mi"
  timeout      = "300s"

  # Scaling configuration
  min_instance_count               = 0
  max_instance_count               = 1
  max_instance_request_concurrency = 80

  # Metadata (matches exported config)
  client         = "gcloud"
  client_version = ""
  launch_stage   = "GA"

  env_vars = merge(
    local.common_env_vars,
    {
      # Scraper-specific variables can be added here
    }
  )

  labels = merge(
    local.common_labels,
    {
      service = "scraper"
    }
  )

  # Scraper is triggered by Cloud Scheduler, not public
  allow_unauthenticated = false
}

# Alerts Service - API for frontend dashboard
module "alerts_service" {
  source = "../../modules/cloud-run"

  service_name          = "alerts-service"
  project_id            = var.project_id
  location              = var.region
  container_image       = var.alerts_image
  service_account_email = var.service_account_email

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

# Archive Service - Moves old data to GCS
module "archive_service" {
  source = "../../modules/cloud-run"

  service_name          = "archive-service"
  project_id            = var.project_id
  location              = var.region
  container_image       = var.archive_image
  service_account_email = var.service_account_email

  cpu_limit    = "1"
  memory_limit = "512Mi"
  timeout      = "600s"

  max_instance_count               = 1
  max_instance_request_concurrency = 1

  env_vars = local.common_env_vars

  labels = merge(
    local.common_labels,
    {
      service = "archive"
    }
  )

  # Archive is triggered by Cloud Scheduler, not public
  allow_unauthenticated = false
}

# =============================================================================
# CLOUD SCHEDULER JOBS
# =============================================================================

# Scraper job - Triggers scraper service every minute
module "scraper_scheduler" {
  source = "../../modules/scheduler"

  name        = "call-scraper"
  description = "Triggers the scraper service every minute."
  schedule    = "* * * * *"
  time_zone   = "UTC"
  region      = var.region

  http_method = "GET"
  target_uri  = module.scraper_service.service_url

  # No OIDC auth - scraper service allows unauthenticated (triggered by scheduler)
  use_oidc_auth = false

  attempt_deadline = "180s"

  # Retry configuration
  max_retry_duration   = "0s"
  min_backoff_duration = "5s"
  max_backoff_duration = "3600s"
  max_doublings        = 5
}

# Archive job - Triggers archive service daily at 00:05 Australia/Canberra
module "archive_scheduler" {
  source = "../../modules/scheduler"

  name        = "archive-police-alerts"
  description = "Triggers the archive service daily to archive previous day's data."
  schedule    = "5 0 * * *"
  time_zone   = "Australia/Canberra"
  region      = var.region

  http_method = "POST"
  target_uri  = module.archive_service.service_url

  # OIDC auth required - archive service requires authentication
  use_oidc_auth         = true
  service_account_email = var.service_account_email
  oidc_audience         = module.archive_service.service_url

  attempt_deadline = "180s"

  # Retry configuration
  max_retry_duration   = "0s"
  min_backoff_duration = "5s"
  max_backoff_duration = "3600s"
  max_doublings        = 5
}

# =============================================================================
# ARTIFACT REGISTRY REPOSITORIES
# =============================================================================

# Scraper service repository
module "scraper_registry" {
  source = "../../modules/artifact-registry"

  project_id    = var.project_id
  location      = var.region
  repository_id = "scraper-service"
  description   = "Docker repository for scraper-service"
  format        = "DOCKER"

  labels = merge(
    local.common_labels,
    {
      service = "scraper"
    }
  )

  # Cleanup policy: delete untagged images older than 30 days
  cleanup_policy_dry_run    = false
  cleanup_older_than        = "2592000s" # 30 days
  cleanup_keep_tag_prefixes = null       # Delete untagged only
}

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

# Archive service repository
module "archive_registry" {
  source = "../../modules/artifact-registry"

  project_id    = var.project_id
  location      = var.region
  repository_id = "archive-service"
  description   = "Docker repository for archive-service"
  format        = "DOCKER"

  labels = merge(
    local.common_labels,
    {
      service = "archive"
    }
  )

  cleanup_policy_dry_run    = false
  cleanup_older_than        = "2592000s"
  cleanup_keep_tag_prefixes = null
}

# =============================================================================
# SERVICE ACCOUNTS & IAM
# =============================================================================

# Default compute service account (GCP-managed, special account)
# Used by all Cloud Run services
# Note: This is a GCP-managed account and cannot be imported as a service account resource
# We only manage its IAM bindings

# IAM binding for compute service account
resource "google_project_iam_member" "compute_sa_editor" {
  project = var.project_id
  role    = "roles/editor"
  member  = "serviceAccount:${var.service_account_email}"

  # NOTE: roles/editor is overly permissive
  # TODO: Replace with least-privilege roles:
  #   - roles/datastore.user (Firestore access)
  #   - roles/storage.objectAdmin (GCS access)
  #   - roles/cloudscheduler.jobRunner (for scheduler triggers)
}

# GitHub Actions runner service account
module "github_actions_sa" {
  source = "../../modules/service-account"

  project_id             = var.project_id
  account_id             = "github-actions-runner"
  display_name           = "GitHub Actions Runner"
  create_service_account = false # Already exists

  # Roles for CI/CD deployment
  project_roles = [
    "roles/artifactregistry.writer", # Push Docker images
    "roles/run.admin"                # Deploy Cloud Run services
  ]
}

# Firebase Admin SDK service account
module "firebase_adminsdk_sa" {
  source = "../../modules/service-account"

  project_id             = var.project_id
  account_id             = "firebase-adminsdk-fbsvc"
  display_name           = "firebase-adminsdk"
  description            = "Firebase Admin SDK Service Agent"
  create_service_account = false # Already exists

  # Firebase-specific roles
  project_roles = [
    "roles/firebase.sdkAdminServiceAgent",
    "roles/firebaseauth.admin",
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
