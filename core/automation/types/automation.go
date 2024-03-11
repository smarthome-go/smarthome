package types

import (
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type AutomationManager interface {
	InitAutomations(config database.ServerConfig) error
	RunAllAutomationsWithTrigger(username string, trigger database.AutomationTrigger, context types.AutomationContext)
}
