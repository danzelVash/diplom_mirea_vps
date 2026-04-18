import logging
import os
import time
from io import BytesIO
from functools import lru_cache
from pathlib import Path

import torch
import torchaudio
from transformers import WhisperForConditionalGeneration, WhisperProcessor

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

DEFAULT_ASR_MODEL = os.getenv("VOICE_RECOGNITION_ASR_MODEL", "artyomboyko/whisper-small-ru-v2")
DEFAULT_FASTER_WHISPER_MODEL = "Systran/faster-whisper-tiny"
# ASR_BACKEND = "transformers"
ASR_BACKEND = "faster_whisper"

ASR_BACKEND_TRANSFORMERS = "transformers"
ASR_BACKEND_FASTER_WHISPER = "faster_whisper"

DEFAULT_DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
DEFAULT_TORCH_DTYPE = torch.float16 if DEFAULT_DEVICE == "cuda" else torch.float32

TARGET_SAMPLE_RATE = 16000
SUPPORTED_AUDIO_EXTENSIONS = {".wav", ".mp3", ".ogg", ".m4a"}


@lru_cache(maxsize=1)
def load_transformers_components():
    logger.info("Loading ASR model: %s on %s", DEFAULT_ASR_MODEL, DEFAULT_DEVICE)
    processor = WhisperProcessor.from_pretrained(DEFAULT_ASR_MODEL)
    model = WhisperForConditionalGeneration.from_pretrained(
        DEFAULT_ASR_MODEL,
        low_cpu_mem_usage=False,
        device_map=None,
        torch_dtype=DEFAULT_TORCH_DTYPE,
    )
    model.config.forced_decoder_ids = None
    model.to(DEFAULT_DEVICE)
    model.eval()
    return processor, model


@lru_cache(maxsize=1)
def load_faster_whisper_model():
    try:
        from faster_whisper import WhisperModel
    except ImportError as exc:
        raise RuntimeError(
            "ASR_BACKEND='faster_whisper', but package 'faster-whisper' is not installed"
        ) from exc

    compute_type = "float32"
    if DEFAULT_DEVICE == "cuda":
        compute_type = "float16"

    logger.info(
        "Loading faster-whisper model: %s on %s with compute_type=%s",
        DEFAULT_FASTER_WHISPER_MODEL,
        DEFAULT_DEVICE,
        compute_type,
    )
    return WhisperModel(
        DEFAULT_FASTER_WHISPER_MODEL,
        device=DEFAULT_DEVICE,
        compute_type=compute_type,
    )


def load_model_components():
    if ASR_BACKEND == ASR_BACKEND_TRANSFORMERS:
        return load_transformers_components()
    if ASR_BACKEND == ASR_BACKEND_FASTER_WHISPER:
        return load_faster_whisper_model()
    raise ValueError(f"Unsupported ASR_BACKEND: {ASR_BACKEND}")


@lru_cache(maxsize=8)
def get_resampler(sample_rate: int):
    return torchaudio.transforms.Resample(sample_rate, TARGET_SAMPLE_RATE)


def _load_waveform(audio_source: str | BytesIO, suffix: str = ".wav") -> torch.Tensor:
    if isinstance(audio_source, str):
        if not os.path.exists(audio_source):
            raise FileNotFoundError(f"audio file not found: {audio_source}")

        file_ext = Path(audio_source).suffix.lower()
        if file_ext not in SUPPORTED_AUDIO_EXTENSIONS:
            raise ValueError(
                f"unsupported audio format: {file_ext}. Supported: .wav, .mp3, .ogg, .m4a"
            )

    else:
        file_ext = suffix.lower()
        if file_ext not in SUPPORTED_AUDIO_EXTENSIONS:
            raise ValueError(
                f"unsupported audio format: {file_ext}. Supported: .wav, .mp3, .ogg, .m4a"
            )

    waveform, sample_rate = torchaudio.load(audio_source)
    if waveform.shape[0] > 1:
        waveform = torch.mean(waveform, dim=0, keepdim=True)

    if sample_rate != TARGET_SAMPLE_RATE:
        waveform = get_resampler(sample_rate)(waveform)

    return waveform


def asr(audio_file_path: str) -> str:
    logger.info("ASR processing file: %s", audio_file_path)
    waveform = _load_waveform(audio_file_path)
    return transcribe_waveform(waveform)


def asr_from_bytes(audio_bytes: bytes, suffix: str = ".wav") -> str:
    logger.info(
        "ASR backend=%s processing %d bytes from request payload",
        ASR_BACKEND,
        len(audio_bytes),
    )
    waveform = _load_waveform(BytesIO(audio_bytes), suffix=suffix)
    return transcribe_waveform(waveform)


def transcribe_waveform(waveform: torch.Tensor) -> str:
    if ASR_BACKEND == ASR_BACKEND_TRANSFORMERS:
        return transcribe_waveform_transformers(waveform)
    if ASR_BACKEND == ASR_BACKEND_FASTER_WHISPER:
        return transcribe_waveform_faster_whisper(waveform)
    raise ValueError(f"Unsupported ASR_BACKEND: {ASR_BACKEND}")


def transcribe_waveform_transformers(waveform: torch.Tensor) -> str:
    started_at = time.perf_counter()
    processor, model = load_transformers_components()
    input_features = processor(
        waveform.squeeze().numpy(),
        sampling_rate=TARGET_SAMPLE_RATE,
        return_tensors="pt",
    ).input_features

    input_features = input_features.to(device=DEFAULT_DEVICE, dtype=DEFAULT_TORCH_DTYPE)

    with torch.inference_mode():
        predicted_ids = model.generate(
            input_features,
            max_new_tokens=96,
            num_beams=1,
            do_sample=False,
            use_cache=True,
        )
    transcription = processor.batch_decode(predicted_ids, skip_special_tokens=True)
    result = transcription[0].strip()
    logger.info(
        "ASR backend=%s transcription=%r in %.3fs",
        ASR_BACKEND_TRANSFORMERS,
        result,
        time.perf_counter() - started_at,
    )
    return result


def transcribe_waveform_faster_whisper(waveform: torch.Tensor) -> str:
    started_at = time.perf_counter()
    model = load_faster_whisper_model()
    audio = waveform.squeeze().numpy()
    segments, _ = model.transcribe(
        audio,
        language="ru",
        beam_size=1,
        best_of=1,
        condition_on_previous_text=False,
    )
    result = " ".join(segment.text.strip() for segment in segments).strip()
    logger.info(
        "ASR backend=%s transcription=%r in %.3fs",
        ASR_BACKEND_FASTER_WHISPER,
        result,
        time.perf_counter() - started_at,
    )
    return result
