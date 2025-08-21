package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	brandPurpleBold = color.RGB(124, 58, 237).Add(color.Bold).SprintFunc() // 	#7C3AED Brand Purple
	whiteDim        = color.New(color.Faint).SprintFunc()
	whiteBold       = color.New(color.Bold).SprintFunc()
)

func GenerateXBOMBanner(version, commit string) string {
	var xBomASCIIText = `
▀▄▀ █▄▄ █▀█ █▀▄▀█   From SafeDep
█░█ █▄█ █▄█ █░▀░█` // It should end here no \n

	if len(commit) >= 6 {
		commit = commit[:6]
	}

	return fmt.Sprintf("%s   %s: %s %s: %s\n\n", brandPurpleBold(xBomASCIIText),
		whiteDim("version"), whiteBold(version),
		whiteDim("commit"), whiteBold(commit),
	)
}
