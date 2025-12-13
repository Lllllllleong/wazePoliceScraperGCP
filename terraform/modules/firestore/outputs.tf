output "database_name" {
  description = "Name of the Firestore database"
  value       = google_firestore_database.database.name
}

output "database_id" {
  description = "Full ID of the Firestore database"
  value       = google_firestore_database.database.id
}

output "database_uid" {
  description = "Unique ID of the Firestore database"
  value       = google_firestore_database.database.uid
}

output "database_location" {
  description = "Location of the Firestore database"
  value       = google_firestore_database.database.location_id
}
