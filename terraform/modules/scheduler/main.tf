resource "google_cloud_scheduler_job" "job" {
  name             = var.name
  description      = var.description
  schedule         = var.schedule
  time_zone        = var.time_zone
  attempt_deadline = var.attempt_deadline
  region           = var.region

  retry_config {
    retry_count          = var.retry_count
    max_retry_duration   = var.max_retry_duration
    min_backoff_duration = var.min_backoff_duration
    max_backoff_duration = var.max_backoff_duration
    max_doublings        = var.max_doublings
  }

  http_target {
    http_method = var.http_method
    uri         = var.target_uri

    # Only set headers if explicitly provided and not empty
    headers = length(var.headers) > 0 ? var.headers : null

    dynamic "oidc_token" {
      for_each = var.use_oidc_auth ? [1] : []
      content {
        service_account_email = var.service_account_email
        audience              = var.oidc_audience != null ? var.oidc_audience : var.target_uri
      }
    }
  }
}
