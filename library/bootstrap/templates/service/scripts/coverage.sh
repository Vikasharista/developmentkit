#!/usr/bin/env bash
# Runs unit + integration tests with coverage across internal/ and fails below the
# threshold. Excluded on purpose: gqlgen output (graphql/generated) and the
# integration-test helpers (internal/testutil). The root main package is not in
# -coverpkg: it only parses config and hands over to core-lib.
#
# COVERAGE_INCLUDE (an extended regex on file paths) narrows the gate, e.g. to the
# files a bootstrap run added to an existing repo whose older code is not at 100%:
#   COVERAGE_INCLUDE='/internal/app/catalog/.*widget' make cover
set -euo pipefail

THRESHOLD="${COVERAGE_THRESHOLD:-100}"
: "${TEST_DATABASE_DSN:?TEST_DATABASE_DSN must point at a migratable test database}"

module=$(go list -m)
go test -tags integration -count=1 -covermode=atomic -coverpkg=./internal/... -coverprofile=coverage.out ./...

{ head -1 coverage.out; tail -n +2 coverage.out | grep -vE "^${module}/internal/(app/[^/]+/graphql/generated|testutil)/" || true; } > coverage.filtered.out

if [ -n "${COVERAGE_INCLUDE:-}" ]; then
  { head -1 coverage.filtered.out; tail -n +2 coverage.filtered.out | grep -E "${COVERAGE_INCLUDE}" || true; } > coverage.include.out
  mv coverage.include.out coverage.filtered.out
fi

total=$(go tool cover -func=coverage.filtered.out | awk '/^total:/ {sub(/%/, "", $3); print $3}')
echo "coverage: ${total}% (threshold ${THRESHOLD}%)"
if awk -v t="$total" -v min="$THRESHOLD" 'BEGIN { exit !(t + 0 < min + 0) }'; then
  echo "uncovered functions:"
  go tool cover -func=coverage.filtered.out | awk '$NF != "100.0%" && !/^total:/'
  exit 1
fi
