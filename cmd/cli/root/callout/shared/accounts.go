package shared

import (
	"context"
	"time"

	jwt "github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/rs/zerolog"
)

type (
	CreateSimpleAccountRequest struct {
		Name          string        `json:"name"`
		IssuerKeyPair nkeys.KeyPair `json:"issuer_key_pair"`
	}
	UpdateSimpleAccountRequest struct {
		Original      *CreateSimpleAccountResponse `json:"original"`
		IssuerKeyPair nkeys.KeyPair                `json:"issuer_key_pair"`
	}
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
		// Audience is either a well known name that is in the static config like "SYS", or a public key id
		Audience string `json:"audience"`
	}
	CreateSimpleAccountResponse struct {
		CommonAccountData
	}
)

func CreateSimpleAccount(ctx context.Context, request *CreateSimpleAccountRequest) (*CreateSimpleAccountResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "CreateSimpleAccount").Logger()
	// create an account keypair
	akp, err := nkeys.CreateAccount()
	if err != nil {
		log.Error().Err(err).Msg("failed to create account")
		return nil, err
	}
	// extract the public key for the account
	apk, err := akp.PublicKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to get public key")
		return nil, err
	}
	askp := akp

	// create the claim for the account using the public key of the account
	ac := jwt.NewAccountClaims(apk)
	ac.Name = request.Name
	ac.Expires = time.Now().Add(time.Minute * 2).Unix()
	ac.Limits.JetStreamLimits.DiskStorage = -1
	ac.Limits.JetStreamLimits.MemoryStorage = -1

	// add the signing key (public) to the account
	ac.SigningKeys.Add(apk)

	// now we could encode an issue the account using the operator
	// key that we generated above, but this will illustrate that
	// the account could be self-signed, and given to the operator
	// who can then re-sign it
	accountJWT, err := ac.Encode(request.IssuerKeyPair)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode account")
		return nil, err
	}

	resp := &CreateSimpleAccountResponse{
		CommonAccountData: CommonAccountData{
			Name: request.Name,
			JWT:  accountJWT,
		},
	}
	resp.KeyPair.PublicKey, _ = akp.PublicKey()
	resp.KeyPair.PrivateKey, _ = akp.PrivateKey()
	resp.KeyPair.Seed, _ = akp.Seed()
	resp.SignerKeyPair.PublicKey, _ = askp.PublicKey()
	resp.SignerKeyPair.PrivateKey, _ = askp.PrivateKey()
	resp.SignerKeyPair.Seed, _ = askp.Seed()
	resp.Audience = resp.KeyPair.PublicKey
	return resp, nil
}
