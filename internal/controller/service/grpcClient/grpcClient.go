package grpcclient

import "github.com/Asliddin3/image-servis/config"

type Clients interface {
}

type ServiceManager struct {
	Config config.Config
}

func New(c config.Config) (Clients, error) {
	return ServiceManager{}, nil
}
