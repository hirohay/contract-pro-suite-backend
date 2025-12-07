package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"contract-pro-suite/internal/shared/config"
	"contract-pro-suite/internal/shared/db"
	"contract-pro-suite/services/auth/domain"
	dbgen "contract-pro-suite/sqlc"
)

// MockOperatorRepository モックオペレーターリポジトリ
type MockOperatorRepository struct {
	mock.Mock
}

func (m *MockOperatorRepository) GetByID(ctx context.Context, operatorID uuid.UUID) (dbgen.Operator, error) {
	args := m.Called(ctx, operatorID)
	if args.Get(0) == nil {
		return dbgen.Operator{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Operator), args.Error(1)
}

func (m *MockOperatorRepository) GetByEmail(ctx context.Context, email string) (dbgen.Operator, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return dbgen.Operator{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Operator), args.Error(1)
}

func (m *MockOperatorRepository) List(ctx context.Context, limit, offset int32) ([]dbgen.Operator, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.Operator), args.Error(1)
}

func (m *MockOperatorRepository) Create(ctx context.Context, params dbgen.CreateOperatorParams) (dbgen.Operator, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.Operator{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Operator), args.Error(1)
}

func (m *MockOperatorRepository) Update(ctx context.Context, params dbgen.UpdateOperatorParams) (dbgen.Operator, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.Operator{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Operator), args.Error(1)
}

func (m *MockOperatorRepository) Delete(ctx context.Context, operatorID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, operatorID, deletedBy)
	return args.Error(0)
}

// MockClientUserRepository モッククライアントユーザーリポジトリ
type MockClientUserRepository struct {
	mock.Mock
}

