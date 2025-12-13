# BigQuery Dataset Module - Main Configuration

resource "google_bigquery_dataset" "dataset" {
  dataset_id                  = var.dataset_id
  location                    = var.location
  project                     = var.project_id
  description                 = var.description
  default_table_expiration_ms = var.default_table_expiration_ms
  max_time_travel_hours       = var.max_time_travel_hours
  delete_contents_on_destroy  = var.delete_contents_on_destroy

  labels = var.labels

  # Default project-level access controls
  access {
    role          = "OWNER"
    special_group = "projectOwners"
  }

  access {
    role          = "READER"
    special_group = "projectReaders"
  }

  access {
    role          = "WRITER"
    special_group = "projectWriters"
  }

  # Individual user access (e.g., your personal email)
  dynamic "access" {
    for_each = var.owner_users
    content {
      role          = "OWNER"
      user_by_email = access.value
    }
  }
}
