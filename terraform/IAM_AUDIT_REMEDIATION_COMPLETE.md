# IAM Audit Remediation Verification Report

**Date:** December 14, 2025  
**Reference Remediation:** [IAM_AUDIT_REMEDIATION.md](./IAM_AUDIT_REMEDIATION.md)  
**Status:** ✅ VERIFIED & COMPLETE  


## Executive Summary

This document verifies the security remediation for the Waze Police Scraper GCP project. All steps from `IAM_AUDIT_REMEDIATION.md` have been validated against the live environment using `gcloud`.

The project has transitioned from an overprivileged state to a **Least Privilege** model.

## 1. Service Account Segmentation Verification

**Verification Method:** `gcloud projects get-iam-policy`

| Service Account | Expected Role | Verified Status |
|-----------------|---------------|-----------------|
| **Default Compute SA** (`...-compute@...`) | **NO ROLES** (Previously Editor) | ✅ **VERIFIED** - No project-level roles found. |
| **Scraper SA** (`scraper-sa`) | `roles/datastore.user` | ✅ **VERIFIED** |
| **Alerts SA** (`alerts-sa`) | `roles/datastore.viewer`, `roles/storage.objectViewer` | ✅ **VERIFIED** |
| **Archive SA** (`archive-sa`) | `roles/datastore.user`, `roles/storage.objectAdmin` | ✅ **VERIFIED** |

**Finding:** The risk from the overprivileged default compute service account is resolved.

## 2. Cloud Run Security Verification

**Verification Method:** `gcloud run services describe` & `gcloud run services get-iam-policy`

### 2.1 Service Identity
| Service | Running As | Status |
|---------|------------|--------|
| `scraper-service` | `scraper-sa` | ✅ **VERIFIED** |
| `alerts-service` | `alerts-sa` | ✅ **VERIFIED** |
| `archive-service` | `archive-sa` | ✅ **VERIFIED** |

### 2.2 Access Control (Invoker)
| Service | Access Level | Policy Check | Status |
|---------|--------------|--------------|--------|
| `scraper-service` | **Private** | No `allUsers` in policy | ✅ **VERIFIED** |
| `archive-service` | **Private** | No `allUsers` in policy | ✅ **VERIFIED** |
| `alerts-service` | **Public** | `allUsers` has `roles/run.invoker` | ✅ **VERIFIED** |

## 3. Cloud Scheduler Security Verification

**Verification Method:** `gcloud scheduler jobs describe`

**Job:** `call-scraper`
*   **Authentication:** OIDC Token present.
*   **Identity:** Uses `scraper-sa@wazepolicescrapergcp.iam.gserviceaccount.com`.
*   **Status:** ✅ **VERIFIED** - Scheduler is securely authenticated.

## 4. Storage Security Verification

**Verification Method:** `gcloud storage buckets describe`

**Bucket:** `gs://wazepolicescrapergcp-archive`
*   **Public Access Prevention:** `enforced` ✅
*   **Uniform Bucket Level Access:** `true` ✅
*   **Soft Delete Retention:** `604800s` (7 days) ✅

## 5. CI/CD & Audit Logging Verification

### 5.1 GitHub Actions Runner
*   **Role Check:** `roles/run.admin` replaced with `roles/run.developer`. ✅
*   **Drift Check:** `roles/storage.admin` and other untracked roles are **ABSENT**. ✅

### 5.2 Audit Logging
**Verification Method:** `gcloud projects get-iam-policy` (auditConfigs section)
*   **BigQuery:** DATA_READ, DATA_WRITE enabled. ✅
*   **Firestore:** DATA_READ, DATA_WRITE enabled. ✅
*   **Storage:** DATA_READ, DATA_WRITE enabled. ✅

## 6. Conclusion

The infrastructure is now fully compliant with the intended security design.

*   **Zero Trust:** Services communicate via authenticated channels (OIDC).
*   **Least Privilege:** Each service has only the permissions it needs.
*   **Defense in Depth:** Public access prevention, audit logging, and strict IAM policies are in place.

**Remediation complete.**
