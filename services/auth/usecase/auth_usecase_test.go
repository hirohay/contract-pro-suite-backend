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

// MockOperatorAssignmentRepository モックオペレーター割当リポジトリ
type MockOperatorAssignmentRepository struct {
	mock.Mock
}

func (m *MockOperatorAssignmentRepository) GetByClientAndOperator(ctx context.Context, clientID, operatorID uuid.UUID) (dbgen.OperatorAssignment, error) {
	args := m.Called(ctx, clientID, operatorID)
	if args.Get(0) == nil {
		return dbgen.OperatorAssignment{}, args.Error(1)
	}
	return args.Get(0).(dbgen.OperatorAssignment), args.Error(1)
}

func (m *MockOperatorAssignmentRepository) GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]dbgen.OperatorAssignment, error) {
	args := m.Called(ctx, operatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.OperatorAssignment), args.Error(1)
}

func (m *MockOperatorAssignmentRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]dbgen.OperatorAssignment, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.OperatorAssignment), args.Error(1)
}

func (m *MockOperatorAssignmentRepository) Create(ctx context.Context, params dbgen.CreateOperatorAssignmentParams) (dbgen.OperatorAssignment, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.OperatorAssignment{}, args.Error(1)
	}
	return args.Get(0).(dbgen.OperatorAssignment), args.Error(1)
}

func (m *MockOperatorAssignmentRepository) Update(ctx context.Context, params dbgen.UpdateOperatorAssignmentParams) (dbgen.OperatorAssignment, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.OperatorAssignment{}, args.Error(1)
	}
	return args.Get(0).(dbgen.OperatorAssignment), args.Error(1)
}

func (m *MockOperatorAssignmentRepository) Delete(ctx context.Context, clientID, operatorID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, clientID, operatorID, deletedBy)
	return args.Error(0)
}

// MockClientRoleRepository モッククライアントロールリポジトリ
type MockClientRoleRepository struct {
	mock.Mock
}

func (m *MockClientRoleRepository) GetByID(ctx context.Context, roleID uuid.UUID) (dbgen.ClientRole, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return dbgen.ClientRole{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientRole), args.Error(1)
}

func (m *MockClientRoleRepository) GetByCode(ctx context.Context, clientID uuid.UUID, code string) (dbgen.ClientRole, error) {
	args := m.Called(ctx, clientID, code)
	if args.Get(0) == nil {
		return dbgen.ClientRole{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientRole), args.Error(1)
}

func (m *MockClientRoleRepository) List(ctx context.Context, clientID uuid.UUID) ([]dbgen.ClientRole, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.ClientRole), args.Error(1)
}

func (m *MockClientRoleRepository) Create(ctx context.Context, params dbgen.CreateClientRoleParams) (dbgen.ClientRole, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.ClientRole{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientRole), args.Error(1)
}

func (m *MockClientRoleRepository) Update(ctx context.Context, params dbgen.UpdateClientRoleParams) (dbgen.ClientRole, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.ClientRole{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientRole), args.Error(1)
}

func (m *MockClientRoleRepository) Delete(ctx context.Context, roleID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, roleID, deletedBy)
	return args.Error(0)
}

// MockClientRolePermissionRepository モッククライアントロール権限リポジトリ
type MockClientRolePermissionRepository struct {
	mock.Mock
}

func (m *MockClientRolePermissionRepository) GetByRoleFeatureAction(ctx context.Context, roleID uuid.UUID, feature, action string) (dbgen.ClientRolePermission, error) {
	args := m.Called(ctx, roleID, feature, action)
	if args.Get(0) == nil {
		return dbgen.ClientRolePermission{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientRolePermission), args.Error(1)
}

func (m *MockClientRolePermissionRepository) GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]dbgen.ClientRolePermission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.ClientRolePermission), args.Error(1)
}

func (m *MockClientRolePermissionRepository) GetByFeatureAndAction(ctx context.Context, roleID uuid.UUID, feature, action string) ([]dbgen.ClientRolePermission, error) {
	args := m.Called(ctx, roleID, feature, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.ClientRolePermission), args.Error(1)
}

func (m *MockClientRolePermissionRepository) Create(ctx context.Context, params dbgen.CreateClientRolePermissionParams) (dbgen.ClientRolePermission, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.ClientRolePermission{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientRolePermission), args.Error(1)
}

