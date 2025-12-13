# Storage Bucket Module - Outputs

output "bucket_name" {
  description = "Name of the storage bucket"
  value       = google_storage_bucket.bucket.name
}

output "bucket_url" {
  description = "URL of the bucket"
  value       = google_storage_bucket.bucket.url
}

output "bucket_self_link" {
  description = "Self-link of the bucket"
  value       = google_storage_bucket.bucket.self_link
}

output "bucket_location" {
  description = "Location of the bucket"
  value       = google_storage_bucket.bucket.location
}
