# Service Account resource
resource "google_service_account" "service_account" {
  count = var.create_service_account ? 1 : 0

  account_id   = var.account_id
  display_name = var.display_name
  description  = var.description
  project      = var.project_id
}

# Data source for existing service account
data "google_service_account" "existing" {
  count = var.create_service_account ? 0 : 1

  account_id = var.account_id
  project    = var.project_id
}

# Local to get the email regardless of whether we created it or used existing
locals {
  service_account_email = var.create_service_account ? google_service_account.service_account[0].email : data.google_service_account.existing[0].email
}

# Project-level IAM bindings for this service account
resource "google_project_iam_member" "project_roles" {
  for_each = toset(var.project_roles)

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${local.service_account_email}"
}
