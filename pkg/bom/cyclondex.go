package bom

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	"github.com/safedep/dry/log"
	"github.com/safedep/dry/utils"
	"github.com/safedep/xbom/pkg/codeanalysis"
	"github.com/safedep/xbom/pkg/common"
)

type CycloneDXGeneratorConfig struct {
	Tool common.ToolMetadata

	// Path defines the output file path
	Path string

	// Application component name, this is the top-level component in the BOM
	ApplicationComponentName string

	// Unique identifier for this BOM confirming to UUID RFC 4122 standard
	// If empty, a new UUID will be generated
	SerialNumber string
}

type CycloneDXGenerator struct {
	config              CycloneDXGeneratorConfig
	bom                 *cdx.BOM
	toolComponent       cdx.Component
	rootComponentBomref string
	bomEcosystems       map[string]bool
}

var _ BomGenerator = (*CycloneDXGenerator)(nil)

var cdxUUIDRegexp = regexp.MustCompile(`^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func NewCycloneDXBomGenerator(config CycloneDXGeneratorConfig) (*CycloneDXGenerator, error) {
	bom := cdx.NewBOM()
	bom.SpecVersion = cdx.SpecVersion1_6

	// Set serial number if provided, otherwise generate a RFC 4122 UUID
	if utils.IsEmptyString(config.SerialNumber) {
		generatedSerialNumber, err := uuid.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate UUID for CycloneDX serial number: %v", err)
		}

		bom.SerialNumber = fmt.Sprintf("urn:uuid:%s", generatedSerialNumber.String())
	} else {
		if !cdxUUIDRegexp.MatchString(config.SerialNumber) {
			return nil, fmt.Errorf("serial number '%s' does not match RFC 4122 UUID format", config.SerialNumber)
		}

		bom.SerialNumber = config.SerialNumber
	}

	toolComponent := cdx.Component{
		Type: cdx.ComponentTypeApplication,
		Manufacturer: &cdx.OrganizationalEntity{
			Name: config.Tool.VendorName,
			URL:  utils.PtrTo([]string{config.Tool.VendorInformationURI}),
		},
		Group:      config.Tool.VendorName,
		Name:       config.Tool.Name,
		Version:    config.Tool.Version,
		PackageURL: config.Tool.Purl,
		BOMRef:     config.Tool.Purl,
	}

	rootComponentBomref := "root-application"
	bom.Metadata = &cdx.Metadata{
		// Define metadata about the main component (the root component which BOM describes)
		Component: &cdx.Component{
			BOMRef:     rootComponentBomref,
			Type:       cdx.ComponentTypeApplication,
			Name:       config.ApplicationComponentName,
			Components: utils.PtrTo([]cdx.Component{}),
		},
		Tools: &cdx.ToolsChoice{
			Components: utils.PtrTo([]cdx.Component{
				toolComponent,
			}),
		},
	}

	bom.Components = utils.PtrTo([]cdx.Component{})
	bom.Vulnerabilities = utils.PtrTo([]cdx.Vulnerability{})
	bom.Dependencies = utils.PtrTo([]cdx.Dependency{})
	bom.Services = utils.PtrTo([]cdx.Service{})

	return &CycloneDXGenerator{
		config:              config,
		bom:                 bom,
		toolComponent:       toolComponent,
		rootComponentBomref: rootComponentBomref,
		bomEcosystems:       map[string]bool{},
	}, nil
}

// RecordCodeAnalysisFindings implements BomGenerator.
func (c *CycloneDXGenerator) RecordCodeAnalysisFindings(findings *codeanalysis.CodeAnalysisFindings) error {
	for signatureId, signatureMatchResults := range findings.SignatureWiseMatchResults {
		if len(signatureMatchResults) == 0 {
			continue
		}
		signature := signatureMatchResults[0].MatchedSignature

		occurrences := &[]cdx.EvidenceOccurrence{}
		for _, signatureMatchResult := range signatureMatchResults {
			for _, condition := range signatureMatchResult.MatchedConditions {
				for _, evidence := range condition.Evidences {
					metadata, exists := evidence.Metadata()
					if exists {
						*occurrences = append(*occurrences, cdx.EvidenceOccurrence{
							Location:          signatureMatchResult.FilePath,
							Line:              utils.PtrTo(int(metadata.StartLine + 1)),
							Offset:            utils.PtrTo(int(metadata.StartColumn + 1)),
							AdditionalContext: evidence.Namespace,
						})
					}
				}
			}
		}

		component := cdx.Component{
			BOMRef:      signatureId,
			Name:        signature.Product + " - " + signature.Service,
			Type:        cdx.ComponentTypeLibrary,
			Description: signature.GetDescription(),
			Publisher:   signature.GetVendor(),
			Manufacturer: &cdx.OrganizationalEntity{
				Name:   signature.GetVendor(),
				BOMRef: signature.GetVendor(),
			},
			Evidence: &cdx.Evidence{
				Identity: utils.PtrTo([]cdx.EvidenceIdentity{
					{
						Methods: utils.PtrTo([]cdx.EvidenceIdentityMethod{
							{
								Technique:  cdx.EvidenceIdentityTechniqueSourceCodeAnalysis,
								Confidence: utils.PtrTo(float32(1.0)),
							},
						}),
						Tools: utils.PtrTo([]cdx.BOMReference{
							cdx.BOMReference(c.toolComponent.BOMRef),
						}),
					},
				}),
				Occurrences: occurrences,
			},
			Properties: &[]cdx.Property{},
		}

		*component.Properties = append(*component.Properties, c.getKnownTaggedProperties(signature.Tags)...)

		*c.bom.Components = append(*c.bom.Components, component)
	}
	return nil
}

func (c *CycloneDXGenerator) getKnownTaggedProperties(tags []string) []cdx.Property {
	knownTags := []string{
		"ai",
		"ml",
		"iaas",
		"paas",
		"saas",
	}

	properties := []cdx.Property{}
	for _, tag := range knownTags {
		if slices.Contains(tags, tag) {
			properties = append(properties, cdx.Property{
				Name:  tag,
				Value: "true",
			})
		}
	}

	return properties
}

func (r *CycloneDXGenerator) finaliseBom() {
	bomGenerationTime := time.Now().UTC()

	r.bom.Metadata.Timestamp = bomGenerationTime.Format(time.RFC3339)

	r.bom.Annotations = utils.PtrTo([]cdx.Annotation{
		{
			BOMRef: "metadata-annotations",
			Subjects: utils.PtrTo([]cdx.BOMReference{
				cdx.BOMReference(r.rootComponentBomref),
			}),
			Annotator: &cdx.Annotator{
				Component: &r.toolComponent,
			},
			Timestamp: bomGenerationTime.Format(time.RFC3339),
			Text:      fmt.Sprintf("This Software Bill-of-Materials (SBOM) document was created on %s with %s. The data was captured during the build lifecycle phase. The document describes '%s'. It has total %d components.", bomGenerationTime.Format("Monday, January 2, 2006"), r.config.Tool.Name, r.config.ApplicationComponentName, len(*r.bom.Components)),
		},
	})
}

func (r *CycloneDXGenerator) Finish() error {
	r.finaliseBom()

	log.Infof("Writing CycloneDX report to %s", r.config.Path)

	fd, err := os.Create(r.config.Path)
	if err != nil {
		return err
	}
	defer fd.Close()

	err = cdx.NewBOMEncoder(fd, cdx.BOMFileFormatJSON).
		SetPretty(true).
		Encode(r.bom)
	if err != nil {
		return err
	}

	return nil
}
