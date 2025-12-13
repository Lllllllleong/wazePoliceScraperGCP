variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "database_name" {
  description = "Firestore database name (use '(default)' for default database)"
  type        = string
  default     = "(default)"
}

variable "location_id" {
  description = "Location for the Firestore database"
  type        = string
  default     = "us-central1"
}

variable "database_type" {
  description = "Type of Firestore database (FIRESTORE_NATIVE or DATASTORE_MODE)"
  type        = string
  default     = "FIRESTORE_NATIVE"
}

variable "concurrency_mode" {
  description = "Concurrency mode (OPTIMISTIC or PESSIMISTIC)"
  type        = string
  default     = "PESSIMISTIC"
}

variable "app_engine_integration_mode" {
  description = "App Engine integration mode (ENABLED or DISABLED)"
  type        = string
  default     = "DISABLED"
}

variable "point_in_time_recovery_enablement" {
  description = "Enable point-in-time recovery (POINT_IN_TIME_RECOVERY_ENABLED or POINT_IN_TIME_RECOVERY_DISABLED)"
  type        = string
  default     = "POINT_IN_TIME_RECOVERY_DISABLED"
}

variable "delete_protection_state" {
  description = "Delete protection state (DELETE_PROTECTION_ENABLED or DELETE_PROTECTION_DISABLED)"
  type        = string
  default     = "DELETE_PROTECTION_DISABLED"
}
