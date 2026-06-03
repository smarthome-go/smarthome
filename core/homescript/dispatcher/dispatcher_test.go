package dispatcher

import (
	"sync"
	"testing"

	dispatcherTypes "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
)

func TestMqttTopicMatchesFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		topic  string
		want   bool
	}{
		{
			name:   "exact",
			filter: "zigbee2mqtt/wall_switch",
			topic:  "zigbee2mqtt/wall_switch",
			want:   true,
		},
		{
			name:   "multi level wildcard matches child",
			filter: "zigbee2mqtt/#",
			topic:  "zigbee2mqtt/wall_switch/action",
			want:   true,
		},
		{
			name:   "multi level wildcard matches parent",
			filter: "zigbee2mqtt/#",
			topic:  "zigbee2mqtt",
			want:   true,
		},
		{
			name:   "single level wildcard",
			filter: "zigbee2mqtt/+/action",
			topic:  "zigbee2mqtt/wall_switch/action",
			want:   true,
		},
		{
			name:   "single level wildcard does not span levels",
			filter: "zigbee2mqtt/+/action",
			topic:  "zigbee2mqtt/room/wall_switch/action",
			want:   false,
		},
		{
			name:   "leading slash is significant",
			filter: "/zigbee2mqtt/#",
			topic:  "zigbee2mqtt/wall_switch/action",
			want:   false,
		},
		{
			name:   "shared subscription unwraps filter",
			filter: "$share/smarthome/zigbee2mqtt/#",
			topic:  "zigbee2mqtt/wall_switch",
			want:   true,
		},
		{
			name:   "queue subscription unwraps filter",
			filter: "$queue/zigbee2mqtt/#",
			topic:  "zigbee2mqtt/wall_switch",
			want:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mqttTopicMatchesFilter(test.filter, test.topic); got != test.want {
				t.Fatalf("mqttTopicMatchesFilter(%q, %q) = %v, want %v", test.filter, test.topic, got, test.want)
			}
		})
	}
}

func TestMatchingMqttRegistrationsUsesSubscriptionFilters(t *testing.T) {
	const (
		exactID    dispatcherTypes.RegistrationID = 1
		wildID     dispatcherTypes.RegistrationID = 2
		plusID     dispatcherTypes.RegistrationID = 3
		noMatchID  dispatcherTypes.RegistrationID = 4
		overlapID  dispatcherTypes.RegistrationID = 5
		missingID  dispatcherTypes.RegistrationID = 6
		duplicateA dispatcherTypes.RegistrationID = 7
	)

	instance := InstanceT{
		DoneRegistrations: dispatcherTypes.Registrations{
			Lock: sync.RWMutex{},
			Set: map[dispatcherTypes.RegistrationID]dispatcherTypes.RegisterInfo{
				exactID: {
					ProgramID: "exact",
					Function:  &dispatcherTypes.CalledFunction{},
				},
				wildID: {
					ProgramID: "wild",
					Function:  &dispatcherTypes.CalledFunction{},
				},
				plusID: {
					ProgramID: "plus",
					Function:  &dispatcherTypes.CalledFunction{},
				},
				noMatchID: {
					ProgramID: "no-match",
					Function:  &dispatcherTypes.CalledFunction{},
				},
				overlapID: {
					ProgramID: "overlap",
					Function:  &dispatcherTypes.CalledFunction{},
				},
				missingID: {
					ProgramID: "missing",
					Function:  &dispatcherTypes.CalledFunction{},
				},
				duplicateA: {
					ProgramID: "duplicate",
					Function:  &dispatcherTypes.CalledFunction{},
				},
			},
			MqttRegistrations: map[string][]dispatcherTypes.RegistrationID{
				"zigbee2mqtt/wall_switch/action": {exactID, overlapID, duplicateA},
				"zigbee2mqtt/#":                  {wildID, overlapID},
				"zigbee2mqtt/+/action":           {plusID, duplicateA},
				"other/#":                        {noMatchID, missingID},
			},
		},
	}

	matches := instance.matchingMqttRegistrations("zigbee2mqtt/wall_switch/action")
	got := make(map[string]int)
	for _, match := range matches {
		got[match.ProgramID]++
	}

	for _, want := range []string{"exact", "wild", "plus", "overlap", "duplicate"} {
		if got[want] != 1 {
			t.Fatalf("program %q matched %d times, want exactly once; all matches: %#v", want, got[want], got)
		}
	}

	if got["no-match"] != 0 || got["missing"] != 0 {
		t.Fatalf("non-matching registrations were returned: %#v", got)
	}

	if len(matches) != 5 {
		t.Fatalf("got %d matches, want 5: %#v", len(matches), got)
	}
}
