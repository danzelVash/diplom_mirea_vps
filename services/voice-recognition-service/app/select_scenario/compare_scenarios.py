import logging
import os
from functools import lru_cache
from typing import List, Tuple, Union

import numpy as np
import torch
from sentence_transformers import SentenceTransformer

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

DEFAULT_EMBEDDING_MODEL = os.getenv("VOICE_RECOGNITION_EMBEDDING_MODEL", "all-MiniLM-L6-v2")


@lru_cache(maxsize=1)
def load_embedding_model() -> SentenceTransformer:
    logger.info("Loading embedding model: %s", DEFAULT_EMBEDDING_MODEL)
    return SentenceTransformer(DEFAULT_EMBEDDING_MODEL)


class ScenarioComparator:
    def __init__(self):
        self.model = load_embedding_model()

    def get_scenario_embeddings(self, scenarios: Union[List[str], str]) -> np.ndarray:
        if isinstance(scenarios, str):
            scenarios = [scenarios]
        return self.model.encode(scenarios)

    def compare_scenarios(
        self,
        sentence: str,
        vectors_scenarios: np.ndarray,
        scenarios: List[str],
    ) -> Tuple[str, float]:
        logger.info("Matching transcription against scenarios: %s", sentence)
        embedding = torch.from_numpy(self.model.encode(sentence))
        scenario_vectors = torch.from_numpy(vectors_scenarios)

        similarities = self.model.similarity(embedding, scenario_vectors)[0]
        best_match_idx = int(np.argmax(similarities))
        best_match_score = float(similarities[best_match_idx])
        logger.info("Best matched command: %s (score=%.4f)", scenarios[best_match_idx], best_match_score)
        return scenarios[best_match_idx], best_match_score
