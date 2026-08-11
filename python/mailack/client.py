from __future__ import annotations

import json
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass, field
from typing import Any, Mapping, Optional

from .errors import APIError


@dataclass
class Client:
    """HTTP client for one mailack API deployment.

    Example::

        client = Client("https://api.mailack.com", api_key="mlk_…")
        msg, replay = client.send("idem-1", from_="a@acme.mx", to="b@x.com",
                                  subject="Hi", text="Hello")
    """

    base_url: str
    api_key: str = ""
    timeout: float = 30.0

    def __post_init__(self) -> None:
        self.base_url = self.base_url.rstrip("/")

    # --- HTTP ----------------------------------------------------------------

    def _request(
        self,
        method: str,
        path: str,
        *,
        body: Any = None,
        idempotency_key: str = "",
        query: Optional[Mapping[str, str]] = None,
    ) -> tuple[int, dict[str, str], Any]:
        url = self.base_url + path
        if query:
            url += "?" + urllib.parse.urlencode({k: v for k, v in query.items() if v is not None})
        data = None
        headers = {"Accept": "application/json", "User-Agent": "mailack-python/0.1.0"}
        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"
        if idempotency_key:
            headers["Idempotency-Key"] = idempotency_key
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"

        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                raw = resp.read()
                headers_out = {k.lower(): v for k, v in resp.headers.items()}
                payload = json.loads(raw.decode("utf-8")) if raw else None
                return resp.status, headers_out, payload
        except urllib.error.HTTPError as e:
            raw = e.read()
            code, message = "http_error", raw.decode("utf-8", errors="replace")
            try:
                env = json.loads(raw.decode("utf-8"))
                err = env.get("error") or {}
                code = err.get("code") or code
                message = err.get("message") or message
            except Exception:
                pass
            raise APIError(e.code, code, message) from e

    # --- Messages ------------------------------------------------------------

    def send(
        self,
        idempotency_key: str,
        *,
        from_: str,
        to: str,
        subject: str = "",
        text: str = "",
        html: str = "",
        headers: Optional[dict[str, str]] = None,
        template_id: str = "",
        variables: Optional[dict[str, str]] = None,
    ) -> tuple[dict[str, Any], bool]:
        """POST /v1/messages. Returns (message, replay)."""
        body: dict[str, Any] = {"from": from_, "to": to}
        if template_id:
            body["template_id"] = template_id
            if variables:
                body["variables"] = variables
        else:
            body["subject"] = subject
            if text:
                body["text"] = text
            if html:
                body["html"] = html
        if headers:
            body["headers"] = headers
        status, hdrs, payload = self._request(
            "POST", "/v1/messages", body=body, idempotency_key=idempotency_key
        )
        replay = hdrs.get("idempotent-replay", "") == "true"
        return payload, replay

    def send_batch(self, messages: list[dict[str, Any]]) -> dict[str, Any]:
        """POST /v1/messages/batch (max 100). Each item needs idempotency_key."""
        _, _, payload = self._request("POST", "/v1/messages/batch", body={"messages": messages})
        return payload

    def get_message(self, message_id: str) -> dict[str, Any]:
        _, _, payload = self._request("GET", f"/v1/messages/{message_id}")
        return payload.get("message") or payload

    def rates(self, days: int = 14) -> dict[str, Any]:
        _, _, payload = self._request("GET", "/v1/rates", query={"days": str(days)})
        return payload

    # --- Domains -------------------------------------------------------------

    def list_domains(self) -> list[dict[str, Any]]:
        _, _, payload = self._request("GET", "/v1/domains")
        return payload.get("items") or []

    def create_domain(self, domain: str) -> dict[str, Any]:
        _, _, payload = self._request("POST", "/v1/domains", body={"domain": domain})
        return payload

    def verify_domain(self, domain_id: str) -> dict[str, Any]:
        _, _, payload = self._request("POST", f"/v1/domains/{domain_id}/verify", body={})
        return payload

    # --- Webhooks ------------------------------------------------------------

    def list_webhooks(self) -> tuple[list[dict[str, Any]], list[str]]:
        _, _, payload = self._request("GET", "/v1/webhooks")
        return payload.get("items") or [], payload.get("available_events") or []

    def create_webhook(
        self, url: str, events: list[str], description: str = ""
    ) -> tuple[dict[str, Any], str]:
        body: dict[str, Any] = {"url": url, "events": events}
        if description:
            body["description"] = description
        _, _, payload = self._request("POST", "/v1/webhooks", body=body)
        return payload.get("webhook") or {}, payload.get("secret") or ""

    def disable_webhook(self, webhook_id: str) -> None:
        self._request("DELETE", f"/v1/webhooks/{webhook_id}")

    # --- Templates -----------------------------------------------------------

    def list_templates(self) -> list[dict[str, Any]]:
        _, _, payload = self._request("GET", "/v1/templates")
        return payload.get("items") or []

    def create_template(
        self, name: str, subject: str, text: str = "", html: str = ""
    ) -> dict[str, Any]:
        _, _, payload = self._request(
            "POST",
            "/v1/templates",
            body={"name": name, "subject": subject, "text": text, "html": html},
        )
        return payload
