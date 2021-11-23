#!/bin/bash
make build
export GIPHY_CONNECTOR_PUBLIC_KEY=920Eg+7pHFo2OCLcRBzJ57Eeva/s015PbqsBN8ocx8E=
./dist/giphy-connector --mysql.dsn "root@tcp(localhost)/giphy_connector?parseTime=true" 