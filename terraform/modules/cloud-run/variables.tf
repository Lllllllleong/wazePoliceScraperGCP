# Cloud Run Service Module - Input Variables

variable "service_name" {
  description = "Name of the Cloud Run service"
  type        = string
}

variable "location" {
  description = "GCP region for the service"
  type        = string
  default     = "us-central1"
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "container_image" {
  description = "Full container image URL with tag"
  type        = string
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
  default     = 8080
}

variable "cpu_limit" {
  description = "CPU limit for the container"
  type        = string
  default     = "1"
}

variable "memory_limit" {
  description = "Memory limit for the container"
  type        = string
  default     = "512Mi"
}

variable "max_instance_count" {
  description = "Maximum number of instances"
  type        = number
  default     = 1
}

variable "min_instance_count" {
  description = "Minimum number of instances (0 for scale to zero)"
  type        = number
  default     = 0
}

variable "max_instance_request_concurrency" {
  description = "Maximum number of requests per instance"
  type        = number
  default     = 80
}

variable "timeout" {
  description = "Request timeout in seconds"
  type        = string
  default     = "300s"
}

variable "service_account_email" {
  description = "Service account email for the Cloud Run service"
  type        = string
}

variable "env_vars" {
  description = "Environment variables for the container"
  type        = map(string)
  default     = {}
}

variable "labels" {
  description = "Labels to apply to the service"
  type        = map(string)
  default     = {}
}

variable "ingress" {
  description = "Ingress settings for the service"
  type        = string
  default     = "INGRESS_TRAFFIC_ALL"

  validation {
    condition     = contains(["INGRESS_TRAFFIC_ALL", "INGRESS_TRAFFIC_INTERNAL_ONLY", "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"], var.ingress)
    error_message = "Ingress must be one of: INGRESS_TRAFFIC_ALL, INGRESS_TRAFFIC_INTERNAL_ONLY, INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER."
  }
}

variable "allow_unauthenticated" {
  description = "Allow unauthenticated access to the service"
  type        = bool
  default     = false
}

variable "client" {
  description = "Client that created the service (for tracking)"
  type        = string
  default     = "terraform"
}

variable "client_version" {
  description = "Version of the client that created the service"
  type        = string
  default     = ""
}

variable "launch_stage" {
  description = "Launch stage of the service (ALPHA, BETA, GA)"
  type        = string
  default     = "GA"
}

variable "startup_cpu_boost" {
  description = "Enable startup CPU boost"
  type        = bool
  default     = true
}
