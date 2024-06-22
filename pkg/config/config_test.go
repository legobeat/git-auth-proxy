package config

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func fsWithContent(content string) (afero.Fs, string, error) {
	path := "config.json"
	fs := afero.NewMemMapFs()
	file, err := fs.Create(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return nil, "", err
	}
	return fs, path, nil
}

const invalidJson = `
{
  "host": "dev.example.com",
	}}
}
`

func TestInvalidJson(t *testing.T) {
	fs, path, err := fsWithContent(invalidJson)
	require.NoError(t, err)
	_, err = LoadConfiguration(fs, path)
	require.Error(t, err)
}

const validGitHub = `
{
	"policies": [
		{
      "provider": "github",
			"github": {
        "token": "foobar"
      },
			"host": "github.com",
			"repositories": [
				{
					"owner": "example",
					"name": "gitops-deployment"
				}
			]
		}
	]
}
`

func TestValidGitHub(t *testing.T) {
	fs, path, err := fsWithContent(validGitHub)
	require.NoError(t, err)
	cfg, err := LoadConfiguration(fs, path)
	require.NoError(t, err)

	require.NotEmpty(t, cfg.Policies)
	require.Equal(t, "github", string(cfg.Policies[0].Provider))
	require.Equal(t, "foobar", cfg.Policies[0].GitHub.Token)
	require.Equal(t, "github.com", cfg.Policies[0].Host)
	require.Equal(t, "https", cfg.Policies[0].Scheme)
	require.NotEmpty(t, cfg.Policies[0].Repositories)
	require.Equal(t, "gitops-deployment", cfg.Policies[0].Repositories[0].Name)
}
