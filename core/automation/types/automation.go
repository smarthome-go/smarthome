package types

import (
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type AutomationManager interface {
	RunAllAutomationsWithTrigger(username string, trigger database.AutomationTrigger, context types.ExecutionContextAutomation)
}
