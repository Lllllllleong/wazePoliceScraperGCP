# Storage Bucket Module - Main Configuration

resource "google_storage_bucket" "bucket" {
  name          = var.bucket_name
  location      = var.location
  project       = var.project_id
  storage_class = var.storage_class
  force_destroy = var.force_destroy

  labels = var.labels

  uniform_bucket_level_access = var.uniform_bucket_level_access
  public_access_prevention    = var.public_access_prevention

  # Soft delete policy (7-day retention by default)
  soft_delete_policy {
    retention_duration_seconds = var.soft_delete_retention_seconds
  }

  # Optional versioning
  dynamic "versioning" {
    for_each = var.versioning_enabled ? [1] : []
    content {
      enabled = true
    }
  }
}
