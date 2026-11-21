#!/usr/bin/env bash
# Concatenates each module's schema files and composes the local supergraph with rover.
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p schema
for dir in ../../internal/app/*/graphql/schema; do
  module=$(basename "$(dirname "$(dirname "$dir")")")
  cat "$dir"/*.graphqls > "schema/${module//_/-}.graphql"
done
APOLLO_ELV2_LICENSE=accept rover supergraph compose --config supergraph.yaml > supergraph.graphql
echo "wrote $(pwd)/supergraph.graphql"
