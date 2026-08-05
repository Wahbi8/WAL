# WAL

A Write-Ahead Log (WAL) implementation in Go. A WAL is a durable, append-only log used by databases and storage engines to guarantee crash recovery: all writes are persisted to the log before they are considered committed.

> **Status: work in progress.** Core serialization, fragmentation, and integrity checks are implemented, but the read/recovery path, durability (`fsync`), and tests are not yet complete.

## Features

- Append-only record storage in 32 KB blocks
- Records larger than a block are split into `start` / `middle` / `end` fragments and reassembled on read
- CRC32 checksums for per-fragment and whole-record integrity verification
- Key-value payload encoding (`operation | keyLength | valueLength | key | value`)
- Little-endian binary format, no external dependencies

## On-disk format

Each record is written as:

```
+----------+----------+---------+---------+------------+
| recordId | checkSum | logType | length  |  payload   |
| 2 bytes  | 4 bytes  | 1 byte  | 4 bytes |  N bytes   |
+----------+----------+---------+---------+------------+
```

- `recordId` — unique record ID (`uint16`); all fragments of one record share the same ID
- `checkSum` — CRC32 of the payload (whole payload for `full`/`start`, chunk for `middle`/`end`)
- `logType` — `0` full, `1` start, `2` middle, `3` end
- `length` — payload length in bytes
- `payload` — `operation (1 byte) | keyLength (2 bytes) | valueLength (4 bytes) | key | value`

## Getting started

```bash
go build ./...
```

Requires Go 1.23.4+.

## Current API

- `OpenFile(path)` — open/create the log file for appending
- `Store.serialize(payload)` — write a payload, fragmenting it if it exceeds one block
- `FragmentReassembler.Assemble(r)` — reassemble fragmented records and verify checksums
- `deserializeHeader(bytes)` — parse a record header
- `parseRecord(data, r)` — decode a payload into its key-value parts

