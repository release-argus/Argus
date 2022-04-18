#!/usr/bin/env bash
#
# (de)compress static web files

find web/ui/static -type f -exec gzip "$@" {} \;
