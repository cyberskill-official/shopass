package secrets

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type awsSMProvider struct {
	client *secretsmanager.Client
}

// NewAWSSMProvider creates a provider using AWS Secrets Manager.
func NewAWSSMProvider(client *secretsmanager.Client) SecretProvider {
	return &awsSMProvider{client: client}
}

func (p *awsSMProvider) Get(ctx context.Context, path string) (Secret, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(path),
		VersionStage: aws.String("AWSCURRENT"),
	}
	out, err := p.client.GetSecretValue(ctx, input)
	if err != nil {
		return Secret{}, err
	}

	val := ""
	if out.SecretString != nil {
		val = *out.SecretString
	} else if out.SecretBinary != nil {
		val = string(out.SecretBinary)
	}

	ver := "unknown"
	if out.VersionId != nil {
		ver = *out.VersionId
	}

	return Secret{value: val, Version: ver}, nil
}
