# Output values for production environment

output "project_id" {
  description = "GCP Project ID"
  value       = var.project_id
}

output "project_number" {
  description = "GCP Project Number"
  value       = data.google_project.current.number
}

output "environment" {
  description = "Environment name"
  value       = var.environment
}

output "region" {
  description = "Primary GCP region"
  value       = var.region
}

# =============================================================================
# BIGQUERY OUTPUTS
# =============================================================================

output "bigquery_dataset_id" {
  description = "BigQuery dataset ID"
  value       = module.bigquery_dataset.dataset_id
}

output "bigquery_dataset_location" {
  description = "BigQuery dataset location"
  value       = module.bigquery_dataset.dataset_location
}

# =============================================================================
# STORAGE OUTPUTS
# =============================================================================

output "archive_bucket_name" {
  description = "Name of the archive storage bucket"
  value       = module.archive_bucket.bucket_name
}

output "archive_bucket_url" {
  description = "URL of the archive bucket"
  value       = module.archive_bucket.bucket_url
}

# =============================================================================
# CLOUD RUN SERVICE OUTPUTS
# =============================================================================

output "scraper_service_url" {
  description = "URL of the scraper service"
  value       = module.scraper_service.service_url
}

output "alerts_service_url" {
  description = "URL of the alerts service (public API)"
  value       = module.alerts_service.service_url
}

output "archive_service_url" {
  description = "URL of the archive service"
  value       = module.archive_service.service_url
}

output "all_service_urls" {
  description = "Map of all Cloud Run service URLs"
  value = {
    scraper = module.scraper_service.service_url
    alerts  = module.alerts_service.service_url
    archive = module.archive_service.service_url
  }
}

# =============================================================================
# CLOUD SCHEDULER OUTPUTS
# =============================================================================

output "scraper_scheduler_name" {
  description = "Name of the scraper scheduler job"
  value       = module.scraper_scheduler.job_name
}

output "archive_scheduler_name" {
  description = "Name of the archive scheduler job"
  value       = module.archive_scheduler.job_name
}

output "scheduler_jobs" {
  description = "Map of all Cloud Scheduler jobs with their schedules"
  value = {
    scraper = {
      name     = module.scraper_scheduler.job_name
      schedule = module.scraper_scheduler.schedule
      state    = module.scraper_scheduler.state
    }
    archive = {
      name     = module.archive_scheduler.job_name
      schedule = module.archive_scheduler.schedule
      state    = module.archive_scheduler.state
    }
  }
}

# =============================================================================
# ARTIFACT REGISTRY OUTPUTS
# =============================================================================

output "scraper_registry_url" {
  description = "Scraper service registry URL"
  value       = module.scraper_registry.repository_url
}

output "alerts_registry_url" {
  description = "Alerts service registry URL"
  value       = module.alerts_registry.repository_url
}

output "archive_registry_url" {
  description = "Archive service registry URL"
  value       = module.archive_registry.repository_url
}

output "all_registry_urls" {
  description = "Map of all Artifact Registry repository URLs"
  value = {
    scraper = module.scraper_registry.repository_url
    alerts  = module.alerts_registry.repository_url
    archive = module.archive_registry.repository_url
  }
}

# =============================================================================
# SERVICE ACCOUNT OUTPUTS
# =============================================================================

output "github_actions_sa_email" {
  description = "Email of the GitHub Actions service account"
  value       = module.github_actions_sa.service_account_email
}

output "firebase_adminsdk_sa_email" {
  description = "Email of the Firebase Admin SDK service account"
  value       = module.firebase_adminsdk_sa.service_account_email
}

# =============================================================================
# FIRESTORE OUTPUTS
# =============================================================================

output "firestore_database_name" {
  description = "Name of the Firestore database"
  value       = module.firestore_database.database_name
}

output "firestore_database_location" {
  description = "Location of the Firestore database"
  value       = module.firestore_database.database_location
}
