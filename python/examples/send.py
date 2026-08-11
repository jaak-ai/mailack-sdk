#!/usr/bin/env python3
"""Example: send one certified message.

  export MAILACK_API_URL=http://localhost:8080
  export MAILACK_API_KEY=mlk_…
  python examples/send.py
"""
from __future__ import annotations

import os
import sys
from datetime import datetime, timezone

# monorepo: allow running without install
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from mailack import APIError, Client


def main() -> None:
    base = os.environ.get("MAILACK_API_URL", "http://localhost:8080")
    key = os.environ.get("MAILACK_API_KEY")
    if not key:
        raise SystemExit("MAILACK_API_KEY is required")

    client = Client(base, api_key=key)
    key_id = "py-sdk-" + datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S")
    try:
        msg, replay = client.send(
            key_id,
            from_=os.environ.get("MAILACK_FROM", "noreply@example.com"),
            to=os.environ.get("MAILACK_TO", "you@example.com"),
            subject="mailack Python SDK example",
            text="Hello from the Python SDK.",
        )
    except APIError as e:
        raise SystemExit(f"{e.code}: {e.message}") from e

    print(f"id={msg['id']} state={msg['state']} hash={msg.get('canonical_hash')} replay={replay}")
    rates = client.rates(7)
    print(
        f"rates(7d): delivery={rates.get('delivery_rate')}% "
        f"bounce={rates.get('bounce_rate')}% ingested={rates.get('ingested')}"
    )


if __name__ == "__main__":
    main()
