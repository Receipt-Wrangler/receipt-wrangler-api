#!/bin/bash

if [ -z "$1" ]; then
  echo "Please provide a tag"
  exit 1
fi

docker build . --no-cache -t noah231515/receipt-wrangler-api:$1
docker push noah231515/receipt-wrangler-api:$1
