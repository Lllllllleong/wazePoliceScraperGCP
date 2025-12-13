# Cloud Run Service Module - Outputs

output "service_name" {
  description = "Name of the Cloud Run service"
  value       = google_cloud_run_v2_service.service.name
}

output "service_id" {
  description = "Full resource ID of the service"
  value       = google_cloud_run_v2_service.service.id
}

output "service_url" {
  description = "URL of the Cloud Run service"
  value       = google_cloud_run_v2_service.service.uri
}

output "service_location" {
  description = "Location of the service"
  value       = google_cloud_run_v2_service.service.location
}

output "latest_revision" {
  description = "Name of the latest revision"
  value       = google_cloud_run_v2_service.service.latest_ready_revision
}
