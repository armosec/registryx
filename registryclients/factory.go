package registryclients

import (
	"fmt"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
)

func GetRegistryClient(registry armotypes.ContainerImageRegistry, registryOptions *common.RegistryOptions) (interfaces.RegistryClient, error) {
	provider := registry.GetBase().Provider
	switch provider {
	case armotypes.Quay:
		if quayRegistry, ok := registry.(*armotypes.QuayImageRegistry); ok {
			return &QuayRegistryClient{Registry: quayRegistry, Options: registryOptions}, nil
		} else {
			return nil, fmt.Errorf("failed to convert registry to QuayImageRegistry type")
		}
	case armotypes.Harbor:
		if harborRegistry, ok := registry.(*armotypes.HarborImageRegistry); ok {
			return &HarborRegistryClient{Registry: harborRegistry, Options: registryOptions}, nil
		} else {
			return nil, fmt.Errorf("failed to convert registry to HarborImageRegistry type")
		}
	case armotypes.Nexus:
		if nexusRegistry, ok := registry.(*armotypes.NexusImageRegistry); ok {
			return &NexusRegistryClient{Registry: nexusRegistry, Options: registryOptions}, nil
		} else {
			return nil, fmt.Errorf("failed to convert registry to NexusImageRegistry type")
		}
	case armotypes.Google:
		if googleRegistry, ok := registry.(*armotypes.GoogleImageRegistry); ok {
			return NewGoogleArtifactRegistryClient(googleRegistry, registryOptions)
		} else {
			return nil, fmt.Errorf("failed to convert registry to GoogleImageRegistry type")
		}
	case armotypes.Azure:
		if azureRegistry, ok := registry.(*armotypes.AzureImageRegistry); ok {
			return &AzureRegistryClient{Registry: azureRegistry, Options: registryOptions}, nil
		} else {
			return nil, fmt.Errorf("failed to convert registry to AzureImageRegistry type")
		}
	case armotypes.AWS:
		if awsRegistry, ok := registry.(*armotypes.AWSImageRegistry); ok {
			return NewAWSRegistryClient(awsRegistry, registryOptions)
		} else {
			return nil, fmt.Errorf("failed to convert registry to AWSImageRegistry type")
		}
	}
	return nil, fmt.Errorf("unsupported provider %s", provider)
}
