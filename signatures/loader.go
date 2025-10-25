// Package signatures contains the out of box signatures and a Go specific loader
// for easy use in Go applications.
package signatures

import (
	"embed"

	pkgsignatures "github.com/safedep/xbom/pkg/signatures"
)

//go:embed *
var embeddedSignatureFS embed.FS

func init() {
	pkgsignatures.SetEmbeddedSignatureFS(embeddedSignatureFS)
}
