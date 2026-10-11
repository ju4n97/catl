---
title: regi vendor
description: Vendor dependencies into .regi/vendor for offline use
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi vendor

Vendor dependencies into .regi/vendor for offline use

Copies all remote dependencies and referenced blobs into .regi/vendor/.
Local path dependencies are not copied.

**Category:** Dependencies

## Usage

```bash
regi vendor [flags]
```

## Flags

| Flag           | Description                                         |
| :------------- | :-------------------------------------------------- |
| `--skip-blobs` | Skip downloading large blob payloads when vendoring |