func (m *MockClientRolePermissionRepository) Update(ctx context.Context, params dbgen.UpdateClientRolePermissionParams) (dbgen.ClientRolePermission, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.ClientRolePermission{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientRolePermission), args.Error(1)
}

func (m *MockClientRolePermissionRepository) Delete(ctx context.Context, roleID uuid.UUID, feature, action string, deletedBy uuid.UUID) error {
	args := m.Called(ctx, roleID, feature, action, deletedBy)
	return args.Error(0)
}

func (m *MockClientRolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, roleID, deletedBy)
	return args.Error(0)
}

// MockClientUserRoleRepository モッククライアントユーザーロールリポジトリ
type MockClientUserRoleRepository struct {
	mock.Mock
}

func (m *MockClientUserRoleRepository) GetByUserAndRole(ctx context.Context, clientID, clientUserID, roleID uuid.UUID) (dbgen.ClientUserRole, error) {
	args := m.Called(ctx, clientID, clientUserID, roleID)
	if args.Get(0) == nil {
		return dbgen.ClientUserRole{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientUserRole), args.Error(1)
}

func (m *MockClientUserRoleRepository) GetByUserID(ctx context.Context, clientID, clientUserID uuid.UUID) ([]dbgen.ClientUserRole, error) {
	args := m.Called(ctx, clientID, clientUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.ClientUserRole), args.Error(1)
}

func (m *MockClientUserRoleRepository) GetByRoleID(ctx context.Context, clientID, roleID uuid.UUID) ([]dbgen.ClientUserRole, error) {
	args := m.Called(ctx, clientID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.ClientUserRole), args.Error(1)
}

func (m *MockClientUserRoleRepository) Create(ctx context.Context, params dbgen.CreateClientUserRoleParams) (dbgen.ClientUserRole, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.ClientUserRole{}, args.Error(1)
	}
	return args.Get(0).(dbgen.ClientUserRole), args.Error(1)
}

func (m *MockClientUserRoleRepository) Revoke(ctx context.Context, clientID, clientUserID, roleID uuid.UUID) error {
	args := m.Called(ctx, clientID, clientUserID, roleID)
	return args.Error(0)
}

func (m *MockClientUserRoleRepository) Delete(ctx context.Context, clientID, clientUserID, roleID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, clientID, clientUserID, roleID, deletedBy)
	return args.Error(0)
}

