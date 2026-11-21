#!/usr/bin/env bash
# Verifies the templates: renders an example spec into a scratch directory, generates,
# builds, then runs the coverage gate and the e2e suite against a docker Postgres.
#   library/bootstrap/scripts/golden.sh examples/widget-service.json /tmp/golden-widget
# Needs Go, Docker and access to your core library (service.core_module in the spec).
set -euo pipefail

here="$(cd "$(dirname "$0")/.." && pwd)"
spec="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
out="$2"
export GOPRIVATE="${GOPRIVATE:?set GOPRIVATE to your core library module path, e.g. bitbucket.org/your-org/core-lib}"

rm -rf "$out" && mkdir -p "$out"
(cd "$here/render" && go run . -spec "$spec" -out "$out" >/dev/null)
cd "$out"
git init -q
go mod tidy -e >/dev/null 2>&1
for config in $(ls internal/app/*/gqlgen.yml 2>/dev/null); do
  (cd "$(dirname "$config")" && go run github.com/99designs/gqlgen generate)
done
go mod tidy
go build ./...
go vet ./... && go vet -tags integration ./...
make db-up db-create-test
trap 'docker compose down -v >/dev/null 2>&1' EXIT
make cover
make e2e
echo "golden run passed: $out"
