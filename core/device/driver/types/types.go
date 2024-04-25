package types

type DriverInvocationIDs struct {
	DeviceID *string
	VendorID string
	ModelID  string
}

//
// Driver manager.
//

type DriverManager interface {
}
