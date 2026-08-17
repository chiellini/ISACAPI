package setup

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestDecideAdminBootstrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		totalUsers             int64
		configuredAdminMatches int64
		shouldCreate           bool
		canProceed             bool
		reason                 string
	}{
		{
			name:         "empty database should create admin",
			totalUsers:   0,
			shouldCreate: true,
			canProceed:   true,
			reason:       adminBootstrapReasonEmptyDatabase,
		},
		{
			name:                   "configured active admin exists should skip",
			totalUsers:             10,
			configuredAdminMatches: 1,
			canProceed:             true,
			reason:                 adminBootstrapReasonAdminExists,
		},
		{
			name:       "users exist without configured active admin should fail",
			totalUsers: 5,
			canProceed: false,
			reason:     adminBootstrapReasonConfiguredAdminMismatch,
		},
		{
			name:                   "duplicate configured admin identities should fail closed",
			totalUsers:             2,
			configuredAdminMatches: 2,
			canProceed:             false,
			reason:                 adminBootstrapReasonConfiguredAdminMismatch,
		},
		{
			name:                   "inconsistent empty database snapshot should fail closed",
			totalUsers:             0,
			configuredAdminMatches: 1,
			canProceed:             false,
			reason:                 adminBootstrapReasonConfiguredAdminMismatch,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := decideAdminBootstrap(tc.totalUsers, tc.configuredAdminMatches)
			if got.shouldCreate != tc.shouldCreate {
				t.Fatalf("shouldCreate=%v, want %v", got.shouldCreate, tc.shouldCreate)
			}
			if got.canProceed != tc.canProceed {
				t.Fatalf("canProceed=%v, want %v", got.canProceed, tc.canProceed)
			}
			if got.reason != tc.reason {
				t.Fatalf("reason=%q, want %q", got.reason, tc.reason)
			}
		})
	}
}

func TestCreateAdminUserWithDBRequiresConfiguredActiveAdminForExistingDatabase(t *testing.T) {
	tests := []struct {
		name                 string
		totalUsers           int64
		matchingActiveAdmins int64
		wantErr              bool
		wantCreated          bool
		wantReason           string
		expectInsert         bool
	}{
		{
			name:                 "matching configured admin permits recovery",
			totalUsers:           4,
			matchingActiveAdmins: 1,
			wantReason:           adminBootstrapReasonAdminExists,
		},
		{
			name:       "different administrator email fails closed",
			totalUsers: 4,
			wantErr:    true,
			wantReason: adminBootstrapReasonConfiguredAdminMismatch,
		},
		{
			name:       "ordinary users without administrator fail closed",
			totalUsers: 2,
			wantErr:    true,
			wantReason: adminBootstrapReasonConfiguredAdminMismatch,
		},
		{
			name:       "inactive or soft-deleted configured administrator fails closed",
			totalUsers: 1,
			wantErr:    true,
			wantReason: adminBootstrapReasonConfiguredAdminMismatch,
		},
		{
			name:         "empty database creates configured administrator",
			wantCreated:  true,
			wantReason:   adminBootstrapReasonEmptyDatabase,
			expectInsert: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RUN_MODE", "standard")
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })

			mock.ExpectQuery(regexp.QuoteMeta(configuredAdminStateQuery)).
				WithArgs("owner@example.com", service.RoleAdmin, service.RoleAdminProvider, service.StatusActive).
				WillReturnRows(sqlmock.NewRows([]string{"total_users", "configured_admin_matches"}).
					AddRow(tc.totalUsers, tc.matchingActiveAdmins))
			if tc.expectInsert {
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (email, password_hash, role, balance, concurrency, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)).
					WithArgs("Owner@example.com", sqlmock.AnyArg(), service.RoleAdmin, float64(0), defaultUserConcurrency, service.StatusActive, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			created, reason, err := createAdminUserWithDB(context.Background(), db, &SetupConfig{
				Admin: AdminConfig{Email: " Owner@example.com ", Password: "strong-password"},
			})
			if (err != nil) != tc.wantErr {
				t.Fatalf("error=%v, wantErr=%v", err, tc.wantErr)
			}
			if created != tc.wantCreated {
				t.Fatalf("created=%v, want %v", created, tc.wantCreated)
			}
			if reason != tc.wantReason {
				t.Fatalf("reason=%q, want %q", reason, tc.wantReason)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestConfiguredAdminStateQueryRequiresMatchingActiveNonDeletedAdmin(t *testing.T) {
	for _, requiredClause := range []string{
		"LOWER(TRIM(email)) = $1",
		"role IN ($2, $3)",
		"status = $4",
		"deleted_at IS NULL",
	} {
		if !strings.Contains(configuredAdminStateQuery, requiredClause) {
			t.Fatalf("configured admin query missing %q", requiredClause)
		}
	}
}

func TestSetupDefaultAdminConcurrency(t *testing.T) {
	t.Run("simple mode admin uses higher concurrency", func(t *testing.T) {
		t.Setenv("RUN_MODE", "simple")
		if got := setupDefaultAdminConcurrency(); got != simpleModeAdminConcurrency {
			t.Fatalf("setupDefaultAdminConcurrency()=%d, want %d", got, simpleModeAdminConcurrency)
		}
	})

	t.Run("standard mode keeps existing default", func(t *testing.T) {
		t.Setenv("RUN_MODE", "standard")
		if got := setupDefaultAdminConcurrency(); got != defaultUserConcurrency {
			t.Fatalf("setupDefaultAdminConcurrency()=%d, want %d", got, defaultUserConcurrency)
		}
	})
}

func TestNeedsSetupSkipsWhenSkipSetupIsEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "true", value: "true"},
		{name: "one", value: "1"},
		{name: "yes", value: "yes"},
		{name: "trimmed mixed case true", value: "  TrUe  "},
		{name: "trimmed mixed case yes", value: "  YeS  "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DATA_DIR", t.TempDir())
			t.Setenv("SKIP_SETUP", tc.value)

			if NeedsSetup() {
				t.Fatalf("NeedsSetup() = true, want false when SKIP_SETUP is enabled")
			}
		})
	}
}

