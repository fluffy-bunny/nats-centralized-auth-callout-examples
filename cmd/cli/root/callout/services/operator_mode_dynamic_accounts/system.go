package operator_mode_dynamic_accounts

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"

	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	nkeys "github.com/nats-io/nkeys"
)

type SystemConfig struct {
	OperatorAccount struct {
		Name    string `json:"name"`
		KeyPair struct {
			PublicKey  string `json:"public_key"`
			PrivateKey string `json:"private_key"`
			Seed       string `json:"seed"`
		} `json:"key_pair"`
		SignerKeyPair struct {
			PublicKey  string `json:"public_key"`
			PrivateKey string `json:"private_key"`
			Seed       string `json:"seed"`
		} `json:"signer_key_pair"`
		Jwt string `json:"jwt"`
	} `json:"operator_account"`
	SystemAccount struct {
		Name    string `json:"name"`
		KeyPair struct {
			PublicKey  string `json:"public_key"`
			PrivateKey string `json:"private_key"`
			Seed       string `json:"seed"`
		} `json:"key_pair"`
		SignerKeyPair struct {
			PublicKey  string `json:"public_key"`
			PrivateKey string `json:"private_key"`
			Seed       string `json:"seed"`
		} `json:"signer_key_pair"`
		Jwt         string `json:"jwt"`
		AccountUser struct {
			UserKeyPair struct {
				PublicKey  string `json:"public_key"`
				PrivateKey string `json:"private_key"`
				Seed       string `json:"seed"`
			} `json:"user_key_pair"`
			Jwt   string `json:"jwt"`
			Creds string `json:"creds"`
		} `json:"account_user"`
	} `json:"system_account"`
	AuthAccount struct {
		Name    string `json:"name"`
		KeyPair struct {
			PublicKey  string `json:"public_key"`
			PrivateKey string `json:"private_key"`
			Seed       string `json:"seed"`
		} `json:"key_pair"`
		SignerKeyPair struct {
			PublicKey  string `json:"public_key"`
			PrivateKey string `json:"private_key"`
			Seed       string `json:"seed"`
		} `json:"signer_key_pair"`
		Jwt         string `json:"jwt"`
		AccountUser struct {
			UserKeyPair struct {
				PublicKey  string `json:"public_key"`
				PrivateKey string `json:"private_key"`
				Seed       string `json:"seed"`
			} `json:"user_key_pair"`
			Jwt   string `json:"jwt"`
			Creds string `json:"creds"`
		} `json:"account_user"`
		SentinelUser struct {
			UserKeyPair struct {
				PublicKey  string `json:"public_key"`
				PrivateKey string `json:"private_key"`
				Seed       string `json:"seed"`
			} `json:"user_key_pair"`
			Jwt   string `json:"jwt"`
			Creds string `json:"creds"`
		} `json:"sentinel_user"`
	} `json:"auth_account"`
}

func (s *SystemConfig) GetCalloutIssuerKeyPair() (nkeys.KeyPair, error) {
	seed, err := base64.StdEncoding.DecodeString(s.AuthAccount.SignerKeyPair.Seed)
	if err != nil {
		return nil, err
	}
	return loadAndParseKeysB([]byte(seed), 'A')
}
func (s *SystemConfig) GetOperatorKeyPair() (nkeys.KeyPair, error) {
	seed, err := base64.StdEncoding.DecodeString(s.OperatorAccount.SignerKeyPair.Seed)
	if err != nil {
		return nil, err
	}
	return loadAndParseKeysB(seed, 'O')
}
func loadAndParseKeysB(seed []byte, kind byte) (nkeys.KeyPair, error) {
	if fluffycore_utils.IsEmptyOrNil(seed) {
		return nil, errors.New("key seed required")
	}

	if !bytes.HasPrefix(seed, []byte{'S', kind}) {
		return nil, fmt.Errorf("key must be a private key")
	}
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %w", err)
	}
	return kp, nil
}
