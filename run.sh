#!/bin/bash
make build
# export GIPHY_CONNECTOR_PUBLIC_KEY=LmcKidcjmjvjpFaWU1jmDyH8Nn1Yg8IWOgS9+9GuBck=
export GIPHY_CONNECTOR_PUBLIC_KEY=HB1nauUIDcnfLDXhTscCo6vKWUb3tFXg0XPME491zyE=
./dist/giphy-connector --mysql.dsn "root@tcp(localhost)/giphy_connector?parseTime=true" 