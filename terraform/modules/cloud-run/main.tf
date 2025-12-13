# Cloud Run Service Module - Main Configuration

resource "google_cloud_run_v2_service" "service" {
  name           = var.service_name
  location       = var.location
  project        = var.project_id
  ingress        = var.ingress
  client         = var.client
  client_version = var.client_version
  launch_stage   = var.launch_stage

  labels = var.labels

  template {
    service_account = var.service_account_email
    timeout         = var.timeout

    # Labels applied to template (in addition to service labels)
    labels = var.labels

    max_instance_request_concurrency = var.max_instance_request_concurrency

    scaling {
      min_instance_count = var.min_instance_count
      max_instance_count = var.max_instance_count
    }

    containers {
      image = var.container_image

      ports {
        name           = "http1"
        container_port = var.container_port
      }

      resources {
        cpu_idle = true

        limits = {
          cpu    = var.cpu_limit
          memory = var.memory_limit
        }

        startup_cpu_boost = var.startup_cpu_boost
      }

      # Startup probe for health checking
      startup_probe {
        initial_delay_seconds = 0
        timeout_seconds       = 240
        period_seconds        = 240
        failure_threshold     = 1

        tcp_socket {
          port = var.container_port
        }
      }

      # Dynamic environment variables
      dynamic "env" {
        for_each = var.env_vars
        content {
          name  = env.key
          value = env.value
        }
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

# IAM policy for public access (if enabled)
resource "google_cloud_run_v2_service_iam_member" "public_access" {
  count = var.allow_unauthenticated ? 1 : 0

  project  = google_cloud_run_v2_service.service.project
  location = google_cloud_run_v2_service.service.location
  name     = google_cloud_run_v2_service.service.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
