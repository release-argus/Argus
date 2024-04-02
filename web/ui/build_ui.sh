#!/bin/sh

set -e
current=$(pwd)

buildReactApp() {
  cd react-app
  echo "build react-app"
  npm ci
  npm run build
  cd "${current}"
  rm -rf ./static
  mv ./react-app/dist ./static
}

for i in "$@"; do
  case ${i} in
  --all)
    buildReactApp
    shift
    ;;
  esac
done

