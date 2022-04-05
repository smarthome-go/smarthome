package automation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lnquy/cron"
)

// This file contains cron utils functions, mostly for parsing cron-expressions

// Generates a cron-expression based on the hour, minute, and  days on which the automation should run
func GenerateCronExpression(hour uint8, minute uint8, days []uint8) (string, error) {
	output := [5]string{"", "", "*", "*", ""}
	output[0] = fmt.Sprintf("%d", minute)
	output[1] = fmt.Sprintf("%d", hour)
	if len(days) > 7 {
		log.Error("The maximum allowed amount of days for generating a cron-expression is 7")
		return "", fmt.Errorf("amount of days should not be greater than 7")
	}
	if len(days) == 7 {
		// Set the days to '*' when all days are included in the slice, does not check for duplicate days
		// Duplicate days should be checked by the API-layer
		output[4] = "*"
		return strings.Join(output[:], " "), nil
	}
	for index, day := range days {
		output[4] += fmt.Sprintf("%d", day)
		if index < len(days)-1 {
			output[4] += ","
		}
	}
	return strings.Join(output[:], " "), nil
}

// Returns a slice which contains the days on which a given cron-expression will run
// Used for generating a new cron-expression when the timing function is set to either `sunrise` or `sunset`
func GetDaysFromCronExpression(expr string) ([]uint8, error) {
	if !IsValidCronExpression(expr) {
		return nil, errors.New("cannot get days from cron-expression: invalid cron-expression supplied")
	}
	days := make([]uint8, 0)
	exprSlice := strings.Split(expr, " ")
	if len(exprSlice) != 5 {
		return nil, errors.New("cannot get days from cron-expression: invalid cron-expression supplied: no definite days")
	}
	if exprSlice[4] == "*" {
		// All days are selected for execution
		return []uint8{0, 1, 2, 3, 4, 5, 6}, nil
	}
	daysTemp := strings.Split(exprSlice[4], ",") // The value at index 4 contains the days separated by a comma
	for _, day := range daysTemp {
		dayInt, err := strconv.Atoi(day)
		if err != nil {
			return nil, errors.New("cannot get days from cron-expression: invalid day in cron-expression: day is not a number")
		}
		days = append(days, uint8(dayInt))
	}
	return days, nil
}

// Generates a human-readable representation for a given (valid) cron-expression
func generateHumanReadableCronExpression(expr string) (string, error) {
	descriptor, err := cron.NewDescriptor()
	if err != nil {
		log.Error("Failed to parse cron-expression into human readable format: ", err.Error())
		return "", err
	}
	output, err := descriptor.ToDescription(expr, cron.Locale_en)
	if err != nil {
		log.Error("Failed to parse cron-expression into human readable format: ", err.Error())
		return "", err
	}
	return output, nil
}

// Validates a given cron-expression, returns false if the given cron-expression is invalid
func IsValidCronExpression(expr string) bool {
	descriptor, err := cron.NewDescriptor()
	if err != nil {
		return false
	}
	if _, err = descriptor.ToDescription(expr, cron.Locale_en); err != nil {
		return false
	}
	return true
}
