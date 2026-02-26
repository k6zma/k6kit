#!/usr/bin/env bash

set -euo pipefail

REPORT_FILE=".bench-precommit.txt"
trap 'rm -f "$REPORT_FILE"' EXIT

BENCH_NO_FIELDS_JSON_MAX_NS="${BENCH_NO_FIELDS_JSON_MAX_NS:-380.0}"
BENCH_NO_FIELDS_TEXT_MAX_NS="${BENCH_NO_FIELDS_TEXT_MAX_NS:-220.0}"
BENCH_WITH_FIELDS_JSON_MAX_NS="${BENCH_WITH_FIELDS_JSON_MAX_NS:-900.0}"
BENCH_WITH_FIELDS_TEXT_MAX_NS="${BENCH_WITH_FIELDS_TEXT_MAX_NS:-650.0}"
BENCH_WITH_CONTEXT_JSON_MAX_NS="${BENCH_WITH_CONTEXT_JSON_MAX_NS:-850.0}"
BENCH_WITH_CONTEXT_TEXT_MAX_NS="${BENCH_WITH_CONTEXT_TEXT_MAX_NS:-450.0}"
BENCH_WITH_OTEL_JSON_MAX_NS="${BENCH_WITH_OTEL_JSON_MAX_NS:-700.0}"
BENCH_WITH_OTEL_TEXT_MAX_NS="${BENCH_WITH_OTEL_TEXT_MAX_NS:-380.0}"
BENCH_DEBUG_DISABLED_JSON_MAX_NS="${BENCH_DEBUG_DISABLED_JSON_MAX_NS:-30.0}"
BENCH_DEBUG_DISABLED_TEXT_MAX_NS="${BENCH_DEBUG_DISABLED_TEXT_MAX_NS:-30.0}"

go test -run '^$' -bench '^(BenchmarkNoFields|BenchmarkFiveFields|BenchmarkWithContext|BenchmarkWithOtelTrace|BenchmarkDebugDisabled)$' -benchmem ./log > "$REPORT_FILE"

cat "$REPORT_FILE"

extract_ns() {
	local bench_case="$1"
	awk -v b="$bench_case" '
	  $1 ~ ("^" b "-") {
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
	local bench_case="$1"
	local limit="$2"
	local value

	value="$(extract_ns "$bench_case")"
	if [ -z "$value" ]; then
		echo "Missing benchmark output for $bench_case"
		return 1
	fi

	if awk -v v="$value" -v l="$limit" 'BEGIN { exit !(v <= l) }'; then
		echo "Benchmark within threshold: $bench_case=${value}ns/op <= ${limit}ns/op"
		return 0
	fi

	echo "Benchmark threshold exceeded: $bench_case=${value}ns/op > ${limit}ns/op"
	return 1
}

failed=0

check_threshold "BenchmarkNoFields/json" "$BENCH_NO_FIELDS_JSON_MAX_NS" || failed=1
check_threshold "BenchmarkNoFields/text" "$BENCH_NO_FIELDS_TEXT_MAX_NS" || failed=1
check_threshold "BenchmarkFiveFields/json" "$BENCH_WITH_FIELDS_JSON_MAX_NS" || failed=1
check_threshold "BenchmarkFiveFields/text" "$BENCH_WITH_FIELDS_TEXT_MAX_NS" || failed=1
check_threshold "BenchmarkWithContext/json" "$BENCH_WITH_CONTEXT_JSON_MAX_NS" || failed=1
check_threshold "BenchmarkWithContext/text" "$BENCH_WITH_CONTEXT_TEXT_MAX_NS" || failed=1
check_threshold "BenchmarkWithOtelTrace/json" "$BENCH_WITH_OTEL_JSON_MAX_NS" || failed=1
check_threshold "BenchmarkWithOtelTrace/text" "$BENCH_WITH_OTEL_TEXT_MAX_NS" || failed=1
check_threshold "BenchmarkDebugDisabled/json" "$BENCH_DEBUG_DISABLED_JSON_MAX_NS" || failed=1
check_threshold "BenchmarkDebugDisabled/text" "$BENCH_DEBUG_DISABLED_TEXT_MAX_NS" || failed=1

if [ "$failed" -ne 0 ]; then
  exit 1
fi
