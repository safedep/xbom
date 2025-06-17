<div align="center">
  <h1>xBom</h1>
  
  <p><strong>Generate BOMs enriched with AI, SaaS and more using Static Code Analysis</strong></p>
</div>

<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/safedep/xbom)](https://goreportcard.com/report/github.com/safedep/xbom)
[![License](https://img.shields.io/github/license/safedep/xbom)](https://github.com/safedep/xbom/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/safedep/xbom)](https://github.com/safedep/xbom/releases)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/safedep/xbom/badge)](https://api.securityscorecards.dev/projects/github.com/safedep/xbom)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)
[![CodeQL](https://github.com/safedep/xbom/actions/workflows/codeql.yml/badge.svg?branch=main)](https://github.com/safedep/xbom/actions/workflows/codeql.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/safedep/xbom.svg)](https://pkg.go.dev/github.com/safedep/xbom)

</div>

## ⚡ Quick Start

```bash
# Installation on macOS & Linux
brew install safedep/tap/xbom
```

or download a **[pre-built binary](https://github.com/safedep/xbom/releases)**


```bash
# Generate BOM for your source code
xbom generate --dir /path/to/code --bom /path/to/bom.cdx.json
```

This will generate a [CycloneDX v1.6](https://cyclonedx.org/docs/1.6/json/) SBOM with AI components detected in the code base.

<div align="center">
  <img src="./docs/assets/xbom-cli.png" alt="xbom-cli" width="100%" />
</div>


## Supported Languages
Currently, xBom supports the following programming languages:

| Language | Status |
|-----------|--------|
| Python      | ✅ Active |
| Java      | ✅ Active |
| JavaScript      | 🚧 WIP |

## Limitations

`xbom` is currently limited to AI BOM generation only. It uses static code analysis to identify AI products used in the code base. For generating a full-fledged SBOM with library dependencies, you can use [vet](https://github.com/safedep/vet).

## Development

### Signatures

xBom maintains community-driven signatures for popular SDKs, APIs and libraries in `signatures/` following file naming convention - `signatures/$vendor/$product/$service.yml`


## 🤝 Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md)

For contributing new signatures, refer [this](CONTRIBUTING.md#contributing-signatures)


## Telemetry

`xbom` collects anonymous telemetry to help us understand how it is used and
improve the product. To disable telemetry, set `XBOM_DISABLE_TELEMETRY` environment
variable to `true`.

```bash
export XBOM_DISABLE_TELEMETRY=true
```

## 👀 Visual overview

We generate BOMs as JSON files following [CycloneDX SPIEC](https://cyclonedx.org/docs/1.6/json/). For a quick overview, you can view the BOM in an interactive HTML output linked in console output.

<div align="center">
  <img src="./docs/assets/xbom-demo.gif" alt="xbom-demo" width="100%" />
</div>
