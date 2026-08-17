package setup

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	setupBootstrapTokenFile = ".setup-token"
	setupBootstrapTokenEnv  = "SETUP_BOOTSTRAP_TOKEN"
	setupBootstrapHeader    = "X-Setup-Token"
)

// GetBootstrapTokenPath returns the owner-readable file containing the
// one-time credential required by the HTTP setup wizard.
func GetBootstrapTokenPath() string {
	return filepath.Join(GetDataDir(), setupBootstrapTokenFile)
}

// LoadOrCreateBootstrapToken returns the setup credential without logging it.
// Operators may provide an explicit secret through SETUP_BOOTSTRAP_TOKEN; the
// default path creates a 256-bit token in a mode-0600 file.
func LoadOrCreateBootstrapToken() (string, error) {
	if configured := strings.TrimSpace(os.Getenv(setupBootstrapTokenEnv)); configured != "" {
		if len(configured) < 32 {
			return "", fmt.Errorf("%s must contain at least 32 characters", setupBootstrapTokenEnv)
		}
		return configured, nil
	}

	path := GetBootstrapTokenPath()
	if raw, err := readBootstrapTokenFile(path); err == nil {
		token := strings.TrimSpace(string(raw))
		if len(token) < 32 {
			return "", fmt.Errorf("setup bootstrap token file %s is invalid", path)
		}
		return token, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read setup bootstrap token: %w", err)
	}

	if err := os.MkdirAll(GetDataDir(), 0o750); err != nil {
		return "", fmt.Errorf("create setup data directory: %w", err)
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate setup bootstrap token: %w", err)
	}
	token := hex.EncodeToString(random)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return LoadOrCreateBootstrapToken()
	}
	if err != nil {
		return "", fmt.Errorf("create setup bootstrap token: %w", err)
	}
	if _, err := file.WriteString(token + "\n"); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("write setup bootstrap token: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close setup bootstrap token: %w", err)
	}
	return token, nil
}

func readBootstrapTokenFile(path string) ([]byte, error) {
	linkInfo, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if linkInfo.Mode()&os.ModeSymlink != 0 || !linkInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("setup bootstrap token path must be a regular file")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(linkInfo, openedInfo) {
		return nil, fmt.Errorf("setup bootstrap token file changed while opening")
	}
	if err := file.Chmod(0o600); err != nil {
		return nil, fmt.Errorf("restrict setup bootstrap token permissions: %w", err)
	}
	return io.ReadAll(io.LimitReader(file, 1024))
}

// GetSetupServerAddress keeps an uninitialized instance on loopback unless an
// operator explicitly opts into a remote setup listener with SETUP_HOST.
func GetSetupServerAddress() string {
	host := strings.TrimSpace(os.Getenv("SETUP_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}
	host = strings.Trim(host, "[]")
	port := strings.TrimSpace(os.Getenv("SETUP_PORT"))
	if port == "" {
		port = strings.TrimSpace(os.Getenv("SERVER_PORT"))
	}
	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort < 1 || parsedPort > 65535 {
		port = "8080"
	}
	return net.JoinHostPort(host, port)
}

func validBootstrapToken(expected, supplied string) bool {
	expected = strings.TrimSpace(expected)
	supplied = strings.TrimSpace(supplied)
	if expected == "" || len(expected) != len(supplied) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) == 1
}

func removeBootstrapToken() error {
	if strings.TrimSpace(os.Getenv(setupBootstrapTokenEnv)) != "" {
		return nil
	}
	err := os.Remove(GetBootstrapTokenPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
