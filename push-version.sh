#!/bin/bash

if [ -z "$1" ]; then
  echo "Please provide a version"
  exit 0
else
  echo $1
  sh tag-version.sh $1
  sh docker-update.sh
fi
