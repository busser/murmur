package scwsm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	scwsecret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type client struct {
	scwClient *scw.Client
}

// New returns a client that fetches secrets from Google Secret Manager.
func New() (*client, error) {
	profile := loadDefaultProfile()

	log.Printf("Using Scaleway profile: %#v", profile)

	c, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("failed to setup client: %w", err)
	}

	return &client{
		scwClient: c,
	}, nil
}

func (c *client) Resolve(ctx context.Context, ref string) (string, error) {
	region, id, name, revision, err := parseRef(ref)
	if err != nil {
		return "", fmt.Errorf("invalid reference: %w", err)
	}

	if id != "" {
		return c.resolveByID(ctx, region, id, revision)
	}
	return c.resolveByName(ctx, region, name, revision)
}

func loadDefaultProfile() *scw.Profile {
	return scw.MergeProfiles(loadConfigProfile(), scw.LoadEnvProfile())
}

func loadConfigProfile() *scw.Profile {
	config, err := scw.LoadConfig()
	if err != nil {
		return &scw.Profile{}
	}

	profile, err := config.GetActiveProfile()
	if err != nil {
		return &scw.Profile{}
	}

	return profile
}

func (c *client) resolveByID(ctx context.Context, region scw.Region, id, revision string) (string, error) {
	req := &scwsecret.AccessSecretVersionRequest{
		Region:   region,
		SecretID: id,
		Revision: revision,
	}
	resp, err := scwsecret.NewAPI(c.scwClient).AccessSecretVersion(req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret (region: %q, id: %q, revision: %q): %w", region, id, revision, err)
	}

	return string(resp.Data), nil
}

func (c *client) resolveByName(ctx context.Context, region scw.Region, name, revision string) (string, error) {
	req := &scwsecret.AccessSecretVersionByNameRequest{
		Region:     region,
		SecretName: name,
		Revision:   revision,
	}
	resp, err := scwsecret.NewAPI(c.scwClient).AccessSecretVersionByName(req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret (region: %q, name: %q, revision: %q): %w", region, name, revision, err)
	}

	return string(resp.Data), nil
}

func (c *client) Close() error {
	// No need to close the client.
	return nil
}

const (
	// If the ref contains no region, then we want to use the default region.
	// We delegate this to the Scaleway SDK, which will use the default region if
	// the region is empty.
	defaultRegion scw.Region = ""

	defaultRevision = "latest"
)

func parseRef(ref string) (region scw.Region, id, name, revision string, err error) {
	refParts := strings.SplitN(ref, "#", 2)
	if len(refParts) < 1 {
		return "", "", "", "", errors.New("invalid syntax")
	}

	revision = defaultRevision
	if len(refParts) == 2 {
		revision = refParts[1]
	}

	fullname := refParts[0]
	fullnameParts := strings.SplitN(fullname, "/", 2)
	if len(fullnameParts) < 1 {
		return "", "", "", "", errors.New("invalid syntax")
	}

	region = defaultRegion
	idOrName := fullnameParts[0]
	if len(fullnameParts) == 2 {
		region = scw.Region(fullnameParts[0])
		idOrName = fullnameParts[1]
	}

	id, name = extractIDAndName(idOrName)

	return region, id, name, revision, nil
}

func extractIDAndName(idOrName string) (id, name string) {
	if isUUID(idOrName) {
		return idOrName, ""
	}
	return "", idOrName
}

func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
