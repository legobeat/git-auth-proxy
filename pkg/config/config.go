package config

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
)

const (
	defaultScheme = "https"
)

type ProviderType string

const (
	GitHubProviderType = "github"
)

type Configuration struct {
	Organizations []*Organization `json:"organizations" validate:"required,dive"`
}

type Organization struct {
	Provider     ProviderType  `json:"provider" validate:"required,oneof='forgejo' 'github'"`
	GitHub       GitHub        `json:"github"`
	Host         string        `json:"host,omitempty" validate:"required,hostname"`
	Scheme       string        `json:"scheme,omitempty" validate:"required"`
	UserAuth     UserAuth      `json:"userAuth" validate:"required,dive"`
	Name         string        `json:"name" validate:"required"`
	Repositories []*Repository `json:"repositories" validate:"required,dive"`
}

type UserAuth struct {
	TokenHash string `json:"tokenHash"`
}

type GitHub struct {
	Token string `json:"token"`
}

type Repository struct {
	Owner string `json:"owner"`
	Name  string `json:"name" validate:"required"`
}

func setConfigurationDefaults(cfg *Configuration) *Configuration {
	for i, p := range cfg.Organizations {
		if p.Scheme == "" {
			cfg.Organizations[i].Scheme = defaultScheme
		}
	}
	return cfg
}

// LoadConfiguration parses and validates the configuration file at a given path.
func LoadConfiguration(fs afero.Fs, path string) (*Configuration, error) {
	b, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}

	cfg := &Configuration{}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	cfg = setConfigurationDefaults(cfg)

	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
