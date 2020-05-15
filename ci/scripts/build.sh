#!/bin/bash -eux

pushd dp-image-importer
  make build
  cp build/dp-image-importer Dockerfile.concourse ../build
popd
