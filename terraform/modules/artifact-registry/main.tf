resource "google_artifact_registry_repository" "repository" {
  location      = var.location
  repository_id = var.repository_id
  description   = var.description
  format        = var.format
  project       = var.project_id

  mode = var.mode

  labels = var.labels

  cleanup_policy_dry_run = var.cleanup_policy_dry_run

  # Cleanup policies for managing image retention
  dynamic "cleanup_policies" {
    for_each = var.cleanup_older_than != null ? [1] : []
    content {
      id     = "delete-old-images"
      action = "DELETE"

      dynamic "condition" {
        for_each = var.cleanup_keep_tag_prefixes != null ? [1] : []
        content {
          tag_state    = "TAGGED"
          tag_prefixes = var.cleanup_keep_tag_prefixes
          older_than   = var.cleanup_older_than
        }
      }

      dynamic "condition" {
        for_each = var.cleanup_keep_tag_prefixes == null ? [1] : []
        content {
          tag_state  = "UNTAGGED"
          older_than = var.cleanup_older_than
        }
      }
    }
  }
}
