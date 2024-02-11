package homescript

import (
	"fmt"
	"strconv"
	"strings"
)

type SemanticVersion struct {
	Major uint64 `json:"major"`
	Minor uint64 `json:"minor"`
	Patch uint64 `json:"patch"`
}

const semVerSegments = 3

func ParseDriverVersion(source string) (SemanticVersion, error) {
	delimiter := "."

	split := strings.Split(source, delimiter)
	if len(split) != semVerSegments {
		return SemanticVersion{}, fmt.Errorf("Expected exactly 3 version components, got %d", len(split))
	}

	parsedValues := make([]uint64, semVerSegments)
	for idx, element := range split {
		parsed, err := strconv.ParseUint(element, 10, 64)
		if err != nil {
			return SemanticVersion{}, err
		}

		parsedValues[idx] = parsed
	}

	//nolint:exhaustruct
	return SemanticVersion{}, nil
}
