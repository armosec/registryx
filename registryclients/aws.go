package registryclients

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/registries/defaultregistry"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	dockerregistry "github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"strings"
)

const (
	registryURIFormat   = "%s.dkr.ecr.%s.amazonaws.com"
	registrySessionName = "AWSRegistryClientSession"
)

type AWSRegistryClient struct {
	Registry    *armotypes.AWSImageRegistry
	Options     *common.RegistryOptions
	registryURI string
	username    string
	password    string
}

func NewAWSRegistryClient(registry *armotypes.AWSImageRegistry, options *common.RegistryOptions) (*AWSRegistryClient, error) {
	if registry.AccessKeyID != "" && registry.SecretAccessKey != "" {
		return newAWSRegistryClientUsingCredentials(registry, registry.AccessKeyID, registry.SecretAccessKey, registry.RegistryRegion, options)
	}

	return newAWSRegistryClientUsingIAMRole(registry, registry.RoleARN, registry.RegistryRegion, options)
}

func newAWSRegistryClientUsingCredentials(registry *armotypes.AWSImageRegistry, accessKeyID, secretAccessKey, region string, options *common.RegistryOptions) (*AWSRegistryClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return newAWSRegistryClientUsingConfig(registry, cfg, options)
}

func newAWSRegistryClientUsingIAMRole(registry *armotypes.AWSImageRegistry, roleARN, region string, options *common.RegistryOptions) (*AWSRegistryClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	stsClient := sts.NewFromConfig(cfg)
	assumeRoleOutput, err := stsClient.AssumeRole(context.Background(), &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String(registrySessionName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	assumeRoleCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			aws.ToString(assumeRoleOutput.Credentials.AccessKeyId),
			aws.ToString(assumeRoleOutput.Credentials.SecretAccessKey),
			aws.ToString(assumeRoleOutput.Credentials.SessionToken),
		)),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config with assumed role credentials: %w", err)
	}

	return newAWSRegistryClientUsingConfig(registry, assumeRoleCfg, options)
}

func newAWSRegistryClientUsingConfig(registry *armotypes.AWSImageRegistry, cfg aws.Config, options *common.RegistryOptions) (*AWSRegistryClient, error) {
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	output, err := ecrClient.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization token: %w", err)
	}

	if len(output.AuthorizationData) == 0 {
		return nil, fmt.Errorf("no authorization data received")
	}

	authData := output.AuthorizationData[0]
	decodedToken, err := base64.StdEncoding.DecodeString(aws.ToString(authData.AuthorizationToken))
	if err != nil {
		return nil, fmt.Errorf("failed to decode authorization token: %w", err)
	}

	tokenParts := strings.SplitN(string(decodedToken), ":", 2)
	if len(tokenParts) != 2 {
		return nil, fmt.Errorf("invalid authorization token format")
	}

	username := tokenParts[0]
	password := tokenParts[1]

	return &AWSRegistryClient{
		Registry:    registry,
		registryURI: fmt.Sprintf(registryURIFormat, *identity.Account, cfg.Region),
		username:    username,
		password:    password,
		Options:     options,
	}, nil
}

func (a *AWSRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	registry, err := name.NewRegistry(a.registryURI)
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: a.username, Password: a.password}, &registry, a.Options)
	if err != nil {
		return nil, err
	}

	return getAllRepositories(ctx, iRegistry)
}

func (a *AWSRegistryClient) GetImagesToScan(_ context.Context) (map[string]string, error) {
	registry, err := name.NewRegistry(a.registryURI)
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: a.username, Password: a.password}, &registry, a.Options)
	if err != nil {
		return nil, err
	}
	images := make(map[string]string, len(a.Registry.Repositories))
	for _, repository := range a.Registry.Repositories {
		tag, err := getImageLatestTag(repository, iRegistry)
		if err != nil {
			return nil, err
		}
		if tag != "" {
			images[fmt.Sprintf("%s/%s", a.registryURI, repository)] = tag
		}
	}
	return images, nil
}

func (a *AWSRegistryClient) GetDockerAuth() (*dockerregistry.AuthConfig, error) {
	return &dockerregistry.AuthConfig{
		Username: a.username,
		Password: a.password,
	}, nil
}
