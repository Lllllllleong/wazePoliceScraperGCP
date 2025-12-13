output "service_account_email" {
  description = "Email address of the service account"
  value       = local.service_account_email
}

output "service_account_id" {
  description = "Fully qualified service account ID"
  value       = var.create_service_account ? google_service_account.service_account[0].id : data.google_service_account.existing[0].id
}

output "service_account_name" {
  description = "Fully qualified service account name"
  value       = var.create_service_account ? google_service_account.service_account[0].name : data.google_service_account.existing[0].name
}

output "service_account_unique_id" {
  description = "Unique ID of the service account"
  value       = var.create_service_account ? google_service_account.service_account[0].unique_id : data.google_service_account.existing[0].unique_id
}
