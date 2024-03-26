#!/usr/bin/env sh
#
# (de)compress static web files

find web/ui/static -type f -exec gzip "$@" {} \;
