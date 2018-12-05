#!/bin/sh
export GIN_MODE=release
export OS=`go env GOOS`
nohup ./bin/${OS}/main &
