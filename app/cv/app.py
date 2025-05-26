#!/usr/bin/env python3
from uuid import uuid4
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
    'host': os.getenv('POSTGRES_HOST', 'localhost'),
    'database': os.getenv('POSTGRES_DB', 'image_db'),
    'user': os.getenv('POSTGRES_USER', 'postgres'),
    'password': os.getenv('POSTGRES_PASSWORD', 'postgres'),
    'port': os.getenv('POSTGRES_PORT', '5432')
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
    try:
        logger.info(f"sending photo {filename} is_nsfw: {is_nsfw}")
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
                logger.info(f"file {file_path} doesnt exist")
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
            
            logger.info(result.tags)
            
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

@app.route('/classify', methods=['POST'])
def classify_image():
    """
    Submit an image for classification
    ---
    consumes:
      - application/json
    parameters:
      - in: body
        name: body
        description: JSON containing filename to process
        required: true
        schema:
          type: object
          properties:
            filename:
              type: string
              description: Name of the file in the input folder
    responses:
      202:
        description: Task accepted for processing
      400:
        description: Invalid request
    """
    if not request.is_json:
        logger.warning("Received non-JSON request")
        return jsonify({'error': 'Request must be JSON'}), 400
        
    data = request.get_json()
    if not data or 'filename' not in data:
        logger.warning("Missing filename in request")
        return jsonify({'error': 'Missing filename in request'}), 400
    
    filename = data['filename']
    if not isinstance(filename, str) or not filename.strip():
        logger.warning(f"Invalid filename: {filename}")
        return jsonify({'error': 'Invalid filename'}), 400
    
    task_id = str(uuid4())
    
    task_queue.put((task_id, filename))
    logger.info(f"Queued task {task_id} for file {filename}")
    
    return jsonify({
        'task_id': task_id,
        'status': 'queued',
        'filename': filename,
        'message': 'Image added to processing queue'
    }), 202

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
    NUM_WORKERS = 2
    for i in range(NUM_WORKERS):
        t = threading.Thread(target=worker, daemon=True)
        t.start()
        logger.info(f"Started worker thread {i+1}")
        
    app.run(host='0.0.0.0', port=8050, threaded=True)