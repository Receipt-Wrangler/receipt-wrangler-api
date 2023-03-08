#!/bin/bash

docker build . --no-cache -t noah231515/receipt-wrangler:api
docker push noah231515/receipt-wrangler:api
