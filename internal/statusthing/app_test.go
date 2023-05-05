package statusthing

import (
	"fmt"
	"testing"

	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/storers"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type testCase struct {
		opts      []AppOption
		shouldErr bool
	}
	errOpt := func() AppOption {
		return func(ac *AppConfig) error { return fmt.Errorf("kaboom") }
	}
	testCases := map[string]testCase{
		"with-store": {
			opts:      []AppOption{WithStorer(&storers.UnimplementedStorer{})},
			shouldErr: false,
		},
		"with-provider": {
			opts:      []AppOption{WithProvider(&providers.UnimplementedProvider{})},
			shouldErr: false,
		},
		"with-error": {
			opts:      []AppOption{errOpt()},
			shouldErr: true,
		},
		"nostore-noprovider": {
			opts:      []AppOption{},
			shouldErr: true,
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			a, err := New(tc.opts...)
			if tc.shouldErr {
				require.Error(t, err, "should error")
				require.Nil(t, a, "should be nil")
			} else {
				require.NoError(t, err, "should not error")
				require.NotNil(t, a, "should not be nil")
			}
		})
	}
}
