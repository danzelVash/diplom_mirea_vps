FROM python:3.11-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    ffmpeg \
    libsndfile1 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
ENV PYTHONPATH=/app:/app/app/generated

COPY services/voice-recognition-service/requirements.txt ./requirements.txt
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

COPY services/voice-recognition-service /app
RUN python -m grpc_tools.protoc \
    -I./api \
    --python_out=./app/generated \
    --grpc_python_out=./app/generated \
    ./api/voice_recognition.proto

CMD ["python", "-m", "app.main"]
