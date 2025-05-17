#!/bin/sh
set -e

# Цикл по всем файлам с описанием ролей.
for file in $(find ./roles -type f); do
     # Извлечение переменных окружения, описывающих роль.
     envs=$(cat ${file} | grep -vE "^(#|[[:space:]]|$)")

     # Проверка, что скрипт для создания роли задан.
     if echo "${envs}" | grep -q "PGROLESCRIPT"; then
          # Извлечение логина, пароля и скрипта для создания роли.
          name=$(echo "${envs}" | grep "PGUSER" | cut -d '=' -f 2)
          password=$(echo "${envs}" | grep "PGPASSWORD" | cut -d '=' -f 2)
          script=$(echo "${envs}" | grep "PGROLESCRIPT" | cut -d '=' -f 2)

          # Создание/пересоздание роли с заданным логином и паролем.
          psql \
               --dbname=${PGDATABASE} \
               --host=${PGHOST} \
               --port=${PGPORT} \
               --username=${PGUSER} \
               -v role_name="${name}" \
               -v role_password="'${password}'" \
               --file=./roles_scripts/${script}
     fi
done