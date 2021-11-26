#!/bin/bash
make build
# export GIPHY_CONNECTOR_PUBLIC_KEY=LmcKidcjmjvjpFaWU1jmDyH8Nn1Yg8IWOgS9+9GuBck=
export GIPHY_CONNECTOR_PUBLIC_KEY=H/5tudlJkRjlrPGX3iZu8zWDK8o3OXdUVoQQX34Ux8Y=
./dist/giphy-connector --mysql.dsn "root@tcp(localhost)/giphy_connector?parseTime=true" 