func TestNeedsSetupFallsBackToFileDetectionWhenSkipSetupIsDisabled(t *testing.T) {
	tests := []struct {
		name         string
		skipSetupSet bool
		skipSetup    string
		markerFile   string
		want         bool
	}{
		{
			name: "unset without installation files",
			want: true,
		},
		{
			name:         "false without installation files",
			skipSetupSet: true,
			skipSetup:    " false ",
			want:         true,
		},
		{
			name:         "invalid value without installation files",
			skipSetupSet: true,
			skipSetup:    "enabled",
			want:         true,
		},
		{
			name:         "config file exists",
			skipSetupSet: true,
			skipSetup:    "false",
			markerFile:   ConfigFileName,
			want:         false,
		},
		{
			name:         "install lock file exists",
			skipSetupSet: true,
			skipSetup:    "invalid",
			markerFile:   InstallLockFile,
			want:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dataDir := t.TempDir()
			t.Setenv("DATA_DIR", dataDir)
			if tc.skipSetupSet {
				t.Setenv("SKIP_SETUP", tc.skipSetup)
			} else {
				originalValue, wasSet := os.LookupEnv("SKIP_SETUP")
				if err := os.Unsetenv("SKIP_SETUP"); err != nil {
					t.Fatalf("Unsetenv(SKIP_SETUP) error = %v", err)
				}
				t.Cleanup(func() {
					if wasSet {
						_ = os.Setenv("SKIP_SETUP", originalValue)
						return
					}
					_ = os.Unsetenv("SKIP_SETUP")
				})
			}

			if tc.markerFile != "" {
				if err := os.WriteFile(filepath.Join(dataDir, tc.markerFile), nil, 0o600); err != nil {
					t.Fatalf("WriteFile(%s) error = %v", tc.markerFile, err)
				}
			}

			if got := NeedsSetup(); got != tc.want {
				t.Fatalf("NeedsSetup() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSetupMigrationTimeout(t *testing.T) {
	t.Run("uses default timeout when unset", func(t *testing.T) {
		cfg := &SetupConfig{}
		if got := cfg.migrationTimeout(); got != 60*time.Second {
			t.Fatalf("migrationTimeout()=%s, want 60s", got)
		}
	})

	t.Run("uses configured timeout", func(t *testing.T) {
		cfg := &SetupConfig{MigrationTimeoutSeconds: 300}
		if got := cfg.migrationTimeout(); got != 300*time.Second {
			t.Fatalf("migrationTimeout()=%s, want 300s", got)
		}
	})
}

func TestAutoSetupRequiresExplicitAdminPassword(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "")
	cfg := &SetupConfig{
		Database: DatabaseConfig{Host: "postgres", Port: 5432, User: "postgres", DBName: "sub2api", SSLMode: "disable"},
		Redis:    RedisConfig{Host: "redis", Port: 6379},
		Admin:    AdminConfig{Email: "admin@example.com"},
		Server:   ServerConfig{Host: "0.0.0.0", Port: 8080},
	}
	if err := validateSetupCredentials(cfg); err == nil {
		t.Fatal("expected an empty administrator password to be rejected")
	}
}

func TestSetupCredentialsRejectUnsafeAutoSetupDatabaseName(t *testing.T) {
	cfg := &SetupConfig{
		Database: DatabaseConfig{Host: "postgres", Port: 5432, User: "postgres", DBName: "sub2api;DROP_TABLE", SSLMode: "disable"},
		Redis:    RedisConfig{Host: "redis", Port: 6379},
		Admin:    AdminConfig{Email: "admin@example.com", Password: "strong-password"},
		Server:   ServerConfig{Host: "0.0.0.0", Port: 8080},
	}
	if err := validateSetupCredentials(cfg); err == nil {
		t.Fatal("expected an unsafe database name to be rejected before auto setup")
	}
}

func TestWriteConfigFileKeepsDefaultUserConcurrency(t *testing.T) {
	t.Setenv("RUN_MODE", "simple")
	t.Setenv("DATA_DIR", t.TempDir())

	if err := writeConfigFile(&SetupConfig{}); err != nil {
		t.Fatalf("writeConfigFile() error = %v", err)
	}

	data, err := os.ReadFile(GetConfigFilePath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(data), "user_concurrency: 5") {
		t.Fatalf("config missing default user concurrency, got:\n%s", string(data))
	}
}

func TestWriteConfigFilePersistsAdminIdentityWithoutPassword(t *testing.T) {
	t.Setenv("RUN_MODE", "simple")
	t.Setenv("DATA_DIR", t.TempDir())
	cfg := &SetupConfig{Admin: AdminConfig{Email: "owner@example.com", Password: "do-not-persist"}}

	if err := writeConfigFile(cfg); err != nil {
		t.Fatalf("writeConfigFile() error = %v", err)
	}
	data, err := os.ReadFile(GetConfigFilePath())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "admin_email: owner@example.com") {
		t.Fatalf("config missing super-admin identity, got:\n%s", content)
	}
	if strings.Contains(content, cfg.Admin.Password) {
		t.Fatal("bootstrap administrator password was persisted")
	}
}

func TestWriteConfigFileIncludesRedisUsername(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())

	if err := writeConfigFile(&SetupConfig{
		Redis: RedisConfig{
			Host:     "redis",
			Port:     6379,
			Username: "app-user",
		},
	}); err != nil {
		t.Fatalf("writeConfigFile() error = %v", err)
	}

	data, err := os.ReadFile(GetConfigFilePath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(data), "username: app-user") {
		t.Fatalf("config missing Redis username, got:\n%s", string(data))
	}
}

func TestBuildDatabaseConnectionDSNsUsesPostgresForBootstrap(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "db",
		Port:     5432,
		User:     "sub2api",
		Password: "secret",
		DBName:   "sub2api",
		SSLMode:  "disable",
	}

	bootstrapDSN, targetDSN := buildDatabaseConnectionDSNs(cfg)

	if !strings.Contains(bootstrapDSN, "dbname=postgres") {
		t.Fatalf("bootstrap DSN = %q, want default postgres database", bootstrapDSN)
	}
	if strings.Contains(bootstrapDSN, "dbname=sub2api") {
		t.Fatalf("bootstrap DSN = %q, should not connect to target database before checking/creating it", bootstrapDSN)
	}
	if !strings.Contains(targetDSN, "dbname=sub2api") {
		t.Fatalf("target DSN = %q, want configured database", targetDSN)
	}
}
