#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-image-importer
  make audit
popd 