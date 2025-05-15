package command

import (
	"log"
)

func FailOnError(stage string, err error) {
	if err != nil {
		log.Fatalf("Execution failed [%s] Error: %v", stage, err)
	}
}
