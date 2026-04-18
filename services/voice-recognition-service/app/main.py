import logging
import os
from concurrent import futures

import grpc
import grpc_reflection.v1alpha.reflection as reflection

from app.generated import voice_recognition_pb2, voice_recognition_pb2_grpc
from app.service.audio_recognizer import AudioRecognizerService, warmup_models

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def serve() -> None:
    port = os.getenv("VOICE_RECOGNITION_GRPC_PORT", "9010")
    warmup_models()
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

    bind_address = f"[::]:{port}"
    bound_port = server.add_insecure_port(bind_address)
    if bound_port == 0:
        raise RuntimeError(f"Failed to bind voice-recognition-service to {bind_address}")

    logger.info(
        "Starting voice-recognition-service on port %s (pid=%s, backend process active)",
        bound_port,
        os.getpid(),
    )
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