func TestGetUserContext(t *testing.T) {
	mockOperatorRepo := new(MockOperatorRepository)
	mockClientUserRepo := new(MockClientUserRepository)
	mockClientRepo := new(MockClientRepository)
	mockOperatorAssignmentRepo := new(MockOperatorAssignmentRepository)
	mockClientRoleRepo := new(MockClientRoleRepository)
	mockClientRolePermissionRepo := new(MockClientRolePermissionRepository)
	mockClientUserRoleRepo := new(MockClientUserRoleRepository)
	
	// テスト用の設定とデータベース（モック）
	cfg := &config.Config{
		SupabaseURL:            "https://test.supabase.co",
		SupabaseServiceRoleKey: "test-key",
		SupabaseJWTSecret:      "test-secret",
		SupabaseDBURL:          "postgres://test",
	}
	database := &db.DB{} // モック（実際の接続は不要）

	usecase := NewAuthUsecase(
		mockOperatorRepo,
		mockClientUserRepo,
		mockClientRepo,
		mockOperatorAssignmentRepo,
		mockClientRoleRepo,
		mockClientRolePermissionRepo,
		mockClientUserRoleRepo,
		cfg,
		database,
	)

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
				testClientID := uuid.New()
				assignment := dbgen.OperatorAssignment{
					ClientID:   pgtype.UUID{Bytes: testClientID, Valid: true},
					OperatorID: pgtype.UUID{Bytes: testUserID, Valid: true},
					Role:       "ADMIN",
					Status:     "ACTIVE",
				}
				mockOperatorAssignmentRepo.On("GetByOperatorID", mock.Anything, testUserID).Return([]dbgen.OperatorAssignment{assignment}, nil)
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
			mockOperatorAssignmentRepo.ExpectedCalls = nil
			mockOperatorAssignmentRepo.Calls = nil
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
	mockOperatorAssignmentRepo := new(MockOperatorAssignmentRepository)
	mockClientRoleRepo := new(MockClientRoleRepository)
	mockClientRolePermissionRepo := new(MockClientRolePermissionRepository)
	mockClientUserRoleRepo := new(MockClientUserRoleRepository)
	
	// テスト用の設定とデータベース（モック）
	cfg := &config.Config{
		SupabaseURL:            "https://test.supabase.co",
		SupabaseServiceRoleKey: "test-key",
		SupabaseJWTSecret:      "test-secret",
		SupabaseDBURL:          "postgres://test",
	}
	database := &db.DB{} // モック（実際の接続は不要）

	usecase := NewAuthUsecase(
		mockOperatorRepo,
		mockClientUserRepo,
		mockClientRepo,
		mockOperatorAssignmentRepo,
		mockClientRoleRepo,
		mockClientRolePermissionRepo,
		mockClientUserRoleRepo,
		cfg,
		database,
	)

	testUserID := uuid.New()
	testClientID := uuid.New()

	tests := []struct {
		name      string
		userCtx   *domain.UserContext
		clientID  uuid.UUID
		setupMock func()
		wantErr   bool
	}{
		{
			name: "operator access granted",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			clientID: testClientID,
			setupMock: func() {
				assignment := dbgen.OperatorAssignment{
					ClientID:   pgtype.UUID{Bytes: testClientID, Valid: true},
					OperatorID: pgtype.UUID{Bytes: testUserID, Valid: true},
					Role:       "ADMIN",
					Status:     "ACTIVE",
				}
				mockOperatorAssignmentRepo.On("GetByOperatorID", mock.Anything, testUserID).Return([]dbgen.OperatorAssignment{assignment}, nil)
			},
			wantErr: false,
		},
		{
			name: "client user access granted",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeClientUser,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			clientID:  testClientID,
			setupMock: func() {},
			wantErr:   false,
		},
		{
			name: "client user access denied",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeClientUser,
				Email:    "test@example.com",
				ClientID: uuid.New(),
			},
			clientID:  testClientID,
			setupMock: func() {},
			wantErr:   true,
		},
		{
			name: "operator access denied - not assigned",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			clientID: testClientID,
			setupMock: func() {
				mockOperatorAssignmentRepo.On("GetByOperatorID", mock.Anything, testUserID).Return([]dbgen.OperatorAssignment{}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOperatorAssignmentRepo.ExpectedCalls = nil
			mockOperatorAssignmentRepo.Calls = nil
			if tt.setupMock != nil {
				tt.setupMock()
			}
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
	mockOperatorAssignmentRepo := new(MockOperatorAssignmentRepository)
	mockClientRoleRepo := new(MockClientRoleRepository)
	mockClientRolePermissionRepo := new(MockClientRolePermissionRepository)
	mockClientUserRoleRepo := new(MockClientUserRoleRepository)
	
	// テスト用の設定とデータベース（モック）
	cfg := &config.Config{
		SupabaseURL:            "https://test.supabase.co",
		SupabaseServiceRoleKey: "test-key",
		SupabaseJWTSecret:      "test-secret",
		SupabaseDBURL:          "postgres://test",
	}
	database := &db.DB{} // モック（実際の接続は不要）

	usecase := NewAuthUsecase(
		mockOperatorRepo,
		mockClientUserRepo,
		mockClientRepo,
		mockOperatorAssignmentRepo,
		mockClientRoleRepo,
		mockClientRolePermissionRepo,
		mockClientUserRoleRepo,
		cfg,
		database,
	)

	testUserID := uuid.New()
	testClientID := uuid.New()
	testRoleID := uuid.New()

	tests := []struct {
		name      string
		userCtx   *domain.UserContext
		feature   string
		action    string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "operator ADMIN permission check - allowed",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			feature: "contracts",
			action:  "READ",
			setupMock: func() {
				assignment := dbgen.OperatorAssignment{
					ClientID:   pgtype.UUID{Bytes: testClientID, Valid: true},
					OperatorID: pgtype.UUID{Bytes: testUserID, Valid: true},
					Role:       "ADMIN",
					Status:     "ACTIVE",
				}
				mockOperatorAssignmentRepo.On("GetByClientAndOperator", mock.Anything, testClientID, testUserID).Return(assignment, nil)
			},
			wantErr: false,
		},
		{
			name: "operator VIEWER permission check - read only",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			feature: "contracts",
			action:  "READ",
			setupMock: func() {
				assignment := dbgen.OperatorAssignment{
					ClientID:   pgtype.UUID{Bytes: testClientID, Valid: true},
					OperatorID: pgtype.UUID{Bytes: testUserID, Valid: true},
					Role:       "VIEWER",
					Status:     "ACTIVE",
				}
				mockOperatorAssignmentRepo.On("GetByClientAndOperator", mock.Anything, testClientID, testUserID).Return(assignment, nil)
			},
			wantErr: false,
		},
		{
			name: "operator VIEWER permission check - write denied",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			feature: "contracts",
			action:  "WRITE",
			setupMock: func() {
				assignment := dbgen.OperatorAssignment{
					ClientID:   pgtype.UUID{Bytes: testClientID, Valid: true},
					OperatorID: pgtype.UUID{Bytes: testUserID, Valid: true},
					Role:       "VIEWER",
					Status:     "ACTIVE",
				}
				mockOperatorAssignmentRepo.On("GetByClientAndOperator", mock.Anything, testClientID, testUserID).Return(assignment, nil)
			},
			wantErr: true,
		},
		{
			name: "client user permission check - allowed",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeClientUser,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			feature: "contracts",
			action:  "READ",
			setupMock: func() {
				userRole := dbgen.ClientUserRole{
					ClientID:     pgtype.UUID{Bytes: testClientID, Valid: true},
					ClientUserID: pgtype.UUID{Bytes: testUserID, Valid: true},
					RoleID:       pgtype.UUID{Bytes: testRoleID, Valid: true},
				}
				mockClientUserRoleRepo.On("GetByUserID", mock.Anything, testClientID, testUserID).Return([]dbgen.ClientUserRole{userRole}, nil)
				permission := dbgen.ClientRolePermission{
					RoleID:  pgtype.UUID{Bytes: testRoleID, Valid: true},
					Feature: "contracts",
					Action:  "READ",
					Granted: true,
				}
				mockClientRolePermissionRepo.On("GetByRoleID", mock.Anything, testRoleID).Return([]dbgen.ClientRolePermission{permission}, nil)
			},
			wantErr: false,
		},
		{
			name: "client user permission check - denied",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeClientUser,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			feature: "contracts",
			action:  "WRITE",
			setupMock: func() {
				userRole := dbgen.ClientUserRole{
					ClientID:     pgtype.UUID{Bytes: testClientID, Valid: true},
					ClientUserID: pgtype.UUID{Bytes: testUserID, Valid: true},
					RoleID:       pgtype.UUID{Bytes: testRoleID, Valid: true},
				}
				mockClientUserRoleRepo.On("GetByUserID", mock.Anything, testClientID, testUserID).Return([]dbgen.ClientUserRole{userRole}, nil)
				permission := dbgen.ClientRolePermission{
					RoleID:  pgtype.UUID{Bytes: testRoleID, Valid: true},
					Feature: "contracts",
					Action:  "READ",
					Granted: true,
				}
				mockClientRolePermissionRepo.On("GetByRoleID", mock.Anything, testRoleID).Return([]dbgen.ClientRolePermission{permission}, nil)
			},
			wantErr: true,
		},
		{
			name: "client user permission check - no roles",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeClientUser,
				Email:    "test@example.com",
				ClientID: testClientID,
			},
			feature: "contracts",
			action:  "READ",
			setupMock: func() {
				mockClientUserRoleRepo.On("GetByUserID", mock.Anything, testClientID, testUserID).Return([]dbgen.ClientUserRole{}, nil)
			},
			wantErr: true,
		},
		{
			name: "operator not assigned to client",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
				ClientID: uuid.Nil,
			},
			feature:   "contracts",
			action:    "READ",
			setupMock: func() {},
			wantErr:   true,
		},
		{
			name: "unknown user type",
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserType("UNKNOWN"),
				Email:    "test@example.com",
			},
			feature:   "contracts",
			action:    "READ",
			setupMock: func() {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOperatorAssignmentRepo.ExpectedCalls = nil
			mockOperatorAssignmentRepo.Calls = nil
			mockClientUserRoleRepo.ExpectedCalls = nil
			mockClientUserRoleRepo.Calls = nil
			mockClientRolePermissionRepo.ExpectedCalls = nil
			mockClientRolePermissionRepo.Calls = nil
			if tt.setupMock != nil {
				tt.setupMock()
			}
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

