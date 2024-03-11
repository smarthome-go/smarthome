package dispatcher

type RegisterInfoMetadata struct {
}

type RegisterInfo struct {
	ProgramID string
	Metadata  RegisterInfoMetadata
}

type Dispatcher interface {
	Register(RegisterInfo) error
}

//
// Manager interface
//

type DispatchManager interface {
	Dispatchers() []Dispatcher
}

//
// Manager implementation
//

// type DispatchManager struct {
// 	Dispatchers []Dispatcher
// }
