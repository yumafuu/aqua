package reader

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var errCircularImport = errors.New("circular import detected")

// appendToImportChain checks for circular import and appends the path to the chain.
// Returns the updated chain and an error if a circular import is detected.
func appendToImportChain(chain []string, configFilePath string) ([]string, error) {
	absPath, err := filepath.Abs(configFilePath)
	if err != nil {
		absPath = configFilePath
	}
	for _, ancestor := range chain {
		if ancestor == absPath {
			cycleChain := append(chain, absPath) //nolint:gocritic
			return nil, fmt.Errorf("%w: %s", errCircularImport, formatImportChain(cycleChain))
		}
	}
	return append(chain, absPath), nil
}

// formatImportChain formats the import chain as relative paths from the root config.
// Example: "aqua.yaml -> imports/a.yaml -> imports/b.yaml -> aqua.yaml"
func formatImportChain(chain []string) string {
	if len(chain) == 0 {
		return ""
	}

	// Use the directory of the first file as the base for relative paths
	rootDir := filepath.Dir(chain[0])

	names := make([]string, len(chain))
	for i, p := range chain {
		rel, err := filepath.Rel(rootDir, p)
		if err != nil {
			// Fallback to absolute path if relative path calculation fails
			names[i] = p
		} else {
			names[i] = rel
		}
	}
	return strings.Join(names, " -> ")
}
