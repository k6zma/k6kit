#!/usr/bin/env bash

set -euo pipefail

REPORT_FILE=".bench-precommit.txt"
trap 'rm -f "$REPORT_FILE"' EXIT

BENCH_NO_FIELDS_MAX_NS="${BENCH_NO_FIELDS_MAX_NS:-350.0}"
BENCH_WITH_FIELDS_MAX_NS="${BENCH_WITH_FIELDS_MAX_NS:-1100.0}"
BENCH_WITH_CONTEXT_MAX_NS="${BENCH_WITH_CONTEXT_MAX_NS:-950.0}"
BENCH_WITH_OTEL_MAX_NS="${BENCH_WITH_OTEL_MAX_NS:-750.0}"
BENCH_DEBUG_DISABLED_MAX_NS="${BENCH_DEBUG_DISABLED_MAX_NS:-25.0}"

go test -run '^$' -bench '^(BenchmarkNoFields|BenchmarkFiveFields|BenchmarkWithContext|BenchmarkWithOtelTrace|BenchmarkDebugDisabled)$' -benchmem ./log > "$REPORT_FILE"

cat "$REPORT_FILE"

extract_ns() {
  local bench_name="$1"
  awk -v b="$bench_name" '
    $1 ~ ("^" b "(/|-)") {
      for (i = 1; i <= NF; i++) {
        if ($i == "ns/op") {
          print $(i - 1)
          exit
        }
      }
    }
  ' "$REPORT_FILE"
}

check_threshold() {
  local bench_name="$1"
  local limit="$2"
  local value

  value="$(extract_ns "$bench_name")"
  if [ -z "$value" ]; then
    echo "Missing benchmark output for $bench_name"
    return 1
  fi

  if awk -v v="$value" -v l="$limit" 'BEGIN { exit !(v <= l) }'; then
    echo "Benchmark within threshold: $bench_name=${value}ns/op <= ${limit}ns/op"
    return 0
  fi

  echo "Benchmark threshold exceeded: $bench_name=${value}ns/op > ${limit}ns/op"
  return 1
}

failed=0

check_threshold "BenchmarkNoFields" "$BENCH_NO_FIELDS_MAX_NS" || failed=1
check_threshold "BenchmarkFiveFields" "$BENCH_WITH_FIELDS_MAX_NS" || failed=1
check_threshold "BenchmarkWithContext" "$BENCH_WITH_CONTEXT_MAX_NS" || failed=1
check_threshold "BenchmarkWithOtelTrace" "$BENCH_WITH_OTEL_MAX_NS" || failed=1
check_threshold "BenchmarkDebugDisabled" "$BENCH_DEBUG_DISABLED_MAX_NS" || failed=1

if [ "$failed" -ne 0 ]; then
  exit 1
fi
