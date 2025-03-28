package shared

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	configurator_models "natsauth/cmd/cli/root/configurator/models"

	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	jwt "github.com/nats-io/jwt/v2"
	nkeys "github.com/nats-io/nkeys"
	zerolog "github.com/rs/zerolog"
)

/*
	{
	  "jti": "HQ6INVX2LMI4NDXYZMN5U4L4J5TL42PLEWG3NBXUNH6L6IOQJXHA",
	  "iat": 1737755077,
	  "iss": "ODRSGJVWECSLUL4ZMJITH3TBTTU3M35IY37UXMQAFUD6N7DZ5XMCDIAG",
	  "name": "SYS",
	  "sub": "ADI67YVFPNBQ3HCNK3XRCFYKD2W2MN2KDHXRDKEMNUANPMTEGL35ZULD",
	  "nats": {
	    "exports": [
	      {
	        "name": "account-monitoring-streams",
	        "subject": "$SYS.ACCOUNT.*.>",
	        "type": "stream",
	        "account_token_position": 3,
	        "description": "Account specific monitoring stream",
	        "info_url": "https://docs.nats.io/nats-server/configuration/sys_accounts"
	      },
	      {
	        "name": "account-monitoring-services",
	        "subject": "$SYS.REQ.ACCOUNT.*.*",
	        "type": "service",
	        "response_type": "Stream",
	        "account_token_position": 4,
	        "description": "Request account specific monitoring services for: SUBSZ, CONNZ, LEAFZ, JSZ and INFO",
	        "info_url": "https://docs.nats.io/nats-server/configuration/sys_accounts"
	      }
	    ],
	    "limits": {
	      "subs": -1,
	      "data": -1,
	      "payload": -1,
	      "imports": -1,
	      "exports": -1,
	      "wildcards": true,
	      "conn": -1,
	      "leaf": -1
	    },
	    "signing_keys": [
	      "AAILKV6P4MK3M7IMU26Q2HTYQNEWOUNPGOZXOCF4KPLFDOSM7QH4PSGU"
	    ],
	    "default_permissions": {
	      "pub": {},
	      "sub": {}
	    },
	    "authorization": {},
	    "type": "account",
	    "version": 2
	  }
	}
*/
func CreateSystemAccount(ctx context.Context, request *configurator_models.CreateSystemAccountRequest) (*configurator_models.CreateSystemAccountResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "CreateSystemAccount").Logger()
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

	// create the claim for the account using the public key of the account
	ac := jwt.NewAccountClaims(apk)
	ac.Name = request.Name
	// create a signing key that we can use for issuing users
	askp, err := nkeys.CreateAccount()
	if err != nil {
		log.Error().Err(err).Msg("failed to create account")
		return nil, err
	}
	// extract the public key
	aspk, err := askp.PublicKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to get public key")
		return nil, err
	}
	// add the signing key (public) to the account
	ac.SigningKeys.Add(aspk)

	ac.Exports.Add(&jwt.Export{
		Info: jwt.Info{
			Description: "Account specific monitoring stream",
			InfoURL:     "https://docs.nats.io/nats-server/configuration/sys_accounts",
		},
		Name:                 "account-monitoring-streams",
		Subject:              "$SYS.ACCOUNT.*.>",
		Type:                 jwt.Stream,
		AccountTokenPosition: 3,
	}, &jwt.Export{
		Info: jwt.Info{
			Description: "Request account specific monitoring services for: SUBSZ, CONNZ, LEAFZ, JSZ and INFO",

			InfoURL: "https://docs.nats.io/nats-server/configuration/sys_accounts",
		},
		Name:                 "account-monitoring-services",
		Subject:              "$SYS.REQ.ACCOUNT.*.*",
		Type:                 jwt.Service,
		ResponseType:         jwt.ResponseTypeStream,
		AccountTokenPosition: 4,
	})
	// now we could encode an issue the account using the operator
	// key that we generated above, but this will illustrate that
	// the account could be self-signed, and given to the operator
	// who can then re-sign it
	accountJWT, err := ac.Encode(request.IssuerKeyPair)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode account")
		return nil, err
	}

	resp := &configurator_models.CreateSystemAccountResponse{
		CommonAccountData: configurator_models.CommonAccountData{
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

	// generate a creds formatted file that can be used by a NATS client
	createUserCredsResponse, err := CreateUserWithCreds(ctx, &configurator_models.CreateUserWithCredsRequest{
		IssuerID:      apk,
		SignerKeyPair: askp,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to format user config")
		return nil, err
	}
	resp.AccountUser = createUserCredsResponse
	return resp, nil
}
func randomInt64(min, max int64) int64 {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	// Generate a random number in the range [min, max]
	return min + rand.Int63n(max-min+1)
}

// a POC for updating an account
func UpdateSimpleAccount(ctx context.Context, request *configurator_models.UpdateSimpleAccountRequest) (*configurator_models.CreateSimpleAccountResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "UpdateSimpleAccount").Logger()
	akp, _ := nkeys.FromSeed(request.Original.KeyPair.Seed)
	apk, _ := akp.PublicKey()
	askp, _ := nkeys.FromSeed(request.Original.SignerKeyPair.Seed)
	// extract the public key
	aspk, err := askp.PublicKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to get public key")
		return nil, err
	}

	// create the claim for the account using the public key of the account
	ac := jwt.NewAccountClaims(apk)
	ac.Name = request.Original.Name
	// create a random int64 between 1000000000 and 2000000000
	randomNum := randomInt64(1000000000, 2000000000)
	ac.Limits.JetStreamLimits.DiskStorage = randomNum
	ac.Limits.JetStreamLimits.MemoryStorage = randomNum

	ac.Expires = time.Now().Add(time.Minute * 2).Unix()

	// add the signing key (public) to the account
	ac.SigningKeys.Add(aspk)

	// now we could encode an issue the account using the operator
	// key that we generated above, but this will illustrate that
	// the account could be self-signed, and given to the operator
	// who can then re-sign it
	accountJWT, err := ac.Encode(request.IssuerKeyPair)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode account")
		return nil, err
	}
	request.Original.JWT = accountJWT

	return request.Original, nil
}
func CreateSimpleAccount(ctx context.Context, request *configurator_models.CreateSimpleAccountRequest) (*configurator_models.CreateSimpleAccountResponse, error) {
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
	// create a signing key that we can use for issuing users
	askp, err := nkeys.CreateAccount()
	if err != nil {
		log.Error().Err(err).Msg("failed to create account")
		return nil, err
	}
	// extract the public key
	aspk, err := askp.PublicKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to get public key")
		return nil, err
	}
	// create the claim for the account using the public key of the account
	ac := jwt.NewAccountClaims(apk)
	ac.Name = request.Name
	ac.Expires = time.Now().Add(time.Minute * 2).Unix()
	ac.Limits.JetStreamLimits.DiskStorage = -1
	ac.Limits.JetStreamLimits.MemoryStorage = -1

	// add the signing key (public) to the account
	ac.SigningKeys.Add(aspk)

	// now we could encode an issue the account using the operator
	// key that we generated above, but this will illustrate that
	// the account could be self-signed, and given to the operator
	// who can then re-sign it
	accountJWT, err := ac.Encode(request.IssuerKeyPair)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode account")
		return nil, err
	}

	// this user is our sentinel user that has no rights but lets us pass username/password where password can be our opaque token
	createSentinelUserCredsResponse, err := CreateUserWithCreds(ctx,
		&configurator_models.CreateUserWithCredsRequest{
			Name:          "sentinel",
			IssuerID:      apk,
			SignerKeyPair: askp,
			Permissions: jwt.Permissions{
				Pub: jwt.Permission{
					Deny: []string{">"},
				},
				Sub: jwt.Permission{
					Deny: []string{">"},
				},
			},
		})
	if err != nil {
		log.Error().Err(err).Msg("failed to create sentinel creds")
		return nil, err
	}

	resp := &configurator_models.CreateSimpleAccountResponse{
		CommonAccountData: configurator_models.CommonAccountData{
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
	resp.SentinelUser = createSentinelUserCredsResponse
	return resp, nil
}
func CreateAuthAccount(ctx context.Context, request *configurator_models.CreateAuthAccountRequest) (*configurator_models.CreateAuthAccountResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "CreateAuthAccount").Logger()
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
	// generate a creds formatted file that can be used by a NATS client

	// create the claim for the account using the public key of the account
	ac := jwt.NewAccountClaims(apk)
	ac.Name = request.Name
	// create a signing key that we can use for issuing users
	askp, err := nkeys.CreateAccount()
	if err != nil {
		log.Error().Err(err).Msg("failed to create account")
		return nil, err
	}
	// extract the public key
	aspk, err := askp.PublicKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to get public key")
		return nil, err
	}
	// this user is to allow the callout service to connect and register the micro api
	createUserCredsResponse, err := CreateUserWithCreds(ctx,
		&configurator_models.CreateUserWithCredsRequest{
			Name:          "auth",
			IssuerID:      apk,
			SignerKeyPair: askp,
		})
	if err != nil {
		log.Error().Err(err).Msg("failed to format user config")
		return nil, err
	}

	// this user is our sentinel user that has no rights but lets us pass username/password where password can be our opaque token
	createSentinelUserCredsResponse, err := CreateUserWithCreds(ctx,
		&configurator_models.CreateUserWithCredsRequest{
			Name:          "sentinel",
			IssuerID:      apk,
			SignerKeyPair: askp,
			Permissions: jwt.Permissions{
				Pub: jwt.Permission{
					Deny: []string{">"},
				},
				Sub: jwt.Permission{
					Deny: []string{">"},
				},
			},
		})
	if err != nil {
		log.Error().Err(err).Msg("failed to create sentinel creds")
		return nil, err
	}

	// add the signing key (public) to the account
	ac.SigningKeys.Add(aspk)
	// don't know about this one.
	ac.Authorization.AuthUsers.Add(createUserCredsResponse.UserKeyPair.PublicKey)
	ac.Authorization.AllowedAccounts.Add("*")

	ac.Limits.JetStreamLimits.DiskStorage = -1
	ac.Limits.JetStreamLimits.MemoryStorage = -1
	// now we could encode an issue the account using the operator
	// key that we generated above, but this will illustrate that
	// the account could be self-signed, and given to the operator
	// who can then re-sign it
	accountJWT, err := ac.Encode(request.IssuerKeyPair)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode account")
		return nil, err
	}

	resp := &configurator_models.CreateAuthAccountResponse{
		CommonAccountData: configurator_models.CommonAccountData{
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

	resp.AccountUser = createUserCredsResponse
	resp.SentinelUser = createSentinelUserCredsResponse
	return resp, nil
}

func CreateUserWithCreds(ctx context.Context, request *configurator_models.CreateUserWithCredsRequest) (*configurator_models.CreateUserWithCredsResonse, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "CreateUserWithCreds").Logger()
	// now back to the account, the account can issue users
	// need not be known to the operator - the users are trusted
	// because they will be signed by the account. The server will
	// look up the account get a list of keys the account has and
	// verify that the user was issued by one of those keys
	ukp, err := nkeys.CreateUser()
	if err != nil {
		log.Error().Err(err).Msg("failed to create user")
		return nil, err
	}

	upk, err := ukp.PublicKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to get public key")
		return nil, err
	}
	uc := jwt.NewUserClaims(upk)
	uc.Name = request.Name
	uc.Permissions = request.Permissions
	// since the jwt will be issued by a signing key, the issuer account
	// must be set to the public ID of the account
	uc.IssuerAccount = request.IssuerID
	userJwt, err := uc.Encode(request.SignerKeyPair)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode user")
		return nil, err
	}
	// the seed is a version of the keypair that is stored as text
	useed, err := ukp.Seed()
	if err != nil {
		log.Error().Err(err).Msg("failed to get seed")
		return nil, err
	}
	// generate a creds formatted file that can be used by a NATS client
	creds, err := jwt.FormatUserConfig(userJwt, useed)
	if err != nil {
		log.Error().Err(err).Msg("failed to format user config")
		return nil, err
	}
	privKey, _ := ukp.PrivateKey()
	return &configurator_models.CreateUserWithCredsResonse{
		UserKeyPair: configurator_models.RawKeyPair{
			PublicKey:  upk,
			PrivateKey: privKey,
			Seed:       useed,
		},
		JWT:   userJwt,
		Creds: creds,
	}, nil
}

