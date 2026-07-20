"""Nexus type system — cross-language value representation."""

from __future__ import annotations

import struct
from dataclasses import dataclass, field
from typing import Any

MSG_CALL = 0x0001
MSG_RETURN = 0x0002
MSG_ERROR = 0x0003
MSG_PING = 0x00FE
MSG_PONG = 0x00FF

LANG_RUST = 0
LANG_GO = 1
LANG_PYTHON = 2
LANG_TS = 3


@dataclass
class NdArray:
    dtype: int
    shape: list[int]
    data: bytes


@dataclass
class Value:
    type: str = "null"
    bool_val: bool = False
    i64: int = 0
    u64: int = 0
    f64: float = 0.0
    string: str = ""
    data: bytes = b""
    items: list[Value] = field(default_factory=list)
    mapping: dict[str, Value] = field(default_factory=dict)
    ndarray: NdArray | None = None

    @staticmethod
    def from_f64(v: float) -> "Value":
        return Value(type="f64", f64=v)

    @staticmethod
    def from_str(s: str) -> "Value":
        return Value(type="string", string=s)

    @staticmethod
    def from_i64(v: int) -> "Value":
        return Value(type="i64", i64=v)

    @staticmethod
    def from_bytes(b: bytes) -> "Value":
        return Value(type="bytes", data=b)

    @staticmethod
    def null() -> "Value":
        return Value(type="null")


def value_from(py_value: Any) -> Value:
    if py_value is None:
        return Value.null()
    if isinstance(py_value, bool):
        return Value(type="bool", bool_val=py_value)
    if isinstance(py_value, int):
        return Value.from_i64(py_value)
    if isinstance(py_value, float):
        return Value.from_f64(py_value)
    if isinstance(py_value, str):
        return Value.from_str(py_value)
    if isinstance(py_value, bytes):
        return Value.from_bytes(py_value)
    if isinstance(py_value, list):
        return Value(type="list", items=[value_from(v) for v in py_value])
    if isinstance(py_value, dict):
        return Value(
            type="map",
            mapping={k: value_from(v) for k, v in py_value.items()},
        )
    return Value.from_str(str(py_value))


def value_to(v: Value) -> Any:
    if v.type == "null":
        return None
    if v.type == "bool":
        return v.bool_val
    if v.type in ("i64", "u64"):
        return v.i64 if v.type == "i64" else v.u64
    if v.type == "f64":
        return v.f64
    if v.type == "string":
        return v.string
    if v.type == "bytes":
        return v.data
    if v.type == "list":
        return [value_to(item) for item in v.items]
    if v.type == "map":
        return {k: value_to(val) for k, val in v.mapping.items()}
    if v.type == "ndarray" and v.ndarray:
        import numpy as np
        dtype_map = {0: np.float32, 1: np.float64, 2: np.int32, 3: np.int64}
        np_dtype = dtype_map.get(v.ndarray.dtype, np.float32)
        arr = np.frombuffer(v.ndarray.data, dtype=np_dtype)
        if v.ndarray.shape:
            arr = arr.reshape(v.ndarray.shape)
        return arr
    return str(v)


def serialize_value(v: Value) -> bytes:
    tag = {
        "null": 0, "bool": 1, "i64": 2, "u64": 3,
        "f64": 4, "string": 5, "bytes": 6, "list": 7,
        "map": 8, "ndarray": 9,
    }.get(v.type, 0)
    if tag == 0:
        return struct.pack("<B", 0)
    if tag == 1:
        return struct.pack("<BB", 1, 1 if v.bool_val else 0)
    if tag == 2:
        return struct.pack("<Bq", 2, v.i64)
    if tag == 4:
        return struct.pack("<Bd", 4, v.f64)
    if tag == 5:
        encoded = v.string.encode("utf-8")
        return struct.pack("<BQ", 5, len(encoded)) + encoded
    if tag == 6:
        return struct.pack("<BQ", 6, len(v.data)) + v.data
    return struct.pack("<Bd", 4, 0.0)


def deserialize_value(data: bytes, offset: int = 0) -> tuple[Value, int]:
    if offset >= len(data):
        return Value.null(), offset
    tag = data[offset]
    if tag == 0:
        return Value.null(), offset + 1
    if tag == 1:
        return Value(type="bool", bool_val=bool(data[offset + 1])), offset + 2
    if tag == 2:
        val = struct.unpack_from("<q", data, offset + 1)[0]
        return Value.from_i64(val), offset + 9
    if tag == 4:
        val = struct.unpack_from("<d", data, offset + 1)[0]
        return Value.from_f64(val), offset + 9
    if tag == 5:
        length = struct.unpack_from("<Q", data, offset + 1)[0]
        s = data[offset + 9:offset + 9 + length].decode("utf-8")
        return Value.from_str(s), offset + 9 + length
    if tag == 6:
        length = struct.unpack_from("<Q", data, offset + 1)[0]
        b = data[offset + 9:offset + 9 + length]
        return Value.from_bytes(b), offset + 9 + length
    return Value.null(), offset + 1
