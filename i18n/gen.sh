#!/bin/bash
set -e


#
# FIXME: find out why
#
#
# if put these go files to any paths under `project root path` and the `vendor` directory exists
# gotext will complains error:
#    gotext: extract failed: pipeline: golang.org/x/text/message is not imported
# if rename the `vendor` directory, everything get OK

# parse codes and extract all of text to be translated and
# merge to file: locales/zh/out.gotext.json
gotext -srclang=en update -out=catalog_gen.go -lang=en,zh

# overwrite the messages translation file
cp -avf locales/zh/out.gotext.json locales/zh/messages.gotext.json

## edit the messages.gotext.json to translate any new 
# vi locales/zh/messages.gotext.json

## rebuild to re-generate the catalog_gen.go
# gotext -srclang=en update -out=catalog_gen.go -lang=en,zh