/*
	{
	  "jti": "X53WBIL2MXEZZ2R6WVGKTWHZ3ENVRNTTYBLWEJJSQFQHKBQWUVGA",
	  "iat": 1737755087,
	  "iss": "OCQGZZVCJGGNR65Y72L6CNBRBHSZ7OILNQAXZLBZRAJUTDOA5JABBGWP",
	  "name": "local-callout-resolver2",
	  "sub": "OCQGZZVCJGGNR65Y72L6CNBRBHSZ7OILNQAXZLBZRAJUTDOA5JABBGWP",
	  "nats": {
	    "signing_keys": [
	      "ODRSGJVWECSLUL4ZMJITH3TBTTU3M35IY37UXMQAFUD6N7DZ5XMCDIAG"
	    ],
	    "account_server_url": "nats://localhost:4222",
	    "system_account": "ADI67YVFPNBQ3HCNK3XRCFYKD2W2MN2KDHXRDKEMNUANPMTEGL35ZULD",
	    "strict_signing_key_usage": true,
	    "type": "operator",
	    "version": 2
	  }
	}
*/
func CreateOperator(ctx context.Context, request *configurator_models.CreateOperatorRequest) (*configurator_models.CreateOperatorResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "CreateOperator").Logger()
	// create an operator key pair (private key)
	okp, err := nkeys.CreateOperator()
	if err != nil {
		log.Error().Err(err).Msg("failed to create operator")
		return nil, err
	}

	// extract the public key
	opk, err := okp.PublicKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to get public key")
		return nil, err
	}
	// create an operator claim using the public key for the identifier
	oc := jwt.NewOperatorClaims(opk)
	oc.Name = request.Name
	oc.StrictSigningKeyUsage = true
	// add an operator signing key to sign accounts
	oskp, err := nkeys.CreateOperator()
	if err != nil {
		log.Error().Err(err).Msg("failed to create operator")
		return nil, err
	}
	createSystemAccountResponse, err := CreateSystemAccount(ctx, &configurator_models.CreateSystemAccountRequest{
		CreateSimpleAccountRequest: configurator_models.CreateSimpleAccountRequest{
			Name:          "SYS",
			IssuerKeyPair: oskp,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create system account")
		return nil, err
	}
	createAuthAccountResponse, err := CreateAuthAccount(ctx, &configurator_models.CreateAuthAccountRequest{
		Name:          "AUTH",
		IssuerKeyPair: oskp,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create auth account")
		return nil, err
	}

	// get the public key for the signing key
	ospk, err := oskp.PublicKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to get public key")
		return nil, err
	}
	// add the signing key to the operator - this makes any account
	// issued by the signing key to be valid for the operator
	oc.SigningKeys.Add(ospk)
	oc.SystemAccount = createSystemAccountResponse.KeyPair.PublicKey

	// self-sign the operator JWT - the operator trusts itself
	operatorJWT, err := oc.Encode(okp)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode operator")
		return nil, err
	}
	resp := &configurator_models.CreateOperatorResponse{
		OperatorAccount: configurator_models.CommonAccountData{
			Name: request.Name,
			JWT:  operatorJWT,
		},
	}
	resp.OperatorAccount.KeyPair.PublicKey, _ = okp.PublicKey()
	resp.OperatorAccount.KeyPair.PrivateKey, _ = okp.PrivateKey()
	resp.OperatorAccount.KeyPair.Seed, _ = okp.Seed()
	resp.OperatorAccount.SignerKeyPair.PublicKey, _ = oskp.PublicKey()
	resp.OperatorAccount.SignerKeyPair.PrivateKey, _ = oskp.PrivateKey()
	resp.OperatorAccount.SignerKeyPair.Seed, _ = oskp.Seed()
	resp.SystemAccount = createSystemAccountResponse
	resp.AuthAccount = createAuthAccountResponse
	return resp, nil

}
func WriteCreds(ctx context.Context, root string, request *configurator_models.WriteCredsRequest) (*configurator_models.WriteCredsResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "WriteNatsSystemAccountCreds").Logger()
	exePath, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("failed to get executable path")
		return nil, err
	}

	exeDir := filepath.Dir(exePath)
	if err := os.MkdirAll(path.Join(exeDir, root), 0755); err != nil {
		log.Error().Err(err).Msg("failed to create directory")
		return nil, err
	}
	outputDir := path.Join(exeDir, root)
	credsPath := path.Join(outputDir, fmt.Sprintf(" %s.creds", request.Name))
	if err := os.WriteFile(credsPath,
		request.Creds, 0644); err != nil {
		log.Error().Err(err).Msg("failed to write creds")
		return nil, err

	}
	return &configurator_models.WriteCredsResponse{
		CredsFile: credsPath,
	}, nil
}
func WriteNatsSystemAccountCreds(ctx context.Context, root string, createOperatorResponse *configurator_models.CreateOperatorResponse) error {
	log := zerolog.Ctx(ctx).With().Str("func", "WriteNatsSystemAccountCreds").Logger()
	ctx = log.WithContext(ctx)
	_, err := WriteCreds(ctx, root, &configurator_models.WriteCredsRequest{
		Name:  fmt.Sprintf("%s.sys", createOperatorResponse.OperatorAccount.Name),
		Creds: createOperatorResponse.SystemAccount.AccountUser.Creds,
	})
	return err

}
func WriteNatsAuthAccountCreds(ctx context.Context, root string, createOperatorResponse *configurator_models.CreateOperatorResponse) error {
	log := zerolog.Ctx(ctx).With().Str("func", "WriteNatsAuthAccountCreds").Logger()
	ctx = log.WithContext(ctx)
	_, err := WriteCreds(ctx, root, &configurator_models.WriteCredsRequest{
		Name:  fmt.Sprintf("%s.auth", createOperatorResponse.OperatorAccount.Name),
		Creds: createOperatorResponse.AuthAccount.AccountUser.Creds,
	})
	if err != nil {
		return err
	}
	_, err = WriteCreds(ctx, root, &configurator_models.WriteCredsRequest{
		Name:  fmt.Sprintf("%s.auth.sentinel", createOperatorResponse.OperatorAccount.Name),
		Creds: createOperatorResponse.AuthAccount.SentinelUser.Creds,
	})
	return err
}

