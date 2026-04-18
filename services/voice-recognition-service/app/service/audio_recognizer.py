import logging
import os
import time
from functools import lru_cache
from typing import List

import grpc

from app.asr.asr import asr_from_bytes, load_model_components
from app.generated import voice_recognition_pb2, voice_recognition_pb2_grpc
from app.select_scenario.compare_scenarios import ScenarioComparator, load_embedding_model

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
        start = time.perf_counter()

        try:
            commands = self._prepare_commands(request.commands)
            logger.info("Prepared %d commands for matching", len(commands))
            if not commands:
                return voice_recognition_pb2.GetAudioResponse(command="")

            asr_started_at = time.perf_counter()
            transcription = asr_from_bytes(request.chunk, suffix=".wav")
            logger.info("ASR stage completed in %.3fs", time.perf_counter() - asr_started_at)

            if not transcription:
                logger.info("Empty transcription, returning no match")
                return voice_recognition_pb2.GetAudioResponse(command="")

            comparator = get_comparator()
            scenario_key = tuple(commands)
            matching_started_at = time.perf_counter()
            command, similarity = comparator.compare_scenarios(
                sentence=transcription,
                scenarios=commands,
                vectors_scenarios=comparator.get_cached_scenario_embeddings(scenario_key),
            )
            logger.info(
                "Scenario matching stage completed in %.3fs",
                time.perf_counter() - matching_started_at,
            )

            if similarity < SIMILARITY_THRESHOLD:
                logger.info(
                    "Similarity %.4f is below threshold %.4f, returning no match",
                    similarity,
                    SIMILARITY_THRESHOLD,
                )
                command = ""

            logger.info(
                "Processed request in %.3fs, transcription=%r, selected_command=%r, similarity=%.4f",
                time.perf_counter() - start,
                transcription,
                command,
                similarity,
            )
            return voice_recognition_pb2.GetAudioResponse(command=command)
        except Exception as exc:
            logger.exception("Audio recognition failed: %s", exc)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(exc))
            return voice_recognition_pb2.GetAudioResponse(command="")


def warmup_models() -> None:
    logger.info("Warming up ASR and embedding models")
    load_model_components()
    load_embedding_model()
