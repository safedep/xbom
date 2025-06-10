package main

import (
	"embed"

	"github.com/safedep/xbom/pkg/signatures"
)

//go:embed signatures
var signatureFS embed.FS

func init() {
	signatures.SetEmbeddedSignatureFS(signatureFS)
}
