#!/usr/bin/env bash

set -euo pipefail

PROFILE_FILE=".coverage-precommit.out"
TEST_OUTPUT_FILE=".coverage-precommit.test.log"
trap 'rm -f "$PROFILE_FILE"' EXIT
MIN_COVERAGE="${MIN_COVERAGE:-70.0}"

packages="$(go list ./... | grep -v '/examples/')"
if [ -z "$packages" ]; then
  echo "No packages found for coverage check"
  exit 1
fi

if ! go test $packages -coverprofile "$PROFILE_FILE" >"$TEST_OUTPUT_FILE"; then
  cat "$TEST_OUTPUT_FILE"
  exit 1
fi

trap 'rm -f "$PROFILE_FILE" "$TEST_OUTPUT_FILE"' EXIT

coverage_line="$(go tool cover -func "$PROFILE_FILE" | awk '/^total:/ {print $3}')"
if [ -z "$coverage_line" ]; then
  echo "Unable to read total coverage from profile"
  exit 1
fi

coverage_value="${coverage_line%\%}"

echo "Total coverage: ${coverage_value}%"

if awk -v v="$coverage_value" -v l="$MIN_COVERAGE" 'BEGIN { exit !(v >= l) }'; then
  echo "Coverage threshold satisfied: ${coverage_value}% >= ${MIN_COVERAGE}%"
  exit 0
fi

echo "Coverage threshold failed: ${coverage_value}% < ${MIN_COVERAGE}%"
exit 1
