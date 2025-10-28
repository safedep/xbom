package command

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/safedep/dry/packageregistry/artifactv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestNpmPackage creates a test npm package tarball with the given files
func createTestNpmPackage(t *testing.T, files map[string]string) []byte {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for path, content := range files {
		header := &tar.Header{
			Name:    path,
			Mode:    0o644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}

		require.NoError(t, tarWriter.WriteHeader(header))
		_, err := tarWriter.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzipWriter.Close())

	return buf.Bytes()
}

// createMockNpmRegistry creates an httptest server that mocks the NPM registry
func createMockNpmRegistry(t *testing.T, packageData map[string][]byte, statusCodes map[string]int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NPM registry tarball endpoint: /{package}/-/{package}-{version}.tgz
		// Example: /express/-/express-4.17.1.tgz

		// Check if we have a custom status code for this path
		if code, ok := statusCodes[r.URL.Path]; ok {
			w.WriteHeader(code)
			if code != http.StatusOK {
				_, _ = w.Write([]byte("error"))
				return
			}
		}

		// Return package data if available
		if data, ok := packageData[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}

		// Default 404
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
}

// mockRoundTripper redirects all HTTP requests to the mock server
type mockRoundTripper struct {
	mockServerURL string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect all requests to mock server while preserving the path
	mockURL := m.mockServerURL + req.URL.Path
	mockReq, err := http.NewRequest(req.Method, mockURL, req.Body)
	if err != nil {
		return nil, err
	}

	// Copy headers
	mockReq.Header = req.Header

	// Execute request against mock server
	return http.DefaultClient.Do(mockReq)
}

