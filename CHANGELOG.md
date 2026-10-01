# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Tag scheme

Go (and any other SDK that lives in a subdirectory module) uses tags of the form
`<sdk-dir>/vX.Y.Z`. Example: `go/v0.1.0` installs as
`github.com/jaak-ai/mailack-sdk/go@v0.1.0`.

Other language packages may use unprefixed `vX.Y.Z` tags when they publish from
the repository root.

## [go/v0.1.0] - 2026-10-01

### Added

- `SendRequest.Attachments` and `NewAttachment` helper so `POST /v1/messages`
  matches the Mailack API attachment contract (TO-1204). Limits: max 25
  attachments; assembled message size bound 25 MiB. Batch send has no
  attachments field.
- Optional per-message `certified` flag on send/batch; online evidence APIs
  (`Seal`, `Evidence`, `ProofBundle`, `Verify`).
- Raw evidence download: `GetMessageRaw` (`.eml`) and `GetEventRaw` with
  integrity hashes from response headers (TO-904).

[go/v0.1.0]: https://github.com/jaak-ai/mailack-sdk/releases/tag/go/v0.1.0
