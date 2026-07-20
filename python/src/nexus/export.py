"""Nexus function export decorator and registry."""

import functools
import inspect
import json
from typing import Any, Callable

_registry: dict[str, Callable] = {}


class NexusFunction:
    """A function registered with Nexus for cross-language calling."""

    def __init__(self, fn: Callable, name: str | None = None, lang: str = "python"):
        self.fn = fn
        self.name = name or fn.__name__
        self.lang = lang
        self.signature = self._build_signature()
        _registry[self.name] = self

    def _build_signature(self) -> dict:
        sig = inspect.signature(self.fn)
        params = []
        for p in sig.parameters.values():
            params.append({
                "name": p.name,
                "kind": str(p.kind),
                "default": None if p.default is inspect.Parameter.empty else str(p.default),
            })
        return {
            "name": self.name,
            "params": params,
            "return_annotation": str(sig.return_annotation),
        }

    def __call__(self, *args, **kwargs) -> Any:
        return self.fn(*args, **kwargs)

    def to_json(self) -> str:
        return json.dumps({
            "name": self.name,
            "lang": self.lang,
            "signature": self.signature,
        })

    def __repr__(self) -> str:
        return f"<NexusFunction '{self.name}' ({self.lang})>"


def export(name: str | None = None, lang: str = "python"):
    """Decorator to register a function with the Nexus bridge.

    Usage:
        @nexus.export()
        def my_function(x: float, y: float) -> float:
            return x + y

        @nexus.export(name="math.custom_sqrt")
        def sqrt_impl(x: float) -> float:
            return x ** 0.5
    """
    def decorator(fn: Callable) -> NexusFunction:
        return NexusFunction(fn, name=name, lang=lang)
    return decorator


def register(fn: Callable, name: str | None = None) -> NexusFunction:
    """Register a function directly without the decorator."""
    return NexusFunction(fn, name=name)


def get_registry() -> dict[str, NexusFunction]:
    """Get all registered Nexus functions."""
    return dict(_registry)


def list_functions() -> list[dict]:
    """List all registered functions as JSON-serializable dicts."""
    return [json.loads(fn.to_json()) for fn in _registry.values()]
