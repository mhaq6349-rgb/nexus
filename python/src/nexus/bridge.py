"""Nexus bridge — connects Python to the Go daemon and Rust shared library."""

import json
import os
import struct
import sys
from typing import Any

import requests

from .types import Value, value_from, value_to, serialize_value, LANG_PYTHON, MSG_CALL, MSG_RETURN, MSG_ERROR


class NexusClient:
    """HTTP client to the Nexus Go daemon."""

    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url.rstrip("/")
        self._session = requests.Session()
        self._call_id = 0

    def call(self, function: str, *args: Any, timeout: float = 10.0) -> Any:
        payload = {"function": function, "args": list(args)}
        resp = self._session.post(
            f"{self.base_url}/api/v1/call",
            json=payload,
            timeout=timeout,
        )
        resp.raise_for_status()
        data = resp.json()
        if "error" in data:
            raise RuntimeError(f"Nexus call error: {data['error']}")
        return data.get("result")

    def stats(self) -> dict:
        resp = self._session.get(f"{self.base_url}/api/v1/stats")
        resp.raise_for_status()
        return resp.json()

    def health(self) -> dict:
        resp = self._session.get(f"{self.base_url}/api/v1/health")
        resp.raise_for_status()
        return resp.json()

    def runtimes(self) -> dict:
        resp = self._session.get(f"{self.base_url}/api/v1/runtimes")
        resp.raise_for_status()
        return resp.json()

    def close(self):
        self._session.close()


class NexusBridge:
    """Direct bridge to the Rust shared library via ctypes.

    Falls back to HTTP client if the shared library is not available.
    """

    def __init__(self, daemon_url: str = "http://localhost:8080"):
        self._rust_lib = None
        self._client = NexusClient(daemon_url)
        self._init_rust()

    def _init_rust(self):
        lib_path = os.environ.get("NEXUS_RUST_LIB")
        if lib_path and os.path.exists(lib_path):
            import ctypes
            try:
                self._rust_lib = ctypes.CDLL(lib_path)
                self._rust_lib.nexus_init()
                print(f"[nexus] Rust shared library loaded: {lib_path}", file=sys.stderr)
            except Exception as e:
                print(f"[nexus] Failed to load Rust lib: {e}", file=sys.stderr)
                self._rust_lib = None

    @property
    def has_rust(self) -> bool:
        return self._rust_lib is not None

    def call(self, function: str, *args, use_rust: bool = False, **kwargs) -> Any:
        if use_rust and self.has_rust:
            return self._call_rust(function, *args)
        return self._client.call(function, *args, **kwargs)

    def _call_rust(self, function: str, *args) -> Any:
        if not self._rust_lib:
            raise RuntimeError("Rust library not loaded")
        import ctypes
        self._call_id += 1
        call_id = self._call_id
        values = [value_from(a) for a in args]
        args_data = b""
        for v in values:
            args_data += serialize_value(v)
        func_c = ctypes.create_string_buffer(function.encode())
        result = self._rust_lib.nexus_write_call(
            call_id, func_c, args_data, len(args_data), LANG_PYTHON,
        )
        if result != 0:
            raise RuntimeError(f"Rust write_call failed: {result}")
        return {"status": "dispatched", "call_id": call_id, "function": function}

    def stats(self) -> dict:
        return self._client.stats()

    def health(self) -> dict:
        return self._client.health()

    def close(self):
        self._client.close()

    def __enter__(self):
        return self

    def __exit__(self, *args):
        self.close()
