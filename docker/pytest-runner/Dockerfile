FROM python:3.11-slim

RUN pip install --no-cache-dir pytest pytest-timeout

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app
