#!/bin/bash

MIGRATIONS_DIR="db/migrations"
DB_NAME="postgres_db"
DB_USER="admin"
DB_HOST="localhost"
DB_PORT="5432"
DB_PASSWORD=$(cat *.env | grep POSTGRES_PASSWORD | cut -f2 -d=)

export PGPASSWORD="$DB_PASSWORD"
export PGUSER="$DB_USER"

migrate -source file://"$(pwd)"/$MIGRATIONS_DIR -database postgres://$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable down
