package setup

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoadOrCreateBootstrapTokenPersistsOwnerOnlySecret(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DATA_DIR", dir)
	t.Setenv(setupBootstrapTokenEnv, "")

	first, err := LoadOrCreateBootstrapToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadOrCreateBootstrapToken()
	if err != nil {
		t.Fatal(err)
	}
	if first != second || len(first) != 64 {
		t.Fatalf("bootstrap token was not stably persisted: first=%d second=%d", len(first), len(second))
	}
	info, err := os.Stat(GetBootstrapTokenPath())
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("bootstrap token mode = %o, want 600", got)
	}
	if !validBootstrapToken(first, second) || validBootstrapToken(first, second+"x") {
		t.Fatal("constant-time bootstrap token validation returned an unexpected result")
	}
}

func TestLoadOrCreateBootstrapTokenRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DATA_DIR", dir)
	t.Setenv(setupBootstrapTokenEnv, "")
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("0123456789abcdef0123456789abcdef\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, GetBootstrapTokenPath()); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	if _, err := LoadOrCreateBootstrapToken(); err == nil {
		t.Fatal("expected a symlink token path to be rejected")
	}
}

func TestLoadOrCreateBootstrapTokenRejectsShortEnvironmentSecret(t *testing.T) {
	t.Setenv(setupBootstrapTokenEnv, "short")
	if _, err := LoadOrCreateBootstrapToken(); err == nil {
		t.Fatal("expected a short environment token to be rejected")
	}
}

func TestGetSetupServerAddressIsLoopbackByDefault(t *testing.T) {
	t.Setenv("SETUP_HOST", "")
	t.Setenv("SETUP_PORT", "")
	t.Setenv("SERVER_PORT", "")
	if got := GetSetupServerAddress(); got != "127.0.0.1:8080" {
		t.Fatalf("setup address = %q, want loopback default", got)
	}
	t.Setenv("SETUP_HOST", "0.0.0.0")
	t.Setenv("SETUP_PORT", "9090")
	if got := GetSetupServerAddress(); got != "0.0.0.0:9090" {
		t.Fatalf("explicit setup address = %q", got)
	}
}

func TestSetupGuardRequiresBootstrapHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/install", setupGuard("0123456789abcdef0123456789abcdef"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/install", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want 401", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/install", nil)
	request.Header.Set(setupBootstrapHeader, "0123456789abcdef0123456789abcdef")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("valid token status = %d, want 204", response.Code)
	}
}
