FROM python:3.12-bookworm

WORKDIR /app

RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    libsm6 \
    libxext6 \
    libgl1-mesa-glx \
    && rm -rf /var/lib/apt/lists/*

COPY app/cv/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/cv ./app/cv

RUN mkdir -p static/img
COPY static/img ./static/img

CMD ["python3", "app/cv/app.py"]