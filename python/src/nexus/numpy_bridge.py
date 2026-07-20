"""NumPy bridge — zero-copy array sharing across languages."""

from typing import Any

import numpy as np

from .types import NdArray, Value

DTYPE_F32 = 0
DTYPE_F64 = 1
DTYPE_I32 = 2
DTYPE_I64 = 3


def array_to_value(arr: np.ndarray) -> Value:
    """Convert a NumPy array to a Nexus Value with NdArray payload."""
    if arr.dtype == np.float32:
        dtype = DTYPE_F32
    elif arr.dtype == np.float64:
        dtype = DTYPE_F64
    elif arr.dtype == np.int32:
        dtype = DTYPE_I32
    elif arr.dtype == np.int64:
        dtype = DTYPE_I64
    else:
        arr = arr.astype(np.float32)
        dtype = DTYPE_F32
    return Value(
        type="ndarray",
        ndarray=NdArray(
            dtype=dtype,
            shape=list(arr.shape),
            data=arr.tobytes(),
        ),
    )


def value_to_array(v: Value) -> np.ndarray:
    """Convert a Nexus NdArray Value back to a NumPy array."""
    if v.type != "ndarray" or v.ndarray is None:
        raise TypeError(f"Expected ndarray value, got {v.type}")
    nd = v.ndarray
    dtype_map = {DTYPE_F32: np.float32, DTYPE_F64: np.float64, DTYPE_I32: np.int32, DTYPE_I64: np.int64}
    np_dtype = dtype_map.get(nd.dtype, np.float32)
    arr = np.frombuffer(nd.data, dtype=np_dtype)
    if nd.shape:
        arr = arr.reshape(nd.shape)
    return arr


def analyze_array(arr: np.ndarray) -> dict[str, Any]:
    """Run analytics on a numpy array and return results as dict."""
    return {
        "shape": list(arr.shape),
        "dtype": str(arr.dtype),
        "min": float(arr.min()),
        "max": float(arr.max()),
        "mean": float(arr.mean()),
        "std": float(arr.std()),
        "sum": float(arr.sum()),
        "size": int(arr.size),
        "nbytes": int(arr.nbytes),
    }


def transform_simd(arr: np.ndarray, scalar: float = 2.0) -> np.ndarray:
    """SIMD-like transform (maps to Rust SIMD via Nexus bridge)."""
    return arr * scalar


def fft_analyze(arr: np.ndarray) -> dict[str, Any]:
    """Perform FFT analysis on a 1D array."""
    fft_vals = np.fft.rfft(arr)
    magnitudes = np.abs(fft_vals)
    freqs = np.fft.rfftfreq(len(arr))
    return {
        "dominant_freq": float(freqs[np.argmax(magnitudes[1:]) + 1]) if len(magnitudes) > 1 else 0.0,
        "max_magnitude": float(magnitudes.max()),
        "mean_magnitude": float(magnitudes.mean()),
        "energy": float(np.sum(magnitudes ** 2)),
    }


def stats_report(data: list[float]) -> dict[str, Any]:
    """Generate a comprehensive statistical report."""
    arr = np.array(data, dtype=np.float64)
    return {
        **analyze_array(arr),
        "median": float(np.median(arr)),
        "q25": float(np.percentile(arr, 25)),
        "q75": float(np.percentile(arr, 75)),
        "skewness": float(np.mean(((arr - arr.mean()) / arr.std()) ** 3)) if arr.std() > 0 else 0.0,
        "kurtosis": float(np.mean(((arr - arr.mean()) / arr.std()) ** 4)) - 3 if arr.std() > 0 else 0.0,
    }
