package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"contract-pro-suite/internal/shared/config"
	db "contract-pro-suite/sqlc"
)

// MockClientRepository モッククライアントリポジトリ
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) GetByID(ctx context.Context, clientID uuid.UUID) (db.Client, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return db.Client{}, args.Error(1)
	}
	return args.Get(0).(db.Client), args.Error(1)
}

func (m *MockClientRepository) GetBySlug(ctx context.Context, slug string) (db.Client, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return db.Client{}, args.Error(1)
	}
	return args.Get(0).(db.Client), args.Error(1)
}

func (m *MockClientRepository) GetByCompanyCode(ctx context.Context, companyCode string) (db.Client, error) {
	args := m.Called(ctx, companyCode)
	if args.Get(0) == nil {
		return db.Client{}, args.Error(1)
	}
	return args.Get(0).(db.Client), args.Error(1)
}

func (m *MockClientRepository) List(ctx context.Context, limit, offset int32) ([]db.Client, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Client), args.Error(1)
}

func (m *MockClientRepository) Create(ctx context.Context, params db.CreateClientParams) (db.Client, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return db.Client{}, args.Error(1)
	}
	return args.Get(0).(db.Client), args.Error(1)
}

func (m *MockClientRepository) Update(ctx context.Context, params db.UpdateClientParams) (db.Client, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return db.Client{}, args.Error(1)
	}
	return args.Get(0).(db.Client), args.Error(1)
}

func (m *MockClientRepository) Delete(ctx context.Context, clientID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, clientID, deletedBy)
	return args.Error(0)
}

func TestExtractSlugFromHost(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		baseDomain string
		want      string
		wantErr   bool
	}{
		{
			name:       "valid subdomain",
			host:       "acme-corp.contractprosuite.com",
			baseDomain: "contractprosuite.com",
			want:       "acme-corp",
			wantErr:    false,
		},
		{
			name:       "subdomain with port",
			host:       "acme-corp.contractprosuite.com:8080",
			baseDomain: "contractprosuite.com",
			want:       "acme-corp",
			wantErr:    false,
		},
		{
			name:       "no subdomain",
			host:       "contractprosuite.com",
			baseDomain: "contractprosuite.com",
			want:       "",
			wantErr:    true,
		},
		{
			name:       "invalid domain",
			host:       "example.com",
			baseDomain: "contractprosuite.com",
			want:       "",
			wantErr:    true,
		},
		{
			name:       "empty subdomain",
			host:       ".contractprosuite.com",
			baseDomain: "contractprosuite.com",
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSlugFromHost(tt.host, tt.baseDomain)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSlugFromHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractSlugFromHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateSubdomain(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid domain",
			host: "acme-corp.contractprosuite.com",
			cfg: &config.Config{
				BaseDomain:                "contractprosuite.com",
				AllowedDomainsStr:         "contractprosuite.com",
				EnableSubdomainValidation: true,
			},
			wantErr: false,
		},
		{
			name: "validation disabled",
			host: "example.com",
			cfg: &config.Config{
				BaseDomain:                "contractprosuite.com",
				AllowedDomainsStr:         "contractprosuite.com",
				EnableSubdomainValidation: false,
			},
			wantErr: false,
		},
		{
			name: "not allowed domain",
			host: "example.com",
			cfg: &config.Config{
				BaseDomain:                "contractprosuite.com",
				AllowedDomainsStr:         "contractprosuite.com",
				EnableSubdomainValidation: true,
			},
			wantErr: true,
		},
		{
			name: "localhost allowed",
			host: "localhost",
			cfg: &config.Config{
				BaseDomain:                "contractprosuite.com",
				AllowedDomainsStr:         "contractprosuite.com,localhost",
				EnableSubdomainValidation: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubdomain(tt.host, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubdomain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractClientID(t *testing.T) {
	// モックリポジトリのセットアップ
	mockRepo := new(MockClientRepository)
	testClientID := uuid.New()
	testClient := db.Client{
		ClientID: pgtype.UUID{
			Bytes: testClientID,
			Valid: true,
		},
		Slug:   "acme-corp",
		Status: "ACTIVE",
	}

	cfg := &config.Config{
		BaseDomain:                "contractprosuite.com",
		AllowedDomainsStr:         "contractprosuite.com,localhost",
		EnableSubdomainValidation: true,
		AppEnv:                    "development",
	}

	tests := []struct {
		name      string
		request   *http.Request
		setupMock func()
		want      uuid.UUID
		wantErr   bool
	}{
		{
			name: "extract from subdomain",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Host = "acme-corp.contractprosuite.com"
				return req
			}(),
			setupMock: func() {
				mockRepo.On("GetBySlug", mock.Anything, "acme-corp").Return(testClient, nil)
			},
			want:    testClientID,
			wantErr: false,
		},
		{
			name: "no client_id found",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Host = "example.com"
				return req
			}(),
			setupMock: func() {
				// モックは呼ばれない
			},
			want:    uuid.Nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			if tt.setupMock != nil {
				tt.setupMock()
			}

			got, err := ExtractClientID(tt.request, cfg, mockRepo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractClientID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ExtractClientID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClientIDFromContext(t *testing.T) {
	testClientID := uuid.New()

	tests := []struct {
		name    string
		ctx     func() interface{}
		want    uuid.UUID
		wantOk  bool
	}{
		{
			name: "client_id in context",
			ctx: func() interface{} {
				req := httptest.NewRequest("GET", "/", nil)
				ctx := req.Context()
				ctx = context.WithValue(ctx, clientIDContextKey, testClientID)
				return ctx
			},
			want:   testClientID,
			wantOk: true,
		},
		{
			name: "no client_id in context",
			ctx: func() interface{} {
				req := httptest.NewRequest("GET", "/", nil)
				return req.Context()
			},
			want:   uuid.Nil,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.ctx().(context.Context)
			got, ok := GetClientIDFromContext(ctx)
			if ok != tt.wantOk {
				t.Errorf("GetClientIDFromContext() ok = %v, want %v", ok, tt.wantOk)
			}
			if got != tt.want {
				t.Errorf("GetClientIDFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

