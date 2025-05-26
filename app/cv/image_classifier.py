#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Сервис классификации изображений
Определяет теги, текст и проверяет на 18+ контент
"""

import os
import logging
import re
import time
from typing import List, Tuple
from dataclasses import dataclass
from pathlib import Path
from tqdm import tqdm

import warnings
warnings.filterwarnings("ignore")

import torch
from PIL import Image
import numpy as np

from transformers import BlipProcessor, BlipForConditionalGeneration, pipeline

# OCR
import pytesseract
from better_profanity import profanity

logging.basicConfig(
    level=logging.INFO, 
    format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# путь для кэша моделей
MODEL_CACHE_DIR = os.path.expanduser("~/.cache/huggingface/hub")
os.makedirs(MODEL_CACHE_DIR, exist_ok=True)

# черный список слов для автоматического пометки как NSFW
NSFW_BLACKLIST = {
    'duck', "утка"
}

@dataclass
class ClassificationResult:
    """Результат классификации изображения"""
    tags: List[str]
    is_adult: bool
    confidence_score: float
    detected_text: str
    profanity_detected: bool
    profanity_score: float
    processing_time: float
    timing_info: dict
    nsfw_reason: str = None


class ImageClassificationService:
    def __init__(self):
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        logger.info(f"Используется устройство: {self.device}")
        
        self.tesseract_config = r"--oem 3 --psm 6"
        
        profanity.load_censor_words()
        
        self._load_models()

    def _load_models(self):
        """Загрузка всех моделей с прогресс-баром"""
        logger.info("Загрузка моделей...")

        with tqdm(total=3, desc="Загрузка моделей") as pbar:
            self.blip_processor = BlipProcessor.from_pretrained(
                "Salesforce/blip-image-captioning-base",
                cache_dir=MODEL_CACHE_DIR
            )
            pbar.update(1)
            
            self.blip_model = BlipForConditionalGeneration.from_pretrained(
                "Salesforce/blip-image-captioning-base",
                cache_dir=MODEL_CACHE_DIR
            ).to(self.device)
            pbar.update(1)
            
            # NSFW детектор
            self.nsfw_classifier = pipeline(
                "image-classification", 
                model="Falconsai/nsfw_image_detection"
            )
            pbar.update(1)

        logger.info("Модели успешно загружены.")

    def _preprocess_image(self, image_path: str) -> Image.Image:
        """Предобработка изображения"""
        try:
            image = Image.open(image_path).convert("RGB")
            return image
        except Exception as e:
            logger.error(f"Ошибка открытия изображения {image_path}: {e}")
            raise

    def _generate_tags(self, image: Image.Image) -> Tuple[List[str], float]:
        """Генерация тегов из изображения с замером времени"""
        start_time = time.time()
        try:
            inputs = self.blip_processor(image, return_tensors="pt").to(self.device)
            with torch.no_grad():
                out = self.blip_model.generate(**inputs)
            caption = self.blip_processor.decode(out[0], skip_special_tokens=True)

            # Извлекаем слова из описания
            tags = re.findall(r"\b\w+\b", caption.lower())
            processing_time = time.time() - start_time
            return list(set(t for t in tags if len(t) > 2)), processing_time
        except Exception as e:
            logger.error(f"Ошибка генерации тегов: {e}")
            return [], time.time() - start_time

    def _detect_nsfw(self, image: Image.Image) -> Tuple[bool, float, float]:
        """Детекция NSFW контента с замером времени"""
        start_time = time.time()
        try:
            results = self.nsfw_classifier(image)
            nsfw_score = max(
                (r["score"] for r in results if "nsfw" in r["label"].lower()), 
                default=0
            )
            processing_time = time.time() - start_time
            return nsfw_score > 0.4, nsfw_score, processing_time
        except Exception as e:
            logger.error(f"Ошибка детекции NSFW: {e}")
            return False, 0.0, time.time() - start_time

    def _check_profanity(self, text: str, tags: List[str]) -> Tuple[bool, float, float]:
        """Проверка на мат с замером времени"""
        start_time = time.time()
        try:
            if not text and not tags:
                return False, 0.0, time.time() - start_time
                
            combined_text = f"{text} {' '.join(tags)}"
            has_profanity = profanity.contains_profanity(combined_text)
            processing_time = time.time() - start_time
            return has_profanity, 1.0 if has_profanity else 0.0, processing_time
        except Exception as e:
            logger.error(f"Ошибка проверки мата: {e}")
            return False, 0.0, time.time() - start_time

    def _check_blacklist(self, text: str, tags: List[str]) -> Tuple[bool, str]:
        """Проверка на наличие слов из черного списка"""
        if not text and not tags:
            return False, None
            
        combined = f"{text.lower()} {' '.join(tags).lower()}"
        for word in NSFW_BLACKLIST:
            if re.search(rf'\b{word}\b', combined):
                return True, f"Обнаружено запрещенное слово: {word}"
        return False, None

    def classify_image(self, image_path: str) -> ClassificationResult:
        """Классификация одного изображения с замерами времени"""
        total_start_time = time.time()
        logger.info(f"Обработка изображения: {image_path}")
        
        timing_info = {}
        try:
            # Предобработка
            start_time = time.time()
            image = self._preprocess_image(image_path)
            timing_info["preprocessing"] = time.time() - start_time

            # Генерация тегов
            tags, tags_time = self._generate_tags(image)
            timing_info["tag_generation"] = tags_time

            # Проверка черного списка
            blacklist_check_start = time.time()
            blacklist_detected, nsfw_reason = self._check_blacklist("", tags)
            timing_info["blacklist_check"] = time.time() - blacklist_check_start

            # NSFW детекция (пропускаем если уже обнаружено по черному списку)
            if not blacklist_detected:
                is_nsfw, n_score, nsfw_time = self._detect_nsfw(image)
                timing_info["nsfw_detection"] = nsfw_time
            else:
                is_nsfw, n_score = True, 1.0
                timing_info["nsfw_detection"] = 0.0  # Не выполняли проверку моделью

            # Проверка на мат
            has_profanity, profanity_score, profanity_time = self._check_profanity("", tags)
            timing_info["profanity_check"] = profanity_time

            # Общее время
            total_time = time.time() - total_start_time
            timing_info["total_processing"] = total_time

            # Логирование времени выполнения
            logger.info(f"Время обработки изображения {image_path}:")
            for stage, duration in timing_info.items():
                logger.info(f"- {stage.replace('_', ' ').title()}: {duration:.2f} сек")

            return ClassificationResult(
                tags=tags,
                is_adult=is_nsfw or blacklist_detected,
                confidence_score=n_score if not blacklist_detected else 1.0,
                detected_text="",
                profanity_detected=has_profanity,
                profanity_score=profanity_score,
                processing_time=total_time,
                timing_info=timing_info,
                nsfw_reason=nsfw_reason if blacklist_detected else "Модель классификации"
            )
        except Exception as e:
            logger.error(f"Критическая ошибка обработки {image_path}: {e}")
            return ClassificationResult(
                tags=[],
                is_adult=False,
                confidence_score=0.0,
                detected_text="",
                profanity_detected=False,
                profanity_score=0.0,
                processing_time=time.time() - total_start_time,
                timing_info={},
                nsfw_reason="Ошибка обработки"
            )
