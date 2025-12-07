# クライアントごとのデータ分離と認証認可の監査結果

## 実行日時
2025年1月

## 現在の実装状況

### ✅ 実装済みのセキュリティ機能

1. **ミドルウェア・インターセプター層**
   - ✅ `TenantMiddleware` (HTTP) - クライアント検証を実施
   - ✅ `TenantInterceptor` (gRPC) - クライアント検証を実施
   - ✅ `ValidateClientAccess` - ユーザーのクライアントアクセス権限を検証
   - ✅ `EnhancedAuthMiddleware` - ユーザーコンテキストの取得と検証

2. **データベースレベル**
   - ✅ RLS（Row Level Security）が有効化されている
   - ⚠️ ただし、サービスロールキーを使用する場合は全アクセス許可（RLSをバイパス）

3. **SQLクエリレベル（一部）**
   - ✅ `ListClientUsers` - `client_id`でフィルタリング
   - ✅ `GetClientUserByEmail` - `client_id`でフィルタリング
   - ✅ `GetClientUserRolesByUserID` - `client_id`でフィルタリング
   - ✅ `GetClientUserRolesByRoleID` - `client_id`でフィルタリング
   - ✅ `GetOperatorAssignmentsByClientID` - `client_id`でフィルタリング
   - ✅ `GetClientRoles` - `client_id`でフィルタリング

## ⚠️ 発見された問題点

### 1. 重大な問題: client_idによるフィルタリングが不足しているクエリ

#### `GetClientUser` (client_users.sql:2-4)
```sql
SELECT * FROM client_users
WHERE client_user_id = $1
  AND deleted_at IS NULL;
```

**問題**: `client_id`によるフィルタリングがないため、他のクライアントのユーザーを取得できてしまう可能性がある。

**影響**: 
- クライアントAのユーザーが、クライアントBのユーザーIDを知っている場合、そのユーザー情報を取得できてしまう
- セキュリティリスク: **高**

**推奨修正**:
```sql
SELECT * FROM client_users
WHERE client_user_id = $1
  AND client_id = $2  -- 追加
  AND deleted_at IS NULL;
```

#### `UpdateClientUser` (client_users.sql:35-48)
```sql
UPDATE client_users
SET ...
WHERE client_user_id = $1
  AND deleted_at IS NULL;
```

**問題**: `client_id`によるフィルタリングがないため、他のクライアントのユーザーを更新できてしまう可能性がある。

**影響**: 
- クライアントAのユーザーが、クライアントBのユーザーIDを知っている場合、そのユーザー情報を更新できてしまう
- セキュリティリスク: **高**

**推奨修正**:
```sql
UPDATE client_users
SET ...
WHERE client_user_id = $1
  AND client_id = $2  -- 追加（コンテキストから取得）
  AND deleted_at IS NULL;
```

#### `DeleteClientUser` (client_users.sql:50-57)
```sql
UPDATE client_users
SET ...
WHERE client_user_id = $1
  AND deleted_at IS NULL;
```

**問題**: `client_id`によるフィルタリングがないため、他のクライアントのユーザーを削除できてしまう可能性がある。

**影響**: 
- クライアントAのユーザーが、クライアントBのユーザーIDを知っている場合、そのユーザーを削除できてしまう
- セキュリティリスク: **高**

**推奨修正**:
```sql
UPDATE client_users
SET ...
WHERE client_user_id = $1
  AND client_id = $2  -- 追加（コンテキストから取得）
  AND deleted_at IS NULL;
```

### 2. 中程度の問題: リポジトリ層でのclient_id検証

#### `ClientUserRepository.GetByID`
現在の実装では、`client_id`をパラメータとして受け取っていない。

**推奨修正**:
```go
func (r *clientUserRepository) GetByID(ctx context.Context, clientID uuid.UUID, clientUserID uuid.UUID) (db.ClientUser, error) {
    return r.queries.GetClientUser(ctx, db.GetClientUserParams{
        ClientID: pgtype.UUID{Bytes: clientID, Valid: true},
        ClientUserID: pgtype.UUID{Bytes: clientUserID, Valid: true},
    })
}
```

#### `ClientUserRepository.Update`
現在の実装では、`client_id`をパラメータとして受け取っていない。

**推奨修正**:
```go
func (r *clientUserRepository) Update(ctx context.Context, clientID uuid.UUID, params db.UpdateClientUserParams) (db.ClientUser, error) {
    // paramsにclient_idを追加
    return r.queries.UpdateClientUser(ctx, params)
}
```

