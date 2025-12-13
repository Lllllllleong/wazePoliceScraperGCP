# Storage Bucket Module - Input Variables

variable "bucket_name" {
  description = "Name of the storage bucket (must be globally unique)"
  type        = string
}

variable "location" {
  description = "Bucket location"
  type        = string
  default     = "US"
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "storage_class" {
  description = "Storage class for the bucket"
  type        = string
  default     = "STANDARD"

  validation {
    condition     = contains(["STANDARD", "NEARLINE", "COLDLINE", "ARCHIVE"], var.storage_class)
    error_message = "Storage class must be one of: STANDARD, NEARLINE, COLDLINE, ARCHIVE."
  }
}

variable "versioning_enabled" {
  description = "Enable object versioning"
  type        = bool
  default     = false
}

variable "force_destroy" {
  description = "Allow bucket deletion even when not empty"
  type        = bool
  default     = false
}

variable "public_access_prevention" {
  description = "Prevent public access to bucket"
  type        = string
  default     = "inherited"

  validation {
    condition     = contains(["enforced", "inherited"], var.public_access_prevention)
    error_message = "Public access prevention must be 'enforced' or 'inherited'."
  }
}

variable "soft_delete_retention_seconds" {
  description = "Soft delete retention duration in seconds (default 7 days)"
  type        = number
  default     = 604800
}

variable "uniform_bucket_level_access" {
  description = "Enable uniform bucket-level access"
  type        = bool
  default     = true
}

variable "labels" {
  description = "Labels to apply to the bucket"
  type        = map(string)
  default     = {}
}
