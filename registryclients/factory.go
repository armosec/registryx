package registryclients

import (
	"fmt"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/interfaces"
)

func GetRegistryClient(registry armotypes.ContainerImageRegistry) (interfaces.RegistryClient, error) {
	provider := registry.GetBase().Provider
	switch provider {
	case armotypes.Quay:
		if quayRegistry, ok := registry.(*armotypes.QuayImageRegistry); ok {
			return &QuayRegistryClient{Registry: quayRegistry}, nil
		} else {
			return nil, fmt.Errorf("failed to convert registry to QuayImageRegistry type")
		}
	case armotypes.Harbor:
		if harborRegistry, ok := registry.(*armotypes.HarborImageRegistry); ok {
			return &HarborRegistryClient{Registry: harborRegistry}, nil
		} else {
			return nil, fmt.Errorf("failed to convert registry to HarborImageRegistry type")
		}
	}
	return nil, fmt.Errorf("unsupported provider %s", provider)
}
