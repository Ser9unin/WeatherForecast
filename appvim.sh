#!/bin/sh
cd /app
go mod download
go build -o /app/main /app/cmd/main
/app/main