package command

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/safedep/dry/api/pb"
	"github.com/safedep/dry/packageregistry/artifactv2"
	"github.com/safedep/dry/storage"
)

// PackagePullRequest is the extensible request struct for the operation
type PackagePullRequest struct {
	// The PURL of the package. Must be provided.
	PURL string

	// Optional base dir to pull to. Defaults to a temp dir
	BaseDir string

	// Optional additional adapter options (useful for testing)
	AdapterOpts []artifactv2.Option
}

// PackagePullResponse is the extensible response struct for the operation
type PackagePullResponse struct {
	localDir string

	// References for cleanup
	st     storage.Storage
	reader artifactv2.ArtifactReaderV2
}

// LocalPath returns the local dir path where the artfact was extracted.
func (pp *PackagePullResponse) LocalPath() (string, error) {
	if pp.localDir == "" {
		return "", fmt.Errorf("no local path available for artifact content")
	}

	return pp.localDir, nil
}

// Close closes the response and releases the resources
func (pp *PackagePullResponse) Close() error {
	if err := pp.reader.Close(); err != nil {
		return err
	}

	// The storage interface doesn't support a closer. We will call it here
	// when supported

	return nil
}

// PackagePull creates a local cache of
func PackagePull(ctx context.Context, req PackagePullRequest) (*PackagePullResponse, error) {
	parsedPURL, err := pb.NewPurlPackageVersion(req.PURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PURL: %w", err)
	}

	baseDir := req.BaseDir
	if baseDir == "" {
		baseDir, err = os.MkdirTemp("", "xbom-ppull-*")
		if err != nil {
			return nil, fmt.Errorf("failed to generate tmp dir: %w", err)
		}
	}

	storage, err := storage.NewFilesystemStorageDriver(storage.FilesystemStorageDriverConfig{
		Root: baseDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create storage driver: %w", err)
	}

	// Build adapter options with defaults
	adapterOpts := []artifactv2.Option{
		artifactv2.WithCacheEnabled(true),
		artifactv2.WithPersistArtifacts(true),
		artifactv2.WithMetadataEnabled(true),
		artifactv2.WithStorage(storage),
	}

	// Append any additional options (e.g., for testing)
	adapterOpts = append(adapterOpts, req.AdapterOpts...)

	adapter, err := artifactv2.CreateAdapter(parsedPURL.Ecosystem(), adapterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create artifact adapter: %w", err)
	}

	reader, err := adapter.Fetch(ctx, artifactv2.ArtifactInfo{
		Ecosystem: parsedPURL.Ecosystem(),
		Name:      parsedPURL.Name(),
		Version:   parsedPURL.Version(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package: %w", err)
	}

	ex, err := reader.Extract(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract artifact to storage: %w", err)
	}

	absLocalPath := path.Join(baseDir, ex.ExtractionKey)
	return &PackagePullResponse{
		localDir: absLocalPath,
		st:       storage,
		reader:   reader,
	}, nil
}
