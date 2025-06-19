#!/bin/bash
set -ex


LOG_DIR=/var/lib/docker/volumes/2025_1_superchips_db_data/_data/log
LOG_NAMES=(
    # Имена файлов в кавычках в столбик без запятых
)
LOGS=()
for log_name in "${LOG_NAMES[@]}"; do
    LOGS+=("${LOG_DIR}/${log_name}")
done


OUTPUT_DIR=/home/user/Desktop/reps/2025_1_SuperChips
OUTPUT_FILE_NAME=log_report.html

TEMP_FILE_NAME="temp_${OUTPUT_FILE_NAME}"



sudo pgbadger -O $OUTPUT_DIR -o $TEMP_FILE_NAME "${LOGS[@]}"
    
sudo chmod 777 $TEMP_FILE_NAME
cat $TEMP_FILE_NAME > $OUTPUT_FILE_NAME
rm $TEMP_FILE_NAME
