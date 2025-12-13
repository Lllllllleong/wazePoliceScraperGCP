# Terraform and provider version constraints
# Backend configuration included here (only one terraform block allowed)

terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }

  # Remote state backend
  backend "gcs" {
    bucket = "wazepolicescrapergcp-terraform-state"
    prefix = "terraform/prod/state"
  }
}
