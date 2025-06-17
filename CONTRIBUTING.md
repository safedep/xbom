
# Contributing Guide

You can contribute to `xbom` and help make it better. Apart from signatures, bug fixes, 
features, we particularly value contributions in the form of:

- Documentation improvements
- Bug reports
- Using `xbom` to generate BOM for your projects

## How to contribute

1. Fork the repository
2. Add your changes
3. Submit a pull request

## How to report a bug

Create a new issue and add the label `bug`.

## How to suggest a new feature

Create a new issue and add the label `enhancement`.

## Development workflow

When contributing changes to repository, follow these steps:

1. Ensure tests are passing
2. Ensure you write test cases for new code
3. `Signed-off-by` line is required in commit message (use `-s` flag while committing)

## Developer Setup

### Requirements

* Go 1.24.3+
* Git
* Make

### Install Dependencies

* Install [ASDF](https://asdf-vm.com/)
* Install the development tools

```bash
asdf plugin add golang
asdf plugin add gitleaks
asdf install
```

* Install `lefthook`

```bash
go install github.com/evilmartians/lefthook@latest
```

* Install git hooks

```bash
$(go env GOPATH)/bin/lefthook install
```

### Build

Install dependencies

```bash
go mod download
```

Build `xbom`

```bash
make
```

### Run Tests

```bash
make test
```

## Contributing Signatures

xBom maintains community-driven signatures for popular SDKs, APIs and libraries in `signatures/` following file naming convention - `signatures/$vendor/$product/$service.yml`
You can contribute signatures by opening a PR with new signatures in existing/new signature files in this directory

### Validate new signatures

```bash
# Build xbom
make

# Validate signatures
./bin/xbom validate
```
