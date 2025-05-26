#!/usr/bin/env python3
from image_classifier import *
from flask import Flask, request, jsonify
import os
import threading
from queue import Queue
from typing import Dict
import time
import psycopg2
from psycopg2 import sql
from psycopg2.extras import RealDictCursor

app = Flask(__name__)

service = ImageClassificationService()

DB_CONFIG = {
    'host': os.getenv('DB_HOST', 'localhost'),
    'database': os.getenv('DB_NAME', 'image_db'),
    'user': os.getenv('DB_USER', 'postgres'),
    'password': os.getenv('DB_PASSWORD', 'postgres'),
    'port': os.getenv('DB_PORT', '5432')
}

task_queue = Queue()
results: Dict[str, dict] = {}
lock = threading.Lock()

INPUT_FOLDER = os.getenv('INPUT_FOLDER', '/data/input')

def get_db_connection():
    """Create and return a new database connection"""
    return psycopg2.connect(**DB_CONFIG)

def update_image_status(filename: str, is_nsfw: bool, tags: list, nsfw_reason: str):
    """Update the image status in the database"""
    
    print(f"marking image as is_nsfw: {is_nsfw}")
    
    try:
        conn = get_db_connection()
        with conn.cursor() as cur:
            query = sql.SQL("""
                UPDATE flow
                SET 
                is_nsfw = %s
                WHERE media_url = %s
            """)
            cur.execute(query, (is_nsfw, filename))
            conn.commit()
    except Exception as e:
        app.logger.error(f"Database update failed for {filename}: {e}")
    finally:
        if conn:
            conn.close()

def worker():
    """Background worker to process images from the queue"""
    while True:
        task_id, filename = task_queue.get()
        try:
            file_path = os.path.join(INPUT_FOLDER, filename)
            
            if not os.path.exists(file_path):
                with lock:
                    results[task_id] = {
                        'status': 'error',
                        'error': f'File not found: {file_path}',
                        'timestamp': time.time()
                    }
                continue

            result = service.classify_image(file_path)
            
            update_image_status(
                filename=filename,
                is_nsfw=result.is_adult,
                tags=result.tags,
                nsfw_reason=result.nsfw_reason
            )
            
            with lock:
                results[task_id] = {
                    'status': 'completed',
                    'result': {
                        'filename': filename,
                        'tags': result.tags,
                        'is_adult': result.is_adult,
                        'confidence_score': result.confidence_score,
                        'profanity_detected': result.profanity_detected,
                        'profanity_score': result.profanity_score,
                        'nsfw_reason': result.nsfw_reason,
                        'processing_time': result.processing_time
                    },
                    'timestamp': time.time()
                }
        except Exception as e:
            with lock:
                results[task_id] = {
                    'status': 'error',
                    'error': str(e),
                    'timestamp': time.time()
                }
        finally:
            task_queue.task_done()

@app.route('/status/<task_id>', methods=['GET'])
def get_status(task_id):
    """Check status of a processing task"""
    with lock:
        result = results.get(task_id, None)
    
    if not result:
        return jsonify({'status': 'not_found'}), 404
    
    return jsonify(result)

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'queue_size': task_queue.qsize(),
        'input_folder': INPUT_FOLDER,
        'files_in_folder': os.listdir(INPUT_FOLDER) if os.path.exists(INPUT_FOLDER) else 'Folder not found'
    })
    

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8055, threaded=True)