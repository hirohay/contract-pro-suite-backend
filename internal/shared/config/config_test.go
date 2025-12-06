package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 環境変数を設定
	os.Setenv("SUPABASE_DB_URL", "postgresql://test:test@localhost:5432/test")
	os.Setenv("SUPABASE_SERVICE_ROLE_KEY", "test-key")
	os.Setenv("SUPABASE_JWT_SECRET", "test-secret")
	os.Setenv("SUPABASE_URL", "https://test.supabase.co")
	defer func() {
		os.Unsetenv("SUPABASE_DB_URL")
		os.Unsetenv("SUPABASE_SERVICE_ROLE_KEY")
		os.Unsetenv("SUPABASE_JWT_SECRET")
		os.Unsetenv("SUPABASE_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.SupabaseDBURL != "postgresql://test:test@localhost:5432/test" {
		t.Errorf("Expected SupabaseDBURL to be set, got: %s", cfg.SupabaseDBURL)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				SupabaseDBURL:          "postgresql://test:test@localhost:5432/test",
				SupabaseServiceRoleKey: "test-key",
				SupabaseJWTSecret:     "test-secret",
				SupabaseURL:            "https://test.supabase.co",
			},
			wantErr: false,
		},
		{
			name: "missing SUPABASE_DB_URL",
			cfg: &Config{
				SupabaseServiceRoleKey: "test-key",
				SupabaseJWTSecret:     "test-secret",
				SupabaseURL:            "https://test.supabase.co",
			},
			wantErr: true,
		},
		{
			name: "missing SUPABASE_SERVICE_ROLE_KEY",
			cfg: &Config{
				SupabaseDBURL:      "postgresql://test:test@localhost:5432/test",
				SupabaseJWTSecret: "test-secret",
				SupabaseURL:        "https://test.supabase.co",
			},
			wantErr: true,
		},
		{
			name: "missing SUPABASE_JWT_SECRET",
			cfg: &Config{
				SupabaseDBURL:          "postgresql://test:test@localhost:5432/test",
				SupabaseServiceRoleKey: "test-key",
				SupabaseURL:            "https://test.supabase.co",
			},
			wantErr: true,
		},
		{
			name: "missing SUPABASE_URL",
			cfg: &Config{
				SupabaseDBURL:          "postgresql://test:test@localhost:5432/test",
				SupabaseServiceRoleKey: "test-key",
				SupabaseJWTSecret:     "test-secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAllowedDomains(t *testing.T) {
	tests := []struct {
		name           string
		allowedDomains string
		baseDomain     string
		want           []string
	}{
		{
			name:           "single domain",
			allowedDomains: "contractprosuite.com",
			baseDomain:     "contractprosuite.com",
			want:           []string{"contractprosuite.com"},
		},
		{
			name:           "multiple domains",
			allowedDomains: "contractprosuite.com,localhost",
			baseDomain:     "contractprosuite.com",
			want:           []string{"contractprosuite.com", "localhost"},
		},
		{
			name:           "empty string",
			allowedDomains: "",
			baseDomain:     "contractprosuite.com",
			want:           []string{"contractprosuite.com"},
		},
		{
			name:           "with spaces",
			allowedDomains: "contractprosuite.com, localhost , example.com",
			baseDomain:     "contractprosuite.com",
			want:           []string{"contractprosuite.com", "localhost", "example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				AllowedDomainsStr: tt.allowedDomains,
				BaseDomain:        tt.baseDomain,
			}

			got := cfg.AllowedDomains()
			if len(got) != len(tt.want) {
				t.Errorf("AllowedDomains() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, domain := range got {
				if domain != tt.want[i] {
					t.Errorf("AllowedDomains()[%d] = %s, want %s", i, domain, tt.want[i])
				}
			}
		})
	}
}

