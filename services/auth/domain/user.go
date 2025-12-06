package domain

import (
	"github.com/google/uuid"
)

// UserType ユーザータイプ
type UserType string

const (
	UserTypeOperator   UserType = "OPERATOR"
	UserTypeClientUser UserType = "CLIENT_USER"
)

// UserContext 認証済みユーザーのコンテキスト
type UserContext struct {
	UserID   uuid.UUID
	UserType UserType
	Email    string
	ClientID uuid.UUID // オペレーターの場合は割り当てられたクライアントID、クライアントユーザーの場合は所属クライアントID
}

// Permission 権限
type Permission struct {
	Feature   string
	Action    string
	Granted   bool
	Conditions map[string]interface{}
}

