# xbom
AI BOM using Static Code Analysis

## Usage

Generate an AI BOM from source code:

```bash
xbom generate --code /path/to/code
```

This will by default generate a console statistics of different AI products used in the code base.

```bash
xbom generate --code /path/to/code --cdx /path/to/sbom.cdx.json
```

This will generate a CycloneDX SBOM with AI components detected in the code base.

## Limitations

`xbom` is currently limited to AI BOM generation only. It uses static code analysis to identify AI products used in the code base. For generating a SBOM for library dependencies, you can use [vet](https://github.com/safedep/vet).

## Development

### Signature

```bash
xbom signature new --vendor <vendor> --product <product> --service <name>
```

This will generate a new YAML (if it doesn't exist) file in `signatures/$vendor/$product/$service.yml`. Edit the file to add the necessary patterns to detect the component.

Examples:

```
signatures/openai/api/sdk.yml
signatures/google/gcp/vertexai.yml
signatures/amazon/aws/bedrock.yml
```