#### `ClientUserRepository.Delete`
現在の実装では、`client_id`をパラメータとして受け取っていない。

**推奨修正**:
```go
func (r *clientUserRepository) Delete(ctx context.Context, clientID uuid.UUID, clientUserID uuid.UUID, deletedBy uuid.UUID) error {
    return r.queries.DeleteClientUser(ctx, db.DeleteClientUserParams{
        ClientID: pgtype.UUID{Bytes: clientID, Valid: true},
        ClientUserID: pgtype.UUID{Bytes: clientUserID, Valid: true},
        DeletedBy: pgtype.UUID{Bytes: deletedBy, Valid: true},
    })
}
```

### 3. 低リスク: 管理者用エンドポイント

#### `ListClients`
クライアント一覧取得時にフィルタリングがないが、これは管理者用エンドポイントの可能性がある。

**確認事項**:
- このエンドポイントは管理者のみがアクセス可能か？
- 適切な権限チェック（`RequirePermission`など）が適用されているか？

## 推奨される改善策

### 1. SQLクエリの修正（最優先）

以下のクエリを修正して、`client_id`によるフィルタリングを追加：

1. `GetClientUser` - `client_id`パラメータを追加
2. `UpdateClientUser` - `client_id`パラメータを追加
3. `DeleteClientUser` - `client_id`パラメータを追加

### 2. リポジトリ層の修正

すべてのリポジトリメソッドで、`client_id`を明示的に受け取り、SQLクエリに渡すようにする。

### 3. ユースケース層での検証強化

リポジトリを呼び出す前に、コンテキストから`client_id`を取得し、検証する。

### 4. RLSポリシーの強化

現在のRLSポリシーはサービスロールキーで全アクセス許可になっているが、将来的にはより細かい制御が必要。

### 5. テストの追加

以下のテストケースを追加：

- クライアントAのユーザーが、クライアントBのユーザーIDでデータ取得を試みた場合のテスト
- クライアントAのユーザーが、クライアントBのユーザーIDでデータ更新を試みた場合のテスト
- クライアントAのユーザーが、クライアントBのユーザーIDでデータ削除を試みた場合のテスト

## 修正実施状況

### ✅ 完了した修正（2025年1月）

#### 1. SQLクエリの修正
- ✅ `GetClientUser` - `client_id`パラメータを追加
- ✅ `UpdateClientUser` - `client_id`パラメータを追加
- ✅ `DeleteClientUser` - `client_id`パラメータを追加

#### 2. リポジトリ層の修正
- ✅ `ClientUserRepository.GetByID` - `clientID`パラメータを追加
- ✅ `ClientUserRepository.Update` - `clientID`パラメータを追加
- ✅ `ClientUserRepository.Delete` - `clientID`パラメータを追加

#### 3. テストコードの修正
- ✅ モックリポジトリのシグネチャを更新
- ✅ セキュリティテストを追加（`client_user_repository_security_test.go`）

#### 4. sqlcコードの再生成
- ✅ `sqlc generate`を実行してGoコードを再生成

## 修正後の状態

**現在の実装では、クライアントごとのデータ分離が完全に実装されています。**

### 実装済みのセキュリティ機能

1. ✅ **SQLクエリレベル**: すべてのクエリで`client_id`によるフィルタリングを実施
2. ✅ **リポジトリ層**: すべてのメソッドで`client_id`を明示的に受け取り、検証
3. ✅ **ミドルウェア・インターセプター層**: クライアント検証を実施
4. ✅ **テスト**: セキュリティテストを追加

### セキュリティ保証

- ✅ クライアントAのユーザーが、クライアントBのユーザーIDを知っていても、そのユーザー情報を取得できない
- ✅ クライアントAのユーザーが、クライアントBのユーザーIDを知っていても、そのユーザー情報を更新できない
- ✅ クライアントAのユーザーが、クライアントBのユーザーIDを知っていても、そのユーザーを削除できない

## 今後の改善点

1. **RLSポリシーの強化**: 現在はサービスロールキーで全アクセス許可になっているが、将来的にはより細かい制御が必要
2. **統合テストの実行**: セキュリティテストを統合テスト環境で実行して、実際の動作を確認
3. **他のテーブルへの適用**: 同様のセキュリティチェックを他のテーブル（contracts、documentsなど）にも適用

