package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetAllPermissions 全権限取得関数のテスト
func TestGetAllPermissions(t *testing.T) {
	permissions := getAllPermissions()

	// 全機能（9機能）× 全アクション（4アクション）= 36個の権限が返されることを確認
	assert.Equal(t, 36, len(permissions), "全権限の数が正しくない")

	// 期待される機能リスト
	expectedFeatures := []string{
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

	// 期待されるアクションリスト
	expectedActions := []string{"READ", "WRITE", "DELETE", "APPROVE"}

	// 各機能とアクションの組み合わせが存在することを確認
	for _, feature := range expectedFeatures {
		for _, action := range expectedActions {
			found := false
			for _, perm := range permissions {
				if perm.Feature == feature && perm.Action == action && perm.Granted {
					found = true
					break
				}
			}
			assert.True(t, found, "権限が見つかりません: feature=%s, action=%s", feature, action)
		}
	}
}

// TestGetBusinessAdminPermissions 業務管理者権限取得関数のテスト
func TestGetBusinessAdminPermissions(t *testing.T) {
	permissions := getBusinessAdminPermissions()

	// system_settingsはREADのみ（1個）+ その他の8機能×4アクション（32個）= 33個
	assert.Equal(t, 33, len(permissions), "業務管理者権限の数が正しくない")

	// system_settingsはREADのみであることを確認
	systemSettingsCount := 0
	for _, perm := range permissions {
		if perm.Feature == "system_settings" {
			systemSettingsCount++
			assert.Equal(t, "READ", perm.Action, "system_settingsはREADのみである必要があります")
			assert.True(t, perm.Granted, "system_settingsの権限はGrantedである必要があります")
		}
	}
	assert.Equal(t, 1, systemSettingsCount, "system_settingsの権限は1個のみである必要があります")

	// その他の機能（contracts, quotes, invoices, partners, users, workflows, approvals, documents）は全アクション
	otherFeatures := []string{
		"contracts",
		"quotes",
		"invoices",
		"partners",
		"users",
		"workflows",
		"approvals",
		"documents",
	}
	expectedActions := []string{"READ", "WRITE", "DELETE", "APPROVE"}

	for _, feature := range otherFeatures {
		for _, action := range expectedActions {
			found := false
			for _, perm := range permissions {
				if perm.Feature == feature && perm.Action == action && perm.Granted {
					found = true
					break
				}
			}
			assert.True(t, found, "権限が見つかりません: feature=%s, action=%s", feature, action)
		}
	}
}

// TestGetMemberPermissions メンバー権限取得関数のテスト
func TestGetMemberPermissions(t *testing.T) {
	permissions := getMemberPermissions()

	// system_settingsはREADのみ（1個）+ approvalsはREADのみ（1個）+ その他の7機能×3アクション（21個）= 23個
	assert.Equal(t, 23, len(permissions), "メンバー権限の数が正しくない")

	// system_settingsはREADのみであることを確認
	systemSettingsCount := 0
	for _, perm := range permissions {
		if perm.Feature == "system_settings" {
			systemSettingsCount++
			assert.Equal(t, "READ", perm.Action, "system_settingsはREADのみである必要があります")
			assert.True(t, perm.Granted, "system_settingsの権限はGrantedである必要があります")
		}
	}
	assert.Equal(t, 1, systemSettingsCount, "system_settingsの権限は1個のみである必要があります")

	// approvalsはREADのみ（承認不可）であることを確認
	approvalsCount := 0
	for _, perm := range permissions {
		if perm.Feature == "approvals" {
			approvalsCount++
			assert.Equal(t, "READ", perm.Action, "approvalsはREADのみである必要があります（承認不可）")
			assert.True(t, perm.Granted, "approvalsの権限はGrantedである必要があります")
		}
	}
	assert.Equal(t, 1, approvalsCount, "approvalsの権限は1個のみである必要があります")

	// その他の機能（contracts, quotes, invoices, partners, users, workflows, documents）はREAD, WRITE, DELETE
	otherFeatures := []string{
		"contracts",
		"quotes",
		"invoices",
		"partners",
		"users",
		"workflows",
		"documents",
	}
	expectedActions := []string{"READ", "WRITE", "DELETE"}

	for _, feature := range otherFeatures {
		for _, action := range expectedActions {
			found := false
			for _, perm := range permissions {
				if perm.Feature == feature && perm.Action == action && perm.Granted {
					found = true
					break
				}
			}
			assert.True(t, found, "権限が見つかりません: feature=%s, action=%s", feature, action)
		}
	}

	// APPROVEアクションが含まれていないことを確認（approvals以外）
	for _, perm := range permissions {
		if perm.Action == "APPROVE" && perm.Feature != "approvals" {
			t.Errorf("メンバーロールにはAPPROVEアクションが含まれるべきではありません: feature=%s", perm.Feature)
		}
	}
}

// TestGetReadonlyPermissions 読み取り専用権限取得関数のテスト
func TestGetReadonlyPermissions(t *testing.T) {
	permissions := getReadonlyPermissions()

	// 全機能（9機能）× READのみ = 9個の権限が返されることを確認
	assert.Equal(t, 9, len(permissions), "読み取り専用権限の数が正しくない")

	// 期待される機能リスト
	expectedFeatures := []string{
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

	// 各機能のREAD権限が存在することを確認
	for _, feature := range expectedFeatures {
		found := false
		for _, perm := range permissions {
			if perm.Feature == feature && perm.Action == "READ" && perm.Granted {
				found = true
				break
			}
		}
		assert.True(t, found, "READ権限が見つかりません: feature=%s", feature)
	}

	// READ以外のアクションが含まれていないことを確認
	for _, perm := range permissions {
		assert.Equal(t, "READ", perm.Action, "読み取り専用ロールにはREADアクションのみが含まれるべきです: feature=%s, action=%s", perm.Feature, perm.Action)
		assert.True(t, perm.Granted, "読み取り専用ロールの権限はGrantedである必要があります")
	}
}

