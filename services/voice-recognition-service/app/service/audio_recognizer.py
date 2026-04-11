import logging
import os
import tempfile
import time
from functools import lru_cache
from typing import List

import grpc

from app.asr.asr import asr
from app.generated import voice_recognition_pb2, voice_recognition_pb2_grpc
from app.select_scenario.compare_scenarios import ScenarioComparator

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

SIMILARITY_THRESHOLD = float(os.getenv("VOICE_RECOGNITION_SIMILARITY_THRESHOLD", "0.0"))


@lru_cache(maxsize=1)
def get_comparator() -> ScenarioComparator:
    return ScenarioComparator()


class AudioRecognizerService(voice_recognition_pb2_grpc.VoiceRecognitionServiceServicer):
    def _prepare_commands(self, commands) -> List[str]:
        return [cmd.name for cmd in commands if getattr(cmd, "name", "")]

    def GetAudio(self, request, context):
        logger.info("Received audio recognition request")
        start = time.time()

        try:
            commands = self._prepare_commands(request.commands)
            if not commands:
                return voice_recognition_pb2.GetAudioResponse(command="")

            with tempfile.NamedTemporaryFile(delete=False, suffix=".wav") as temp_file:
                temp_file.write(request.chunk)
                temp_file_path = temp_file.name

            try:
                transcription = asr(temp_file_path)
            finally:
                if os.path.exists(temp_file_path):
                    os.unlink(temp_file_path)

            if not transcription:
                logger.info("Empty transcription, returning no match")
                return voice_recognition_pb2.GetAudioResponse(command="")

            comparator = get_comparator()
            command, similarity = comparator.compare_scenarios(
                sentence=transcription,
                scenarios=commands,
                vectors_scenarios=comparator.get_scenario_embeddings(commands),
            )

            if similarity < SIMILARITY_THRESHOLD:
                logger.info(
                    "Similarity %.4f is below threshold %.4f, returning no match",
                    similarity,
                    SIMILARITY_THRESHOLD,
                )
                command = ""

            logger.info("Processed request in %.3fs", time.time() - start)
            return voice_recognition_pb2.GetAudioResponse(command=command)
        except Exception as exc:
            logger.exception("Audio recognition failed: %s", exc)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(exc))
            return voice_recognition_pb2.GetAudioResponse(command="")
