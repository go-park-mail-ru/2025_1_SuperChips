#!/bin/bash

# вводится айди которому будут принадлежать все фотки
if [ -z "$1" ]; then
  echo "Usage: $0 <author_id>"
  exit 1
fi

AUTHOR_ID=$1
IMAGE_DIR="static/img"
DB_NAME="postgres"
DB_USER="admin"
DB_HOST="localhost"
DB_PORT="5432"
# вот таким хитрым образом достаем пароль :)
DB_PASSWORD=$(cat *.env | grep POSTGRES_PASSWORD | cut -f2 -d=)

export PGPASSWORD="$DB_PASSWORD"

if [ ! -d "$IMAGE_DIR" ]; then
  echo "Error: Directory $IMAGE_DIR not found!"
  exit 1
fi

for FILE in "$IMAGE_DIR"/*.{jpg,jpeg,png,gif}; do
  [ -e "$FILE" ] || continue

  HEADER=$(basename "$FILE" | cut -f1 -d.)
  
  URL="https://yourflow.ru/static/img/$(basename "$FILE")"
  
  QUERY="
    INSERT INTO flow (title, media_url, author_id)
    VALUES ('${HEADER//"'"/"''"}', '${URL//"'"/"''"}', $AUTHOR_ID);
  "

  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$QUERY" | echo

  echo "Processed: $HEADER"
done