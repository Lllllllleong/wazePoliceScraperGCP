# BigQuery Dataset Module - Input Variables

variable "dataset_id" {
  description = "BigQuery dataset ID"
  type        = string
}

variable "location" {
  description = "Dataset location (US, EU, etc.)"
  type        = string
  default     = "US"
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "description" {
  description = "Description of the dataset"
  type        = string
  default     = ""
}

variable "default_table_expiration_ms" {
  description = "Default lifetime of all tables in milliseconds"
  type        = number
  default     = null
}

variable "max_time_travel_hours" {
  description = "Time travel window in hours (168 = 7 days)"
  type        = string
  default     = "168"
}

variable "delete_contents_on_destroy" {
  description = "Delete dataset contents when destroying"
  type        = bool
  default     = false
}

variable "labels" {
  description = "Labels to apply to the dataset"
  type        = map(string)
  default     = {}
}

variable "owner_users" {
  description = "List of user emails to grant OWNER access"
  type        = list(string)
  default     = []
}
