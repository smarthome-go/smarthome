package dispatcher

type RegisterInfoMetadata struct {
}

type RegisterInfo struct {
	ProgramId string
	Metadata  RegisterInfoMetadata
}

type Dispatcher interface {
	Register(RegisterInfo) error
}
