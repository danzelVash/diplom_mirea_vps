import logging
import os
import time
from collections import OrderedDict
from functools import lru_cache
from typing import List, Tuple, Union

import numpy as np
from sentence_transformers import SentenceTransformer
from sentence_transformers.util import cos_sim

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

DEFAULT_EMBEDDING_MODEL = os.getenv("VOICE_RECOGNITION_EMBEDDING_MODEL", "all-MiniLM-L6-v2")
SCENARIO_EMBEDDINGS_TTL_SECONDS = 300
SCENARIO_EMBEDDINGS_CACHE_SIZE = 256


@lru_cache(maxsize=1)
def load_embedding_model() -> SentenceTransformer:
    logger.info("Loading embedding model: %s", DEFAULT_EMBEDDING_MODEL)
    return SentenceTransformer(DEFAULT_EMBEDDING_MODEL)


class ScenarioComparator:
    def __init__(self):
        self.model = load_embedding_model()
        self._scenario_embeddings_cache: OrderedDict[Tuple[str, ...], Tuple[float, np.ndarray]] = (
            OrderedDict()
        )

    def get_scenario_embeddings(self, scenarios: Union[List[str], str]) -> np.ndarray:
        if isinstance(scenarios, str):
            scenarios = [scenarios]
        return self.model.encode(
            scenarios,
            convert_to_numpy=True,
            normalize_embeddings=True,
        )

    def get_cached_scenario_embeddings(self, scenarios_key: Tuple[str, ...]) -> np.ndarray:
        now = time.monotonic()
        cached_item = self._scenario_embeddings_cache.get(scenarios_key)
        if cached_item is not None:
            cached_at, embeddings = cached_item
            if now - cached_at < SCENARIO_EMBEDDINGS_TTL_SECONDS:
                logger.info(
                    "Scenario embeddings cache hit for %d commands (age=%.3fs)",
                    len(scenarios_key),
                    now - cached_at,
                )
                self._scenario_embeddings_cache.move_to_end(scenarios_key)
                return embeddings

            logger.info(
                "Scenario embeddings cache expired for %d commands (age=%.3fs)",
                len(scenarios_key),
                now - cached_at,
            )
            self._scenario_embeddings_cache.pop(scenarios_key, None)

        started_at = time.perf_counter()
        embeddings = self.get_scenario_embeddings(list(scenarios_key))
        logger.info(
            "Scenario embeddings cache miss for %d commands, built in %.3fs",
            len(scenarios_key),
            time.perf_counter() - started_at,
        )
        self._scenario_embeddings_cache[scenarios_key] = (now, embeddings)
        self._scenario_embeddings_cache.move_to_end(scenarios_key)

        while len(self._scenario_embeddings_cache) > SCENARIO_EMBEDDINGS_CACHE_SIZE:
            self._scenario_embeddings_cache.popitem(last=False)

        return embeddings

    def compare_scenarios(
        self,
        sentence: str,
        vectors_scenarios: np.ndarray,
        scenarios: List[str],
    ) -> Tuple[str, float]:
        started_at = time.perf_counter()
        logger.info("Matching transcription against scenarios: %s", sentence)
        embedding = self.model.encode(
            sentence,
            convert_to_numpy=True,
            normalize_embeddings=True,
        )

        similarities = cos_sim(embedding, vectors_scenarios)[0].cpu().numpy()
        best_match_idx = int(np.argmax(similarities))
        best_match_score = float(similarities[best_match_idx])
        logger.info(
            "Best matched command: %s (score=%.4f) in %.3fs",
            scenarios[best_match_idx],
            best_match_score,
            time.perf_counter() - started_at,
        )
        return scenarios[best_match_idx], best_match_score