func WriteNatsServerConfig(ctx context.Context, root string, createOperatorResponse *configurator_models.CreateOperatorResponse) error {
	log := zerolog.Ctx(ctx).With().Str("func", "WriteNatsServerConfig").Logger()
	exePath, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("failed to get executable path")
		return err
	}

	exeDir := filepath.Dir(exePath)
	// create a directory for the operator account
	if err := os.MkdirAll(path.Join(exeDir, root), 0755); err != nil {
		log.Error().Err(err).Msg("failed to create directory")
		return err
	}
	outputDir := path.Join(exeDir, root)

	natsConfPath := path.Join(outputDir, fmt.Sprintf("%s.resolver.conf", createOperatorResponse.OperatorAccount.Name))
	// we are generating a memory resolver server configuration
	// it lists the operator and all account jwts the server should
	// know about
	resolver := fmt.Sprintf(
		`# Operator named %s
operator: %s
# System Account named SYS
system_account: %s

# configuration of the nats based resolver
resolver {
    type: full
    # Directory in which account jwt will be stored
    dir: './jwt'
    # In order to support jwt deletion, set to true
    # If the resolver type is full delete will rename the jwt.
    # This is to allow manual restoration in case of inadvertent deletion.
    # To restore a jwt, remove the added suffix .delete and restart or send a reload signal.
    # To free up storage you must manually delete files with the suffix .delete.
    allow_delete: false
    # Interval at which a nats-server with a nats based account resolver will compare
    # it's state with one random nats based account resolver in the cluster and if needed,
    # exchange jwt and converge on the same set of jwt.
    interval: "2m"
	# limit on the number of jwt stored, will reject new jwt once limit is hit.
    limit: 1000000
}
# Preload the nats based resolver with the system account jwt.
# This is not necessary but avoids a bootstrapping system account. 
# This only applies to the system account. Therefore other account jwt are not included here.
# To populate the resolver:
# 1) make sure that your operator has the account server URL pointing at your nats servers.
#    The url must start with: "nats://" 
#    nsc edit operator --account-jwt-server-url nats://localhost:4222
# 2) push your accounts using: nsc push --all
#    The argument to push -u is optional if your account server url is set as described.
# 3) to prune accounts use: nsc push --prune 
#    In order to enable prune you must set above allow_delete to true
# Later changes to the system account take precedence over the system account jwt listed here.
resolver_preload: {
    %s: %s
}
`,
		createOperatorResponse.OperatorAccount.Name,
		createOperatorResponse.OperatorAccount.JWT,
		createOperatorResponse.SystemAccount.KeyPair.PublicKey,
		createOperatorResponse.SystemAccount.KeyPair.PublicKey,
		createOperatorResponse.SystemAccount.JWT,
	)
	if err := os.WriteFile(natsConfPath,
		[]byte(resolver), 0644); err != nil {
		log.Error().Err(err).Msg("failed to write resolver config")
		return err

	}
	return nil

}
func WriteMasterServiceConfig(ctx context.Context, root string, createOperatorResponse *configurator_models.CreateOperatorResponse) error {
	log := zerolog.Ctx(ctx).With().Str("func", "WriteMasterServiceConfig").Logger()
	exePath, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("failed to get executable path")
		return err
	}

	exeDir := filepath.Dir(exePath)
	// create a directory for the operator account
	if err := os.MkdirAll(path.Join(exeDir, root), 0755); err != nil {
		log.Error().Err(err).Msg("failed to create directory")
		return err
	}
	outputDir := path.Join(exeDir, root)

	natsConfPath := path.Join(outputDir, fmt.Sprintf("%s.resolver.service.json", createOperatorResponse.OperatorAccount.Name))
	if err := os.WriteFile(natsConfPath,
		[]byte(fluffycore_utils.PrettyJSON(createOperatorResponse)), 0644); err != nil {
		log.Error().Err(err).Msg("failed to write resolver config")
		return err
	}
	return nil

}
