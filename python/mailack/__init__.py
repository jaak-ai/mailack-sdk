"""Official Python SDK for the mailack certified email API."""

from .client import Client
from .errors import APIError

__all__ = ["Client", "APIError", "__version__"]
__version__ = "0.1.0"
