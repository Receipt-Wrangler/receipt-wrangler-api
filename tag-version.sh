#!/bin/bash

if [ -z "$1" ]; then
  echo "Please provide a version"
  exit 0
else
  git tag $1
  git push origin $1
fi
