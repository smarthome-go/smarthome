package driver

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
const delimiter = "."

func ParseDriverVersion(source string) (SemanticVersion, error) {
	split := strings.Split(source, delimiter)
	if len(split) != semVerSegments {
		return SemanticVersion{}, fmt.Errorf("expected exactly 3 version components, got %d", len(split))
	}

	parsedValues := make([]uint64, semVerSegments)
	for idx, element := range split {
		parsed, err := strconv.ParseUint(element, 10, 64)
		if err != nil {
			return SemanticVersion{}, err
		}

		parsedValues[idx] = parsed
	}

	return SemanticVersion{
		Major: parsedValues[0],
		Minor: parsedValues[1],
		Patch: parsedValues[2],
	}, nil
}

func (version SemanticVersion) String() string {
	output := make([]string, semVerSegments)
	output[0] = fmt.Sprint(version.Major)
	output[1] = fmt.Sprint(version.Minor)
	output[2] = fmt.Sprint(version.Patch)
	return strings.Join(output, delimiter)
}
