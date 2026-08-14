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
            certified=True,  # omit to use the account default (default_certified)
        )
    except APIError as e:
        raise SystemExit(f"{e.code}: {e.message}") from e

    print(f"id={msg['id']} state={msg['state']} hash={msg.get('canonical_hash')} "
          f"certified={msg.get('certified')} replay={replay}")
    rates = client.rates(7)
    print(
        f"rates(7d): delivery={rates.get('delivery_rate')}% "
        f"bounce={rates.get('bounce_rate')}% ingested={rates.get('ingested')}"
    )

    # certified flow: seal → evidence → verify
    try:
        seal = client.seal_message(msg["id"])
        print(f"sealed: merkle_root={seal.get('merkle_root')} "
              f"certificate_id={seal.get('certificate_id')}")
        evidence = client.message_evidence(msg["id"])
        print(f"evidence: leaf_index={evidence.get('leaf_index')} "
              f"mime_sha256={evidence.get('mime_sha256')}")
        result = client.verify_message(msg["id"])
        print(f"verify: valid={result.get('valid')}")
    except APIError as e:
        # 422 not_certified / missing_proof_data until the message is sealed
        print(f"seal/evidence/verify skipped: {e.code}: {e.message}")


if __name__ == "__main__":
    main()
