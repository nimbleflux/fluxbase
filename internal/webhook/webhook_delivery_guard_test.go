package webhook

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSSRFGuardControl verifies the dial-time control callback rejects
// private, loopback, and link-local destinations.
func TestSSRFGuardControl(t *testing.T) {
	blocked := []string{
		"127.0.0.1:8080",
		"10.1.2.3:443",
		"172.16.0.9:443",
		"192.168.1.1:80",
		"169.254.169.254:80",
		"[::1]:8080",
	}
	for _, addr := range blocked {
		err := ssrfGuardControl("tcp", addr, nil)
		require.Error(t, err, "expected %s to be blocked", addr)
		assert.Contains(t, err.Error(), "not allowed")
	}

	// Public addresses are allowed through the guard itself.
	assert.NoError(t, ssrfGuardControl("tcp", "93.184.216.34:443", nil))

	// Non-tcp networks (e.g. unix sockets) pass through unvalidated.
	assert.NoError(t, ssrfGuardControl("unix", "/tmp/some.sock", nil))
}

// TestNewSSRFGuardDialContext verifies the dial-time guard actually aborts
// connections to private addresses, closing the DNS-rebinding TOCTOU window.
func TestNewSSRFGuardDialContext(t *testing.T) {
	t.Run("private IP dial is blocked", func(t *testing.T) {
		dial := newSSRFGuardDialContext(false)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := dial(ctx, "tcp", "127.0.0.1:1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("allowPrivate disables the guard", func(t *testing.T) {
		dial := newSSRFGuardDialContext(true)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// The guard is off: the dial proceeds and fails with a normal dial
		// error (refused), not the SSRF rejection.
		_, err := dial(ctx, "tcp", "127.0.0.1:1")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "not allowed")
	})
}

// TestIsPrivateIP_LinkLocal checks the shared blocklist helper directly so the
// validation used both at URL validation and dial time stays in sync.
func TestIsPrivateIP_LinkLocal(t *testing.T) {
	assert.True(t, isPrivateIP(net.ParseIP("169.254.169.254")))
	assert.True(t, isPrivateIP(net.ParseIP("fe80::1")))
	assert.False(t, isPrivateIP(net.ParseIP("8.8.8.8")))
	assert.False(t, isPrivateIP(nil))
}

// Compile-time assertion that the Control callback matches net.Dialer's
// expected signature.
var _ func(string, string, syscall.RawConn) error = ssrfGuardControl
