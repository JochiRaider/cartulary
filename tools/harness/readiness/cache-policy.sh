#!/usr/bin/env bash
# shellcheck shell=bash

cartulary_cache_non_cacheable_side_effects() {
  printf '%s\n' \
    "public_target_summary" \
    "failure_classification" \
    "drift_security_service_cleanup_verdicts"
}
