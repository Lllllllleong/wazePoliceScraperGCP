variable "name" {
  description = "Name of the Cloud Scheduler job"
  type        = string
}

variable "description" {
  description = "Description of the Cloud Scheduler job"
  type        = string
  default     = ""
}

variable "schedule" {
  description = "Cron schedule for the job (e.g., '* * * * *' for every minute)"
  type        = string
}

variable "time_zone" {
  description = "Time zone for the schedule (e.g., 'UTC', 'Australia/Canberra')"
  type        = string
  default     = "UTC"
}

variable "attempt_deadline" {
  description = "Maximum time allowed for a single job attempt"
  type        = string
  default     = "180s"
}

variable "region" {
  description = "GCP region for the scheduler job"
  type        = string
}

variable "http_method" {
  description = "HTTP method for the target (GET, POST, etc.)"
  type        = string
  default     = "GET"
}

variable "target_uri" {
  description = "Target URI to invoke"
  type        = string
}

variable "headers" {
  description = "HTTP headers to include in the request (GCP sets User-Agent automatically)"
  type        = map(string)
  default     = {}
}

variable "use_oidc_auth" {
  description = "Whether to use OIDC authentication for the target"
  type        = bool
  default     = false
}

variable "service_account_email" {
  description = "Service account email for OIDC authentication"
  type        = string
  default     = null
}

variable "oidc_audience" {
  description = "OIDC audience (defaults to target_uri if not specified)"
  type        = string
  default     = null
}

variable "retry_count" {
  description = "Number of retry attempts (derived from max_doublings)"
  type        = number
  default     = 0
}

variable "max_retry_duration" {
  description = "Maximum time for all retry attempts"
  type        = string
  default     = "0s"
}

variable "min_backoff_duration" {
  description = "Minimum time between retries"
  type        = string
  default     = "5s"
}

variable "max_backoff_duration" {
  description = "Maximum time between retries"
  type        = string
  default     = "3600s"
}

variable "max_doublings" {
  description = "Maximum number of times the retry delay is doubled"
  type        = number
  default     = 5
}
