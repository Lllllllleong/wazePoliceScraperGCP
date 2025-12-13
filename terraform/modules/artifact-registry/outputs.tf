output "repository_id" {
  description = "Repository ID"
  value       = google_artifact_registry_repository.repository.repository_id
}

output "repository_name" {
  description = "Full repository name"
  value       = google_artifact_registry_repository.repository.name
}

output "repository_location" {
  description = "Repository location"
  value       = google_artifact_registry_repository.repository.location
}

output "repository_url" {
  description = "Repository URL for docker push/pull"
  value       = "${google_artifact_registry_repository.repository.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.repository.repository_id}"
}
