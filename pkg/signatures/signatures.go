package signatures

import (
	"embed"
	"fmt"
	"io/fs"
	"path"

	callgraphv1 "buf.build/gen/go/safedep/api/protocolbuffers/go/safedep/messages/code/callgraph/v1"
	"github.com/safedep/code/plugin/callgraph"
	"github.com/safedep/dry/log"
	"gopkg.in/yaml.v3"
)

// signature file namespace - `pkg/signatures/data/$vendor/$product/$service.yml`. Edit the file to add the necessary patterns to detect the component.
// Examples:
// ```
// pkg/signatures/data/openai/api/sdk.yml
// pkg/signatures/data/google/gcp/vertexai.yml
// pkg/signatures/data/amazon/aws/bedrock.yml
// ```

//go:embed data
var signatureFiles embed.FS

type signatureFile struct {
	Version    string                  `yaml:"version"`
	Signatures []callgraphv1.Signature `yaml:"signatures"`
}

// LoadSignatures loads the signatures from the specified vendor, product, and service.
// It returns a slice of callgraph.Signature and an error if any occurs during the loading process.
//
// If a service is not specified, it will load all signatures for the given vendor and product.
// If a product is not specified, it will load all signatures for the given vendor.
// If a vendor is not specified, it will load all the signatures
func LoadSignatures(vendor string, product string, service string) ([]*callgraphv1.Signature, error) {
	isSingleSignatureFile := false

	subDirs := []string{"data", vendor}
	if product != "" {
		subDirs = append(subDirs, product)
		if service != "" {
			subDirs = append(subDirs, service+".yaml")
			isSingleSignatureFile = true
		}
	}

	signaturesPath := path.Join(subDirs...)

	log.Debugf("Reading signatures from: %s (%t)", signaturesPath, isSingleSignatureFile)

	if isSingleSignatureFile {
		return loadSignatureFile(signaturesPath)
	}

	// Walk through shortlisted files and parse signatures
	targetSignatures := []*callgraphv1.Signature{}

	err := fs.WalkDir(signatureFiles, signaturesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		signatures, err := loadSignatureFile(path)
		if err != nil {
			return fmt.Errorf("failed to load signature file %s: %v", path, err)
		}

		targetSignatures = append(targetSignatures, signatures...)
		return nil
	})
	if err != nil {
		return []*callgraphv1.Signature{}, fmt.Errorf("failed to walk through signature files: %w", err)
	}

	// Validate the loaded signatures
	validationErr := callgraph.ValidateSignatures(targetSignatures)
	if validationErr != nil {
		return []*callgraphv1.Signature{}, fmt.Errorf("invalid signatures: %w", validationErr)
	}

	// Ensure no duplicate signatures
	duplicationErr := checkDuplicateSignatures(targetSignatures)
	if duplicationErr != nil {
		return []*callgraphv1.Signature{}, fmt.Errorf("duplicate signatures found: %w", duplicationErr)
	}

	return targetSignatures, nil
}

// Wrapper of LoadSignatures to get all signatures conveniently
func LoadAllSignatures() ([]*callgraphv1.Signature, error) {
	return LoadSignatures("", "", "")
}

// parse signatures from a given yaml file
func loadSignatureFile(file string) ([]*callgraphv1.Signature, error) {
	signatureData, err := signatureFiles.ReadFile(file)
	if err != nil {
		log.Errorf("Failed to read signature file: %v", err)
		return []*callgraphv1.Signature{}, err
	}

	var parsedSignatureFile signatureFile
	err = yaml.Unmarshal(signatureData, &parsedSignatureFile)
	if err != nil {
		log.Errorf("Failed to parse signature YAML - %s: %v", file, err)
		return []*callgraphv1.Signature{}, err
	}

	parsedSignatures := make([]*callgraphv1.Signature, len(parsedSignatureFile.Signatures))
	for i := range parsedSignatureFile.Signatures {
		parsedSignatures[i] = &parsedSignatureFile.Signatures[i]
	}

	return parsedSignatures, nil
}

func checkDuplicateSignatures(signatures []*callgraphv1.Signature) error {
	signatureMap := make(map[string]bool)
	for _, signature := range signatures {
		if _, exists := signatureMap[signature.Id]; exists {
			return fmt.Errorf("duplicate signature - %s", signature.Id)
		}
		signatureMap[signature.Id] = true
	}
	return nil
}
