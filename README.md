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

## Supported Languages
Currently, xBom supports the following programming languages:

| Language | Status |
|-----------|--------|
| Python      | ✅ Active |
| Java      | ✅ Active |
| JavaScript      | 🚧 WIP |

## Supported BOMs

<div align="center">
  <h1><strong>AI</strong></h1>
  <table>
    <tr>
      <td align="center" width="250">
        <img src="https://github.com/user-attachments/assets/832ae2cf-a807-42b9-8ae7-6366f43f9d05" alt="LangChain" width="202"/>
      </td>
      <td align="center" width="250">
        <img src="https://github.com/user-attachments/assets/f69820cf-c76b-44b0-850c-9cc1df4830a9" alt="Anthropic" width="200"/>
      </td>
    </tr>
    <tr>
      <td align="center" width="250">
        <img src="https://cdn.prod.website-files.com/66cf2bfc3ed15b02da0ca770/66d07240057721394308addd_Logo%20(1).svg" alt="CrewAI" width="200"/>
      </td>
      <td align="center" width="250">
        <img src="https://github.com/user-attachments/assets/789f75fb-a465-4b05-b3c4-2ccee204b1df" alt="OpenAI" width="200"/>
      </td>
    </tr>
  </table>
</div>
<div align="center">
  <h1><strong>Cloud</strong></h1>
  <table>
    <tr>
      <td align="center" width="150">
        <img src="https://pendulum-it.com/wp-content/uploads/2020/05/Google-Cloud-Platform-GCP-logo.png" alt="GCP" width="100"/>
      </td>
      <td align="center" width="150">
        <img src="https://swimburger.net/media/fbqnp2ie/azure.svg" alt="Azure" width="100"/>
      </td>
    </tr>
  </table>
</div>

## Limitations

`xbom` is currently limited to AI BOM generation only. It uses static code analysis to identify AI products used in the code base. For generating a full-fledged SBOM with library dependencies, you can use [vet](https://github.com/safedep/vet).

## Development

### Signatures

xBom maintains community-driven signatures for popular SDKs, APIs and libraries in `signatures/` following file naming convention - `signatures/$vendor/$product/$service.yml` You can generate a new signature file using command -

```bash
xbom signature new --vendor <vendor> --product <product> --service <name>
```

This will generate a new YAML (if it doesn't exist) file in `signatures/$vendor/$product/$service.yml`. Edit the file to add the necessary patterns to detect the component.

Examples:

```
signatures/microsoft/azure/ai.yml
signatures/microsoft/office/integrations.yml
```

## Telemetry

`xbom` collects anonymous telemetry to help us understand how it is used and
improve the product. To disable telemetry, set `XBOM_DISABLE_TELEMETRY` environment
variable to `true`.

```bash
export XBOM_DISABLE_TELEMETRY=true
```
