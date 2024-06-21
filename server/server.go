package server

type Server interface {
	Init() error
	Start() error
	Shutdown() error
}
