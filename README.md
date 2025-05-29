# xbom
Generate BOMs enriched with AI, SaaS and more using Static Code Analysis

## 📑 Table of Contents
- [Usage](#usage)
- [Supported Ecosystems](#supported-ecosystems)
- [Limitations](#limitations)
- [Contributing](#contributing)

## Usage

Generate an AI BOM from source code:

```bash
xbom generate --dir /path/to/code
```

This will by default generate a console statistics of different AI products used in the code base.

```bash
xbom generate --dir /path/to/code --cdx /path/to/sbom.cdx.json
```

This will generate a CycloneDX SBOM with AI components detected in the code base.

## Supported Languages
Currently, xBom supports the following programming languages:

| Language | Status |
|-----------|--------|
| Python      | ✅ Active |

## Limitations

`xbom` is currently limited to AI BOM generation only. It uses static code analysis to identify AI products used in the code base. For generating a SBOM for library dependencies, you can use [vet](https://github.com/safedep/vet).

## Development



### Signatures

xBom maintains community-driven signatures for popular SDKs, APis and libraries in `signatures/` following file naming convention - `signatures/$vendor/$product/$service.yml` You can generate a new signature file using command -

```bash
xbom signature new --vendor <vendor> --product <product> --service <name>
```

This will generate a new YAML (if it doesn't exist) file in `signatures/$vendor/$product/$service.yml`. Edit the file to add the necessary patterns to detect the component.

Examples:

```
signatures/microsoft/azure/ai.yml
signatures/microsoft/office/integrations.yml
```
