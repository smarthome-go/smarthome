package types

//
// Tolerance.
// This describes how to deal with a failed registration.
// For instance, a `trigger` statement probably has no tolerance for failure as the script should fail if,
// let's say the MQTT subsystem is down and registration failed.
// On the other hand, a `trigger` annotation is expected to work if technically possible.
// Therefore, the dispatcher should retry the registration to make it available once possible.
//

type Tolerance uint8

const (
	NoTolerance Tolerance = iota
	ToleranceRetry
)
