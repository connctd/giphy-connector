#!/bin/bash
make build
export GIPHY_CONNECTOR_PUBLIC_KEY=yourgiphyapikey
# ./dist/giphy-connector -migrate
./dist/giphy-connector