import logging
import os
from concurrent import futures

import grpc
import grpc_reflection.v1alpha.reflection as reflection

from app.generated import voice_recognition_pb2, voice_recognition_pb2_grpc
from app.service.audio_recognizer import AudioRecognizerService

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def serve() -> None:
    port = os.getenv("VOICE_RECOGNITION_GRPC_PORT", "9010")
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10),
        options=[
            ("grpc.max_send_message_length", 50 * 1024 * 1024),
            ("grpc.max_receive_message_length", 50 * 1024 * 1024),
        ],
    )

    voice_recognition_pb2_grpc.add_VoiceRecognitionServiceServicer_to_server(
        AudioRecognizerService(), server
    )

    service_name = voice_recognition_pb2.DESCRIPTOR.services_by_name[
        "VoiceRecognitionService"
    ].full_name
    reflection.enable_server_reflection((service_name, reflection.SERVICE_NAME), server)

    server.add_insecure_port(f"[::]:{port}")
    logger.info("Starting voice-recognition-service on port %s", port)
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
