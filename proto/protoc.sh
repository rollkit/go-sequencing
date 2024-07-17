#!/usr/bin/env bash

set -eo pipefail

buf generate --path="./proto/sequencing" --template="buf.gen.yaml" --config="buf.yaml"