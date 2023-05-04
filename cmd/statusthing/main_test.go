package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromEnv(t *testing.T) {
	envVars := map[string]string{
		basePathEnvKey:    t.Name() + "basepath",
		addrEnvKey:        t.Name() + "addr",
		apiKeyEnvKey:      t.Name() + "apikey",
		dbFileNameEnvKey:  t.Name() + "dbfile",
		debugEnvKey:       t.Name() + "debug",
		enableDashEnvKey:  t.Name() + "dash",
		"NGROK_AUTHTOKEN": t.Name() + "ngrok_token",
		"NGROK_ENDPOINT":  t.Name() + "ngrok_endpoint",
	}
	defer func() {
		for k := range envVars {
			if err := os.Unsetenv(k); err != nil {
				t.Logf("unable to unset env var %s", k)
			}
		}
	}()

	for k, v := range envVars {
		err := os.Setenv(k, v)
		require.NoError(t, err, "env vars should get set")
	}
	cfg, err := configFromEnv()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, envVars[basePathEnvKey], cfg.basepath)
	require.Equal(t, envVars[addrEnvKey], cfg.addr)
	require.Equal(t, envVars[apiKeyEnvKey], cfg.apikey)
	require.Equal(t, envVars[dbFileNameEnvKey], cfg.dbfile)
	require.True(t, cfg.enableNgrok)
	require.Equal(t, envVars["NGROK_ENDPOINT"], cfg.ngrokEndpointName)
	require.True(t, cfg.debug)
	require.True(t, cfg.enableDash)
}
