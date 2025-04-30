# ADR

The scope of this project is to generate `xBOM` using code analysis. This project's primary goal is to solve the developer experience problem of leveraging [CAF](https://github.com/safedep/code) and using custom signatures for analyzing code base and identify application specific components such as:

- AI usage (AI BOM)
- SaaS usage (SaaS BOM)
- Cryptography usage (Crypto BOM)
- ML Model usage (ML BOM)
- Etc.

We will restrict this project to information that can be generated through code analysis. We may however choose to integrate other tools to collection additional information such as dependencies from `package-lock.json` or `requirements.txt` if required to enrich the generated BOM.

## Design Decisions

### Primary Goals

- Make it very easy to generate a list of AI (or other) products used in a code base
- Make it very easy to write signatures to detect custom application specific components
- Make it very easy to generate reports using standard formats like CycloneDX

### Out of Scope

This project will not evolve into a general purpose SBOM or SCA tool.

## Consequences

The `xbom` project will be a `cli` tool that focusses primarily on developer experience. It may or may not eventually get integrated with [vet](https://github.com/safedep/vet) to enrich the generated BOM with additional information.