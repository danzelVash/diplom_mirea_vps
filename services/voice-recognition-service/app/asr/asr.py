import logging
import os
from pathlib import Path
from functools import lru_cache

import torch
import torchaudio
from transformers import WhisperForConditionalGeneration, WhisperProcessor

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

DEFAULT_ASR_MODEL = os.getenv("VOICE_RECOGNITION_ASR_MODEL", "artyomboyko/whisper-small-ru-v2")


@lru_cache(maxsize=1)
def load_model_components():
    logger.info("Loading ASR model: %s", DEFAULT_ASR_MODEL)
    processor = WhisperProcessor.from_pretrained(DEFAULT_ASR_MODEL)
    model = WhisperForConditionalGeneration.from_pretrained(DEFAULT_ASR_MODEL)
    model.config.forced_decoder_ids = None
    return processor, model


def asr(audio_file_path: str) -> str:
    logger.info("ASR processing file: %s", audio_file_path)

    if not os.path.exists(audio_file_path):
        raise FileNotFoundError(f"audio file not found: {audio_file_path}")

    file_ext = Path(audio_file_path).suffix.lower()
    if file_ext not in {".wav", ".mp3", ".ogg", ".m4a"}:
        raise ValueError(
            f"unsupported audio format: {file_ext}. Supported: .wav, .mp3, .ogg, .m4a"
        )

    waveform, sample_rate = torchaudio.load(audio_file_path)
    if waveform.shape[0] > 1:
        waveform = torch.mean(waveform, dim=0, keepdim=True)

    if sample_rate != 16000:
        resampler = torchaudio.transforms.Resample(sample_rate, 16000)
        waveform = resampler(waveform)

    processor, model = load_model_components()
    input_features = processor(
        waveform.squeeze().numpy(),
        sampling_rate=16000,
        return_tensors="pt",
    ).input_features

    predicted_ids = model.generate(input_features)
    transcription = processor.batch_decode(predicted_ids, skip_special_tokens=True)
    return transcription[0].strip()