func (m *MockClientUserRepository) GetByID(ctx context.Context, clientUserID uuid.UUID) (dbgen.ClientUser, error) {
	args := m.Called(ctx, clientUserID)
	if args.Get(0) == nil {
		return dbgen.ClientUser{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) GetByEmail(ctx context.Context, clientID uuid.UUID, email string) (dbgen.ClientUser, error) {
	args := m.Called(ctx, clientID, email)
	if args.Get(0) == nil {
		return dbgen.ClientUser{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) List(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]dbgen.ClientUser, error) {
	args := m.Called(ctx, clientID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) Create(ctx context.Context, params dbgen.CreateClientUserParams) (dbgen.ClientUser, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.ClientUser{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) Update(ctx context.Context, params dbgen.UpdateClientUserParams) (dbgen.ClientUser, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.ClientUser{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) Delete(ctx context.Context, clientUserID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, clientUserID, deletedBy)
	return args.Error(0)
}

// MockClientRepository モッククライアントリポジトリ
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) GetByID(ctx context.Context, clientID uuid.UUID) (dbgen.Client, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) GetBySlug(ctx context.Context, slug string) (dbgen.Client, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) GetByCompanyCode(ctx context.Context, companyCode string) (dbgen.Client, error) {
	args := m.Called(ctx, companyCode)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) List(ctx context.Context, limit, offset int32) ([]dbgen.Client, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) Create(ctx context.Context, params dbgen.CreateClientParams) (dbgen.Client, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) Update(ctx context.Context, params dbgen.UpdateClientParams) (dbgen.Client, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) Delete(ctx context.Context, clientID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, clientID, deletedBy)
	return args.Error(0)
}

func TestGetUserContext(t *testing.T) {
	mockOperatorRepo := new(MockOperatorRepository)
	mockClientUserRepo := new(MockClientUserRepository)
	mockClientRepo := new(MockClientRepository)
	
	// テスト用の設定とデータベース（モック）
	cfg := &config.Config{
		SupabaseURL:            "https://test.supabase.co",
		SupabaseServiceRoleKey: "test-key",
		SupabaseJWTSecret:      "test-secret",
		SupabaseDBURL:          "postgres://test",
	}
	database := &db.DB{} // モック（実際の接続は不要）

	usecase := NewAuthUsecase(mockOperatorRepo, mockClientUserRepo, mockClientRepo, cfg, database)

	testUserID := uuid.New()
	operator := dbgen.Operator{
		OperatorID: pgtype.UUID{Bytes: testUserID, Valid: true},
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "Operator",
		Status:     "ACTIVE",
	}

	tests := []struct {
		name      string
		jwtUserID string
		setupMock func()
		wantErr   bool
		wantType  domain.UserType
	}{
		{
			name:      "operator found",
			jwtUserID: testUserID.String(),
			setupMock: func() {
				mockOperatorRepo.On("GetByID", mock.Anything, testUserID).Return(operator, nil)
			},
			wantErr:  false,
			wantType: domain.UserTypeOperator,
		},
		{
			name:      "invalid user ID",
			jwtUserID: "invalid-uuid",
			setupMock: func() {},
			wantErr:   true,
		},
		{
			name:      "user not found",
			jwtUserID: uuid.New().String(),
			setupMock: func() {
				mockOperatorRepo.On("GetByID", mock.Anything, mock.Anything).Return(dbgen.Operator{}, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOperatorRepo.ExpectedCalls = nil
			mockOperatorRepo.Calls = nil
			tt.setupMock()

			ctx := context.Background()
			userCtx, err := usecase.GetUserContext(ctx, tt.jwtUserID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, userCtx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, userCtx)
				assert.Equal(t, tt.wantType, userCtx.UserType)
			}
		})
	}
}

func TestValidateClientAccess(t *testing.T) {
	mockOperatorRepo := new(MockOperatorRepository)
	mockClientUserRepo := new(MockClientUserRepository)
	mockClientRepo := new(MockClientRepository)
	
	// テスト用の設定とデータベース（モック）
	cfg := &config.Config{
		SupabaseURL:            "https://test.supabase.co",
		SupabaseServiceRoleKey: "test-key",
		SupabaseJWTSecret:      "test-secret",
		SupabaseDBURL:          "postgres://test",
	}
	database := &db.DB{} // モック（実際の接続は不要）

	usecase := NewAuthUsecase(mockOperatorRepo, mockClientUserRepo, mockClientRepo, cfg, database)

	testUserID := uuid.New()
	testClientID := uuid.New()

	tests := []struct {
		name      string
		userCtx   *domain.UserContext
		clientID  uuid.UUID
		wantErr   bool
	}{
		{
			name: "operator access granted",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
			},
			clientID: testClientID,
			wantErr:  false,
		},
		{
			name: "client user access granted",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeClientUser,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			clientID: testClientID,
			wantErr:  false,
		},
		{
			name: "client user access denied",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeClientUser,
				Email:    "test@example.com",
				ClientID: uuid.New(),
			},
			clientID: testClientID,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := usecase.ValidateClientAccess(ctx, tt.userCtx, tt.clientID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckPermission(t *testing.T) {
	mockOperatorRepo := new(MockOperatorRepository)
	mockClientUserRepo := new(MockClientUserRepository)
	mockClientRepo := new(MockClientRepository)
	
	// テスト用の設定とデータベース（モック）
	cfg := &config.Config{
		SupabaseURL:            "https://test.supabase.co",
		SupabaseServiceRoleKey: "test-key",
		SupabaseJWTSecret:      "test-secret",
		SupabaseDBURL:          "postgres://test",
	}
	database := &db.DB{} // モック（実際の接続は不要）

	usecase := NewAuthUsecase(mockOperatorRepo, mockClientUserRepo, mockClientRepo, cfg, database)

	testUserID := uuid.New()

	tests := []struct {
		name     string
		userCtx  *domain.UserContext
		feature  string
		action   string
		wantErr  bool
	}{
		{
			name: "operator permission check",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
			},
			feature: "contracts",
			action:  "READ",
			wantErr: false,
		},
		{
			name: "client user permission check",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeClientUser,
				Email:    "test@example.com",
			},
			feature: "contracts",
			action:  "READ",
			wantErr: false,
		},
		{
			name: "unknown user type",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserType("UNKNOWN"),
				Email:    "test@example.com",
			},
			feature: "contracts",
			action:  "READ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := usecase.CheckPermission(ctx, tt.userCtx, tt.feature, tt.action)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

