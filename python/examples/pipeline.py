"""Nexus cross-language pipeline demo — Python analytics stage."""

import json
import sys
import time

import numpy as np

sys.path.insert(0, "src")

from nexus import export, register
from nexus.numpy_bridge import analyze_array, fft_analyze, stats_report, array_to_value


@export(name="analytics.summary")
def analytics_summary(data: list[float]) -> dict:
    """Generate a statistical summary of input data."""
    report = stats_report(data)
    return {
        "status": "ok",
        "language": "python",
        "library": "numpy",
        "report": report,
    }


@export(name="analytics.fft")
def analytics_fft(data: list[float]) -> dict:
    """Run FFT analysis on 1D data."""
    arr = np.array(data, dtype=np.float64)
    result = fft_analyze(arr)
    return {
        "status": "ok",
        "language": "python",
        "fft": result,
    }


@export(name="analytics.transform")
def analytics_transform(data: list[float], scalar: float = 2.0) -> dict:
    """Transform data using SIMD-style multiplication."""
    arr = np.array(data, dtype=np.float64)
    transformed = arr * scalar
    return {
        "status": "ok",
        "language": "python",
        "original_size": len(data),
        "scalar": scalar,
        "sample": transformed[:10].tolist(),
        "statistics": analyze_array(transformed),
    }


if __name__ == "__main__":
    print("╔══════════════════════════════════════════╗")
    print("║   Nexus Python Analytics Pipeline        ║")
    print("╚══════════════════════════════════════════╝")

    np.random.seed(42)
    data = np.random.randn(10000).cumsum().tolist()

    print(f"\n1. Input: 10,000 data points generated")

    start = time.time()
    summary = analytics_summary(data)
    elapsed = (time.time() - start) * 1000
    report = summary["report"]
    print(f"   Python analysis ({elapsed:.1f}ms):")
    print(f"   Mean={report['mean']:.4f}  Std={report['std']:.4f}")
    print(f"   Min={report['min']:.4f}  Max={report['max']:.4f}")
    print(f"   Skew={report['skewness']:.4f}  Kurtosis={report['kurtosis']:.4f}")

    start = time.time()
    fft = analytics_fft(data.tolist())
    elapsed = (time.time() - start) * 1000
    print(f"\n2. FFT analysis ({elapsed:.1f}ms):")
    print(f"   Dominant freq={fft['fft']['dominant_freq']:.4f}")
    print(f"   Energy={fft['fft']['energy']:.2f}")

    start = time.time()
    transformed = analytics_transform(data.tolist(), 2.5)
    elapsed = (time.time() - start) * 1000
    print(f"\n3. Transform ({elapsed:.1f}ms):")
    print(f"   First 5 transformed values: {transformed['sample'][:5]}")

    print(f"\n4. Registered Nexus functions:")
    for name in ["analytics.summary", "analytics.fft", "analytics.transform"]:
        print(f"   - {name}")

    print(f"\n{'='*50}")
    print(f"Pipeline demonstrates Python (numpy) as the analytics stage.")
    print(f"Data flows: Go (fetch) → Rust (SIMD transform) → Python (analytics)")
    print(f"          → TS (visualization)")
    print(f"{'='*50}")
