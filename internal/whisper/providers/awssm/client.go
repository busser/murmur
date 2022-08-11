package awssm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/google/uuid"
)

type client struct {
	awsClient *secretsmanager.SecretsManager
}

// New returns a client that fetches secrets from AWS Secrets Manager.
func New() (*client, error) {
	session, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	c := secretsmanager.New(session)

	return &client{
		awsClient: c,
	}, nil
}

func (c *client) Resolve(ctx context.Context, ref string) (string, error) {
	secretID, versionID, versionStage, err := parseRef(ref)
	if err != nil {
		return "", fmt.Errorf("invalid reference: %w", err)
	}

	req := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}
	if versionID != "" {
		req.VersionId = aws.String(versionID)
	}
	if versionStage != "" {
		req.VersionStage = aws.String(versionStage)
	}

	resp, err := c.awsClient.GetSecretValueWithContext(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %q version ID: %q version stage: %q): %w", secretID, versionID, versionStage, err)
	}

	if resp.SecretString != nil {
		return *resp.SecretString, nil
	}
	return string(resp.SecretBinary), nil
}

func (c *client) Close() error {
	// The client does not need to close its underlying AWS client.
	// ?(busser): are we sure about this? do any connections need to be closed?
	return nil
}

func parseRef(ref string) (secretID, versionID, versionStage string, err error) {
	refParts := strings.SplitN(ref, "#", 2)
	if len(refParts) < 1 {
		return "", "", "", errors.New("invalid syntax")
	}
	secretID = refParts[0]

	if len(refParts) < 2 {
		return secretID, "", "AWSCURRENT", nil
	}

	rawVersion := refParts[1]
	switch {
	case rawVersion == "":
		return secretID, "", "AWSCURRENT", nil
	case isUUID(rawVersion):
		return secretID, rawVersion, "", nil
	default:
		return secretID, "", rawVersion, nil
	}
}

func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
