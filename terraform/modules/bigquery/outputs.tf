# BigQuery Dataset Module - Outputs

output "dataset_id" {
  description = "BigQuery dataset ID"
  value       = google_bigquery_dataset.dataset.dataset_id
}

output "dataset_project" {
  description = "Project containing the dataset"
  value       = google_bigquery_dataset.dataset.project
}

output "dataset_location" {
  description = "Location of the dataset"
  value       = google_bigquery_dataset.dataset.location
}

output "dataset_self_link" {
  description = "Self-link of the dataset"
  value       = google_bigquery_dataset.dataset.self_link
}
