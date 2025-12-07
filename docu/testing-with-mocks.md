# モックを使ったテストの仕組み

## 概要

現在のテストは、**実際のデータベースに接続せずに**、**モック（モックオブジェクト）**を使用してテストを実行しています。

## モックとは？

モックは、実際のオブジェクトの代わりに使用される「偽物」のオブジェクトです。テスト時に、データベースや外部サービスに接続せずに、事前に定義した動作を返すことができます。

## 実装例

### 1. モックリポジトリの定義

```go
// MockOperatorRepository モックオペレーターリポジトリ
type MockOperatorRepository struct {
    mock.Mock  // testify/mockパッケージを使用
}

func (m *MockOperatorRepository) GetByID(ctx context.Context, operatorID uuid.UUID) (dbgen.Operator, error) {
    // モックが呼び出されたときの引数を記録
    args := m.Called(ctx, operatorID)
    
    // 事前に設定された戻り値を返す
    if args.Get(0) == nil {
        return dbgen.Operator{}, args.Error(1)
    }
    return args.Get(0).(dbgen.Operator), args.Error(1)
}
```

### 2. テストでの使用方法

```go
func TestGetUserContext(t *testing.T) {
    // モックリポジトリを作成
    mockOperatorRepo := new(MockOperatorRepository)
    
    // テスト用のデータを準備（実際のDBには保存されない）
    testUserID := uuid.New()
    operator := dbgen.Operator{
        OperatorID: pgtype.UUID{Bytes: testUserID, Valid: true},
        Email:      "test@example.com",
        FirstName:  "Test",
        LastName:   "Operator",
        Status:     "ACTIVE",
    }
    
    // モックの動作を定義
    // 「GetByIDが呼ばれたら、operatorとnil（エラーなし）を返す」と設定
    mockOperatorRepo.On("GetByID", mock.Anything, testUserID).Return(operator, nil)
    
    // 実際のデータベース接続は不要
    database := &db.DB{} // 空の構造体（実際の接続なし）
    
    // モックリポジトリを使ってユースケースを作成
    usecase := NewAuthUsecase(
        mockOperatorRepo,  // 実際のリポジトリの代わりにモックを使用
        // ... 他のモックリポジトリ
    )
    
    // テスト実行
    ctx := context.Background()
    userCtx, err := usecase.GetUserContext(ctx, testUserID.String())
    
    // アサーション
    assert.NoError(t, err)
    assert.NotNil(t, userCtx)
}
```

## モックの動作フロー

```
1. テスト開始
   ↓
2. モックリポジトリを作成
   ↓
3. モックの動作を定義（On().Return()）
   「このメソッドが呼ばれたら、この値を返す」
   ↓
4. ユースケースにモックを注入
   ↓
5. ユースケースのメソッドを呼び出し
   ↓
6. ユースケース内部でリポジトリのメソッドを呼び出し
   ↓
7. モックが事前に定義した値を返す
   ↓
8. ユースケースが結果を返す
   ↓
9. テストで結果を検証
```

## モックの利点

### ✅ 高速
- データベースへの接続が不要
- ネットワーク通信が不要
- テストが非常に高速に実行される

### ✅ 独立性
- データベースの状態に依存しない
- 他のテストの影響を受けない
- 並列実行が可能

### ✅ 再現性
- 同じ条件で常に同じ結果が得られる
- データベースの状態に依存しない

### ✅ エッジケースのテストが容易
- エラーケースを簡単にテストできる
- 存在しないデータのテストが容易

## 実際のコード例

### テストケース: "operator found"

```go
{
    name:      "operator found",
    jwtUserID: testUserID.String(),
    setupMock: func() {
        // モックの動作を定義
        // 「GetByIDが呼ばれたら、operatorとnil（エラーなし）を返す」
        mockOperatorRepo.On("GetByID", mock.Anything, testUserID).Return(operator, nil)
        
        // オペレーター割り当てのモックも定義
        testClientID := uuid.New()
        assignment := dbgen.OperatorAssignment{
            ClientID:   pgtype.UUID{Bytes: testClientID, Valid: true},
            OperatorID: pgtype.UUID{Bytes: testUserID, Valid: true},
            Role:       "ADMIN",
            Status:     "ACTIVE",
        }
        mockOperatorAssignmentRepo.On("GetByOperatorID", mock.Anything, testUserID).
            Return([]dbgen.OperatorAssignment{assignment}, nil)
    },
    wantErr:  false,
    wantType: domain.UserTypeOperator,
}
```

### テストケース: "user not found"

```go
{
    name:      "user not found",
    jwtUserID: uuid.New().String(),
    setupMock: func() {
        // エラーケースのモック
        // 「GetByIDが呼ばれたら、エラーを返す」
        mockOperatorRepo.On("GetByID", mock.Anything, mock.Anything).
            Return(dbgen.Operator{}, assert.AnError)
    },
    wantErr: true,
}
```

## モック vs 統合テスト

### モックを使ったテスト（ユニットテスト）
- ✅ 高速
- ✅ 独立性が高い
- ✅ ロジックのテストに適している
- ❌ 実際のデータベースとの連携はテストできない

### 統合テスト
- ✅ 実際のデータベースとの連携をテストできる
- ✅ エンドツーエンドの動作を確認できる
- ❌ 遅い
- ❌ データベースのセットアップが必要

## 現在のテスト構成

### ユニットテスト（モック使用）
- `TestGetUserContext` - モックを使用
- `TestValidateClientAccess` - モックを使用
- `TestCheckPermission` - モックを使用

### 統合テスト（実際のDB使用）
- `TestClientUserRepository_GetByID_ClientIsolation` - スキップ（統合テスト環境で実行）
- `TestClientUserRepository_Update_ClientIsolation` - スキップ（統合テスト環境で実行）
- `TestClientUserRepository_Delete_ClientIsolation` - スキップ（統合テスト環境で実行）

## まとめ

現在のテストは、**モックオブジェクト**を使用して、実際のデータベースに接続せずにテストを実行しています。

- モックは事前に定義した動作を返す「偽物」のオブジェクト
- データベースへの接続が不要で高速
- ロジックのテストに適している
- 実際のデータベースとの連携をテストするには、統合テストが必要

## 参考

- [testify/mock パッケージ](https://github.com/stretchr/testify#mock-package)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)

