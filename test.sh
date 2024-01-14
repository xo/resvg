#!/bin/bash

(set -x;
  go test -ldflags '-extldflags "-static"' $@
)