func TestPackagePull(t *testing.T) {
	tests := []struct {
		name         string
		purl         string
		baseDir      string
		setupMock    func(t *testing.T) *httptest.Server
		wantErr      bool
		errContains  string
		verifyResult func(t *testing.T, resp *PackagePullResponse, baseDir string)
	}{
		{
			name:    "success - npm package with auto temp dir",
			purl:    "pkg:npm/express@4.17.1",
			baseDir: "",
			setupMock: func(t *testing.T) *httptest.Server {
				testPackage := createTestNpmPackage(t, map[string]string{
					"package/index.js":     "console.log('express');",
					"package/package.json": `{"name": "express", "version": "4.17.1"}`,
					"package/README.md":    "# Express",
				})

				// Mock the NPM registry endpoint
				packageData := map[string][]byte{
					"/express/-/express-4.17.1.tgz": testPackage,
				}

				return createMockNpmRegistry(t, packageData, nil)
			},
			wantErr: false,
			verifyResult: func(t *testing.T, resp *PackagePullResponse, baseDir string) {
				require.NotNil(t, resp)

				// Verify LocalPath returns a valid path
				localPath, err := resp.LocalPath()
				require.NoError(t, err)
				assert.NotEmpty(t, localPath)

				// Verify the directory exists
				info, err := os.Stat(localPath)
				require.NoError(t, err)
				assert.True(t, info.IsDir())

				// Verify at least one file exists
				entries, err := os.ReadDir(localPath)
				require.NoError(t, err)
				assert.NotEmpty(t, entries)

				// Clean up
				defer func() { _ = resp.Close() }()
			},
		},
		{
			name:    "success - npm package with custom base dir",
			purl:    "pkg:npm/lodash@4.17.21",
			baseDir: "",
			setupMock: func(t *testing.T) *httptest.Server {
				testPackage := createTestNpmPackage(t, map[string]string{
					"package/lodash.js":    "module.exports = {};",
					"package/package.json": `{"name": "lodash", "version": "4.17.21"}`,
				})

				packageData := map[string][]byte{
					"/lodash/-/lodash-4.17.21.tgz": testPackage,
				}

				return createMockNpmRegistry(t, packageData, nil)
			},
			wantErr: false,
			verifyResult: func(t *testing.T, resp *PackagePullResponse, baseDir string) {
				require.NotNil(t, resp)

				localPath, err := resp.LocalPath()
				require.NoError(t, err)

				// Verify the path is under the custom base dir
				assert.True(t, strings.HasPrefix(localPath, baseDir))

				// Clean up
				defer func() { _ = resp.Close() }()
			},
		},
		{
			name:    "failure - invalid PURL format",
			purl:    "invalid-purl-string",
			baseDir: "",
			setupMock: func(t *testing.T) *httptest.Server {
				// No mock needed - should fail before HTTP request
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Error("should not make HTTP request for invalid PURL")
				}))
			},
			wantErr:     true,
			errContains: "failed to parse PURL",
		},
		{
			name:    "failure - HTTP 404 package not found",
			purl:    "pkg:npm/nonexistent@1.0.0",
			baseDir: "",
			setupMock: func(t *testing.T) *httptest.Server {
				statusCodes := map[string]int{
					"/nonexistent/-/nonexistent-1.0.0.tgz": http.StatusNotFound,
				}
				return createMockNpmRegistry(t, nil, statusCodes)
			},
			wantErr:     true,
			errContains: "failed to fetch package",
		},
		{
			name:    "failure - HTTP 500 server error",
			purl:    "pkg:npm/error-package@1.0.0",
			baseDir: "",
			setupMock: func(t *testing.T) *httptest.Server {
				statusCodes := map[string]int{
					"/error-package/-/error-package-1.0.0.tgz": http.StatusInternalServerError,
				}
				return createMockNpmRegistry(t, nil, statusCodes)
			},
			wantErr:     true,
			errContains: "failed to fetch package",
		},
		{
			name:    "failure - invalid tarball data",
			purl:    "pkg:npm/corrupt@1.0.0",
			baseDir: "",
			setupMock: func(t *testing.T) *httptest.Server {
				// Return invalid tarball data
				packageData := map[string][]byte{
					"/corrupt/-/corrupt-1.0.0.tgz": []byte("invalid tarball data"),
				}
				return createMockNpmRegistry(t, packageData, nil)
			},
			wantErr:     true,
			errContains: "failed to extract artifact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			mockServer := tt.setupMock(t)
			defer mockServer.Close()

			// Create a temp base directory for tests that need it
			baseDir := tt.baseDir
			if tt.baseDir == "" && !tt.wantErr {
				var err error
				baseDir, err = os.MkdirTemp("", "xbom-test-*")
				require.NoError(t, err)
				defer func() { _ = os.RemoveAll(baseDir) }()
			}

			// Create custom HTTP client that redirects to mock server
			mockClient := &http.Client{
				Transport: &mockRoundTripper{
					mockServerURL: mockServer.URL,
				},
			}

			// Create request with custom HTTP client for testing
			req := PackagePullRequest{
				PURL:    tt.purl,
				BaseDir: baseDir,
				AdapterOpts: []artifactv2.Option{
					artifactv2.WithHTTPClient(mockClient),
				},
			}

			// Execute
			resp, err := PackagePull(context.Background(), req)

			// Verify error cases
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, resp)
				return
			}

			// Verify success cases
			require.NoError(t, err)
			require.NotNil(t, resp)

			if tt.verifyResult != nil {
				tt.verifyResult(t, resp, baseDir)
			}
		})
	}
}

func TestPackagePullResponse_LocalPath(t *testing.T) {
	tests := []struct {
		name     string
		response *PackagePullResponse
		wantPath string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "returns local path when set",
			response: &PackagePullResponse{
				localDir: "/tmp/test-path",
			},
			wantPath: "/tmp/test-path",
			wantErr:  false,
		},
		{
			name: "returns error when local path is empty",
			response: &PackagePullResponse{
				localDir: "",
			},
			wantErr: true,
			errMsg:  "no local path available for artifact content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := tt.response.LocalPath()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, path)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPath, path)
			}
		})
	}
}
