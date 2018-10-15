# docs.sourcegraph.com

``` shell
gsutil mb gs://docs.sourcegraph.com
gsutil iam ch allUsers:objectViewer gs://docs.sourcegraph.com
gsutil web set -m index.html -e 404.html gs://docs.sourcegraph.com
```