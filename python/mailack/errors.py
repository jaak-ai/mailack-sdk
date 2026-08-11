from __future__ import annotations


class APIError(Exception):
    """mailack API error envelope: {"error":{"code","message"}}."""

    def __init__(self, status: int, code: str, message: str) -> None:
        self.status = status
        self.code = code
        self.message = message
        super().__init__(f"mailack: HTTP {status} {code}: {message}")

    def is_code(self, code: str) -> bool:
        return self.code == code
