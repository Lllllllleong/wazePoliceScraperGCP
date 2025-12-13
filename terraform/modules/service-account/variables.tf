variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "account_id" {
  description = "Service account ID (the part before @)"
  type        = string
}

variable "display_name" {
  description = "Display name for the service account"
  type        = string
  default     = ""
}

variable "description" {
  description = "Description of the service account"
  type        = string
  default     = ""
}

variable "create_service_account" {
  description = "Whether to create the service account or use existing one"
  type        = bool
  default     = true
}

variable "project_roles" {
  description = "List of project-level IAM roles to grant to this service account"
  type        = list(string)
  default     = []
}
