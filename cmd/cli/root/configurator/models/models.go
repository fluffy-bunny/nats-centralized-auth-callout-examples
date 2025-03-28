package models

import (
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type (
	RawKeyPair struct {
		PublicKey  string `json:"public_key"`
		PrivateKey []byte `json:"private_key"`
		Seed       []byte `json:"seed"`
	}
	CommonAccountData struct {
		Name          string     `json:"name"`
		KeyPair       RawKeyPair `json:"key_pair"`
		SignerKeyPair RawKeyPair `json:"signer_key_pair"`
		JWT           string     `json:"jwt"`
	}
	CreateOperatorRequest struct {
		Name string `json:"name"`
	}
	WriteCredsRequest struct {
		Name  string `json:"name"`
		Creds []byte `json:"creds"`
	}
	WriteCredsResponse struct {
		CredsFile string `json:"creds_file"`
	}
	CreateOperatorResponse struct {
		OperatorAccount CommonAccountData            `json:"operator_account"`
		SystemAccount   *CreateSystemAccountResponse `json:"system_account"`
		AuthAccount     *CreateAuthAccountResponse   `json:"auth_account"`
	}

	CreateSimpleAccountRequest struct {
		Name          string        `json:"name"`
		IssuerKeyPair nkeys.KeyPair `json:"issuer_key_pair"`
	}
	CreateSimpleAccountResponse struct {
		CommonAccountData
		SentinelUser *CreateUserWithCredsResonse `json:"sentinel_user"`
	}

	UpdateSimpleAccountRequest struct {
		Original      *CreateSimpleAccountResponse `json:"original"`
		IssuerKeyPair nkeys.KeyPair                `json:"issuer_key_pair"`
	}
	CreateSystemAccountRequest struct {
		CreateSimpleAccountRequest
	}
	CreateSystemAccountResponse struct {
		CommonAccountData
		AccountUser *CreateUserWithCredsResonse `json:"account_user"`
	}
	CreateAuthAccountRequest struct {
		Name          string        `json:"name"`
		IssuerKeyPair nkeys.KeyPair `json:"issuer_key_pair"`
	}
	CreateAuthAccountResponse struct {
		CommonAccountData
		AccountUser  *CreateUserWithCredsResonse `json:"account_user"`
		SentinelUser *CreateUserWithCredsResonse `json:"sentinel_user"`
	}

	CreateUserWithCredsRequest struct {
		Name          string          `json:"name"`
		IssuerID      string          `json:"issuer_id"`
		SignerKeyPair nkeys.KeyPair   `json:"signer_key_pair"`
		Permissions   jwt.Permissions `json:"permissions"`
	}
	CreateUserWithCredsResonse struct {
		UserKeyPair RawKeyPair `json:"user_key_pair"`
		JWT         string     `json:"jwt"`
		Creds       []byte     `json:"creds"`
	}
)
