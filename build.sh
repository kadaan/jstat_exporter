#!/usr/bin/env bash

BUILD_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
BINARY_DIR="$BUILD_DIR/.bin"
VERSION=$(cat $BUILD_DIR/.version)
REVISION="$(git rev-parse HEAD)"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
USER="${USER}"
HOST="$(hostname)"
BUILD_DATE="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

function verbose() { echo -e "$*"; }
function error() { echo -e "ERROR: $*" 1>&2; }
function fatal() { echo -e "ERROR: $*" 1>&2; exit 1; }
function pushd () { command pushd "$@" > /dev/null; }
function popd () { command popd > /dev/null; }

function trap_add() {
  localtrap_add_cmd=$1; shift || fatal "${FUNCNAME} usage error"
  for trap_add_name in "$@"; do
    trap -- "$(
      extract_trap_cmd() { printf '%s\n' "$3"; }
      eval "extract_trap_cmd $(trap -p "${trap_add_name}")"
      printf '%s\n' "${trap_add_cmd}"
    )" "${trap_add_name}" || fatal "unable to add to trap ${trap_add_name}"
  done
}
declare -f -t trap_add

function get_platform() {
  local unameOut="$(uname -s)"
  case "${unameOut}" in
    Linux*)
      echo "linux"
    ;;
    Darwin*)
      echo "darwin"
    ;;
    *)
      echo "Unsupported machine type :${unameOut}"
      exit 1
    ;;
  esac
}

PLATFORM=$(get_platform)
GOX="gox"

function download_gox() {
  if ! gox --version 2>1 /dev/null; then
    verbose "   --> $GOX"
    go install github.com/mitchellh/gox || fatal "go get 'github.com/mitchellh/gox' failed: $?"
  fi
}

function download_binaries() {
  download_gox || fatal "failed to download 'gox': $?"
  export PATH=$PATH:${BINARY_DIR}
}

function usage() {
  echo "Usage: build.sh [OPTIONS ...]"
  echo "Builds the binary for all supported platforms."
  echo ""
  echo "Options:"
  echo "    --help:        display this help"
  echo ""
}

function parse_args() {
  for var in "${@}"; do
    case "$var" in
      --help)
        usage
        exit 0
      ;;
    esac
  done
}

function run() {
  parse_args "$@"

  local revision=`git rev-parse HEAD`
  local branch=`git rev-parse --abbrev-ref HEAD`
  local host=`hostname`
  local buildDate=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
  local go_version="$(cat ${BUILD_DIR}/.go-version)"
  go version | grep -q "go version go${go_version%*.0} " || fatal "go version is not ${go_version%*.0}"

  if [[ -z "$TRAVIS" ]]; then
    verbose "Cleanup dist..."
    rm -rf dist/*
  fi

  verbose "Fetching binaries..."
  download_binaries

  XC_ARCH=${XC_ARCH:-"386 amd64 arm arm64"}
  XC_OS=${XC_OS:-"linux"}

  verbose "Building binaries..."
  ${GOX} -os="${XC_OS}" -arch="${XC_ARCH}" -tags 'osusergo netgo static_build' -ldflags "-d -s -w -extldflags \"-fno-PIC -static\" -X github.com/prometheus/common/version.Version=$VERSION -X github.com/prometheus/common/version.Revision=$REVISION -X github.com/prometheus/common/version.Branch=$BRANCH -X github.com/prometheus/common/version.BuildUser=$USER@$HOST -X github.com/prometheus/common/version.BuildDate=$BUILD_DATE" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" || fatal "gox failed: $?"

  if [[ -n "$TRAVIS" ]]; then
    verbose "Creating archives..."
    cd dist
    set -x
    for f in *; do
      local filename=$(basename "$f")
      local extension="${filename##*.}"
      local filename="${filename%.*}"
      if [[ "$filename" != "$extension" ]] && [[ -n "$extension" ]]; then
        extension=".$extension"
      else
        extension=""
      fi
      local archivename="$filename.tar.gz"
      verbose "   --> $archivename"
      local genericname="jstat_exporter$extension"
      mv -f "$f" "$genericname"
      tar -czf ${archivename} "$genericname"
      rm -rf "$genericname"
    done
  fi
}

run "$@"
