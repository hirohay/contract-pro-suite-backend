package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	dbgen "contract-pro-suite/sqlc"
)

// DefaultRoleConfig デフォルトロール設定
type DefaultRoleConfig struct {
	Code        string
	Name        string
	Description string
	Permissions []PermissionConfig
}

// PermissionConfig 権限設定
type PermissionConfig struct {
	Feature string
	Action  string
	Granted bool
}

// createDefaultRoles デフォルトロールと権限を作成
func (u *authUsecase) createDefaultRoles(ctx context.Context, queries *dbgen.Queries, clientID uuid.UUID) error {

	// デフォルトロール定義
	defaultRoles := []DefaultRoleConfig{
		{
			Code:        "system_admin",
			Name:        "システム管理者",
			Description: "システム設定含めた全ての操作が可能",
			Permissions: getAllPermissions(),
		},
		{
			Code:        "business_admin",
			Name:        "業務管理者",
			Description: "設定以外の全ての操作が可能、ワークフローの承認が可能",
			Permissions: getBusinessAdminPermissions(),
		},
		{
			Code:        "member",
			Name:        "メンバー",
			Description: "全ての操作が可能、ワークフローの承認ができない",
			Permissions: getMemberPermissions(),
		},
		{
			Code:        "readonly",
			Name:        "閲覧のみ",
			Description: "全て読み取りのみ",
			Permissions: getReadonlyPermissions(),
		},
	}

	// 各デフォルトロールを作成
	for _, roleConfig := range defaultRoles {
		roleID := uuid.New()
		roleParams := dbgen.CreateClientRoleParams{
			RoleID:      pgtype.UUID{Bytes: roleID, Valid: true},
			ClientID:    pgtype.UUID{Bytes: clientID, Valid: true},
			Code:        roleConfig.Code,
			Name:        roleConfig.Name,
			Description: pgtype.Text{String: roleConfig.Description, Valid: true},
			IsSystem:    true,
		}

		_, err := queries.CreateClientRole(ctx, roleParams)
		if err != nil {
			return fmt.Errorf("failed to create role %s: %w", roleConfig.Code, err)
		}

		// 権限を作成
		for _, permConfig := range roleConfig.Permissions {
			conditionsJSON, _ := json.Marshal(map[string]interface{}{})
			permParams := dbgen.CreateClientRolePermissionParams{
				RoleID:     pgtype.UUID{Bytes: roleID, Valid: true},
				Feature:    permConfig.Feature,
				Action:     permConfig.Action,
				Granted:    permConfig.Granted,
				Conditions: conditionsJSON,
			}

			_, err := queries.CreateClientRolePermission(ctx, permParams)
			if err != nil {
				return fmt.Errorf("failed to create permission for role %s: %w", roleConfig.Code, err)
			}
		}
	}

	return nil
}

// getAllPermissions 全機能の全アクション権限を返す
func getAllPermissions() []PermissionConfig {
	features := []string{
		"system_settings",
		"contracts",
		"quotes",
		"invoices",
		"partners",
		"users",
		"workflows",
		"approvals",
		"documents",
	}
	actions := []string{"READ", "WRITE", "DELETE", "APPROVE"}

	permissions := make([]PermissionConfig, 0, len(features)*len(actions))
	for _, feature := range features {
		for _, action := range actions {
			permissions = append(permissions, PermissionConfig{
				Feature: feature,
				Action:  action,
				Granted: true,
			})
		}
	}
	return permissions
}

// getBusinessAdminPermissions 業務管理者の権限を返す
func getBusinessAdminPermissions() []PermissionConfig {
	features := []string{
		"contracts",
		"quotes",
		"invoices",
		"partners",
		"users",
		"workflows",
		"approvals",
		"documents",
	}
	actions := []string{"READ", "WRITE", "DELETE", "APPROVE"}

	permissions := make([]PermissionConfig, 0)
	// system_settingsはREADのみ
	permissions = append(permissions, PermissionConfig{
		Feature: "system_settings",
		Action:  "READ",
		Granted: true,
	})

	// その他の全機能は全アクション
	for _, feature := range features {
		for _, action := range actions {
			permissions = append(permissions, PermissionConfig{
				Feature: feature,
				Action:  action,
				Granted: true,
			})
		}
	}
	return permissions
}

// getMemberPermissions メンバーの権限を返す
func getMemberPermissions() []PermissionConfig {
	features := []string{
		"contracts",
		"quotes",
		"invoices",
		"partners",
		"users",
		"workflows",
		"documents",
	}
	actions := []string{"READ", "WRITE", "DELETE"}

	permissions := make([]PermissionConfig, 0)
	// system_settingsはREADのみ
	permissions = append(permissions, PermissionConfig{
		Feature: "system_settings",
		Action:  "READ",
		Granted: true,
	})

	// approvalsはREADのみ（承認不可）
	permissions = append(permissions, PermissionConfig{
		Feature: "approvals",
		Action:  "READ",
		Granted: true,
	})

	// その他の全機能はREAD, WRITE, DELETE
	for _, feature := range features {
		for _, action := range actions {
			permissions = append(permissions, PermissionConfig{
				Feature: feature,
				Action:  action,
				Granted: true,
			})
		}
	}
	return permissions
}

// getReadonlyPermissions 読み取りのみの権限を返す
func getReadonlyPermissions() []PermissionConfig {
	features := []string{
		"system_settings",
		"contracts",
		"quotes",
		"invoices",
		"partners",
		"users",
		"workflows",
		"approvals",
		"documents",
	}

	permissions := make([]PermissionConfig, 0, len(features))
	for _, feature := range features {
		permissions = append(permissions, PermissionConfig{
			Feature: feature,
			Action:  "READ",
			Granted: true,
		})
	}
	return permissions
}

