#!/bin/sh

set -e
current=$(pwd)

function buildReactApp() {
  cd react-app
  echo "build react-app"
  npm run build
  cd "${current}"
  rm -rf ./static
  mv ./react-app/build ./static
}

for i in "$@"; do
  case ${i} in
  --all)
    buildReactApp
    shift
    ;;
  esac
done

