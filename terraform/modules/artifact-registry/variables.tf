variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "location" {
  description = "Repository location (e.g., us-central1)"
  type        = string
}

variable "repository_id" {
  description = "Repository ID (name)"
  type        = string
}

variable "description" {
  description = "Repository description"
  type        = string
  default     = ""
}

variable "format" {
  description = "Repository format (DOCKER, MAVEN, NPM, etc.)"
  type        = string
  default     = "DOCKER"
}

variable "mode" {
  description = "Repository mode (STANDARD_REPOSITORY, VIRTUAL_REPOSITORY, REMOTE_REPOSITORY)"
  type        = string
  default     = "STANDARD_REPOSITORY"
}

variable "labels" {
  description = "Labels to apply to the repository"
  type        = map(string)
  default     = {}
}

variable "cleanup_policy_dry_run" {
  description = "Whether to run cleanup policy in dry-run mode (test without deleting)"
  type        = bool
  default     = false
}

variable "cleanup_older_than" {
  description = "Delete images older than this duration (e.g., '2592000s' for 30 days)"
  type        = string
  default     = "2592000s" # 30 days
}

variable "cleanup_keep_tag_prefixes" {
  description = "Keep images with these tag prefixes (e.g., ['latest', 'prod']). Set to null to delete untagged images."
  type        = list(string)
  default     = null
}
