#!/bin/bash
make build
export GIPHY_CONNECTOR_PUBLIC_KEY=LmcKidcjmjvjpFaWU1jmDyH8Nn1Yg8IWOgS9+9GuBck=
export GIPHY_API_KEY=3cONWofONEM18h3Gn0nNZrO3cnTaQcor
./dist/giphy-connector --mysql.dsn "root@tcp(localhost)/giphy_connector?parseTime=true" 