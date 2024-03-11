package types

type HomescriptInitiator uint8

const (
	InitiatorAutomation         HomescriptInitiator = iota // triggered by a normal automation
	InitiatorAutomationOnNotify                            // triggered by an automation which runs on every notification
	InitiatorSchedule                                      // triggered by a schedule
	InitiatorExec                                          // triggered by a call to `exec`
	InitiatorInternal                                      // triggered internally
	InitiatorAPI                                           // triggered through the API
	InitiatorWidget                                        // triggered through a widget
)
