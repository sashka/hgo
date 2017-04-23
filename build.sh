#!/usr/bin/env bash

# Adopted from https://github.com/hashicorp/terraform/blob/master/scripts/build.sh

# Get the git commit
GIT_COMMIT=$(git rev-parse --verify --short=10 HEAD)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_SEQ=$(git rev-list HEAD --count)

# Allow LD_FLAGS to be appended during development compilations
# LD_FLAGS="-X main.GitCommit=${GIT_SEQ}.${GIT_COMMIT}${GIT_DIRTY} $LD_FLAGS"

# In relase mode we don't want debug information in the binary
LD_FLAGS="-s -w"
# GC_FLAGS="-m -m"

go build -ldflags "${LD_FLAGS}" -gcflags "${GC_FLAGS}"
