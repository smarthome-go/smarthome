package hardware

type Node struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Token string `json:"token"`
}

type HardwareComfig struct {
	Nodes []Node `json:"nodes"`
}