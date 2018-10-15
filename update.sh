#!/bin/bash

set -eu -o pipefail

docsite -assets assets -sources ../sourcegraph/doc -templates templates gen -out out
cd out && gsutil -m cp -a public-read -z html,css,js -r . gs://docs.sourcegraph.com
