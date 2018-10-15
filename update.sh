#!/bin/bash

set -eu -o pipefail

docsite -assets assets -sources ../sourcegraph/doc -templates templates gen -out out
gsutil cp -a public-read -j html,css,js -z html,css,js -r out gs://docs.sourcegraph.com
