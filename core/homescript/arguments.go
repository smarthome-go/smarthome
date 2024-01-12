package homescript

import (
	"github.com/smarthome-go/smarthome/core/database"
)

type HomescriptWithArguments struct {
	Data      database.Homescript      `json:"data"`
	Arguments []database.HomescriptArg `json:"arguments"`
}

// Returns a slice containing the user's Homescripts
// Each Homescript also contains its arguments as a slice
func ListPersonalHomescriptWithArgs(username string) ([]HomescriptWithArguments, error) {
	outputData := make([]HomescriptWithArguments, 0)

	// Retrieve Homescripts from the database, including virtual ones (drivers)
	homescripts, err := ListPersonal(username)
	if err != nil {
		return nil, err
	}

	// Retrieve arguments from the database
	arguments, err := database.ListAllHomescriptArgsOfUser(username)
	if err != nil {
		return nil, err
	}

	// Arrange the arguments in a map in order to decrease time complexity
	argBelongsTo := make(map[string][]database.HomescriptArg)
	for _, arg := range arguments {
		_, ok := argBelongsTo[arg.Data.HomescriptId]
		if !ok {
			// If no arguments were appended, create a slice and append the first argument
			argBelongsTo[arg.Data.HomescriptId] = make([]database.HomescriptArg, 0)
			argBelongsTo[arg.Data.HomescriptId] = append(argBelongsTo[arg.Data.HomescriptId], arg)
		} else {
			argBelongsTo[arg.Data.HomescriptId] = append(argBelongsTo[arg.Data.HomescriptId], arg)
		}
	}

	// Append the final Homescripts with their args to the output
	for _, hms := range homescripts {
		if len(argBelongsTo[hms.Data.Id]) > 0 {
			outputData = append(outputData, HomescriptWithArguments{
				Data:      hms,
				Arguments: argBelongsTo[hms.Data.Id],
			})
		} else {
			outputData = append(outputData, HomescriptWithArguments{
				Data:      hms,
				Arguments: make([]database.HomescriptArg, 0),
			})
		}
	}
	return outputData, nil
}
