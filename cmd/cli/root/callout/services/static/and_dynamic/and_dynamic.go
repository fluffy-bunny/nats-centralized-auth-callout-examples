package and_dynamic

import (
	"context"
	"fmt"
	"math/rand"
	cobra_utils "natsauth/internal/cobra_utils"
	shared "natsauth/internal/shared"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	callout "github.com/aricart/callout.go"
	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	status "github.com/gogo/status"
	jwt "github.com/nats-io/jwt/v2"
	nats "github.com/nats-io/nats.go"
	nkeys "github.com/nats-io/nkeys"
	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
	codes "google.golang.org/grpc/codes"
)

const use = "and_dynamic"

var (
	appInputs                 = shared.NewInputs()
	users              string = "./configs/users.json"
	wellknownAudiences        = map[string]string{
		"SYS":  "SYS",
		"Auth": "Auth",
	}
)

// Init command
func Init(parentCmd *cobra.Command) {
	var command = &cobra.Command{
		Use:               use,
		Short:             use,
		PersistentPreRunE: cobra_utils.ParentPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := shared.GetContext()
			log := zerolog.Ctx(ctx).With().Str("command", use).Logger()

			printer := cobra_utils.NewPrinter()
			printer.EnableColors = true
			printer.PrintBold(cobra_utils.Bold, use)

			// Parse the issuer account signing key.
			issuerKeyPair, err := nkeys.FromSeed([]byte(appInputs.IssuerSeed))
			if err != nil {
				log.Error().Err(err).Msg("error parsing issuer seed")
				return status.Error(codes.Internal, "error parsing issuer seed")
			}

			accountMap := sync.Map{}
			accountNameToPublicKeyMap := sync.Map{}

			accountMap.Store("SYS", &CreateSimpleAccountResponse{
				CommonAccountData: CommonAccountData{
					Name:     "SYS",
					Audience: "SYS",
					KeyPair: RawKeyPair{
						PublicKey: "SYS",
					},
				},
			})
			accountNameToPublicKeyMap.Store("SYS", "SYS")

			accountLock := sync.Mutex{}
			tryFetchFriendlyAccountPublicKey := func(accountPublicKey string) (string, bool) {
				accountLock.Lock()
				defer accountLock.Unlock()
				val, ok := accountNameToPublicKeyMap.Load(accountPublicKey)
				if ok {
					return val.(string), true
				}
				return "", false
			}
			getOrCreateAccount := func(friendlyAccountName string) (*CreateSimpleAccountResponse, error) {
				accountLock.Lock()
				defer accountLock.Unlock()
				var err error
				val, ok := accountMap.Load(friendlyAccountName)
				if ok {
					resp := val.(*CreateSimpleAccountResponse)
					_, ok = wellknownAudiences[resp.Audience]
					if !ok {

						resp, err = UpdateSimpleAccount(ctx, &UpdateSimpleAccountRequest{
							Original:      resp,
							IssuerKeyPair: issuerKeyPair,
						})
						if err != nil {
							return nil, err
						}
					}
					accountMap.Store(friendlyAccountName, resp)
					accountNameToPublicKeyMap.Store(resp.KeyPair.PublicKey, friendlyAccountName)
					return resp, nil
				}
				resp, err := CreateSimpleAccount(ctx,
					&CreateSimpleAccountRequest{
						Name:          friendlyAccountName,
						IssuerKeyPair: issuerKeyPair,
					})
				if err != nil {
					return nil, err
				}
				resp.Audience = resp.KeyPair.PublicKey
				accountMap.Store(friendlyAccountName, resp)
				accountNameToPublicKeyMap.Store(resp.KeyPair.PublicKey, friendlyAccountName)
				return resp, nil
			}

			// Get the connection as the auth user
			nc, err := appInputs.MakeConn(ctx)
			if err != nil {
				log.Error().Err(err).Msg("failed to connect to nats server")
				return err
			}
			defer nc.Drain()
			printer.Infof("%s connected to %s", appInputs.NatsUser, nc.ConnectedUrl())

			// parse the private key
			akp, err := nkeys.FromSeed([]byte(appInputs.IssuerSeed))
			if err != nil {
				log.Error().Err(err).Msg("error parsing issuer seed")
				return err
			}
			akpPublickKey, _ := akp.PublicKey()
			log.Info().Str("issuer", akpPublickKey).Msg("issuer")

			if !shared.FileExists(users) {
				log.Error().Str("file", users).Msg("file does not exist")
				return status.Error(codes.Internal, "file does not exist")
			}
			usersData, err := shared.LoadUsersData(users)
			if err != nil {
				log.Error().Err(err).Msg("error loading users data")
				return status.Error(codes.Internal, "error loading users data")
			}

			sysAccount := "sys"
			sysUser := shared.User{
				Username: sysAccount,
				Password: sysAccount, // in the real world this would be a configed password hash, or argon2
				Sub: shared.Permissions{
					Allow: []string{">"},
				},
				Pub: shared.Permissions{
					Allow: []string{">"},
				},
				AllowedAccounts: []string{"SYS"},
			}
			usersData.Users = append(usersData.Users, sysUser)

			printer.Println(cobra_utils.Blue, fluffycore_utils.PrettyJSON(usersData))
			// Parse the xkey seed if present.
			var curveKeyPair nkeys.KeyPair
			if fluffycore_utils.IsNotEmptyOrNil(appInputs.XKeySeed) {
				curveKeyPair, err = nkeys.FromSeed([]byte(appInputs.XKeySeed))
				if err != nil {
					log.Error().Err(err).Msg("error parsing xkey seed")
					return status.Error(codes.Internal, "error parsing xkey seed")
				}
			}
			if curveKeyPair != nil {
				curveKeyPairPublicKey, _ := curveKeyPair.PublicKey()
				log.Info().Str("xkey", curveKeyPairPublicKey).Msg("xkey")
			}
			// a function that creates the users
			authorizer := func(req *jwt.AuthorizationRequest) (string, error) {
				// peek at the req for information - for brevity
				// in the example, we simply allow them in
				log.Info().Str("user", req.UserNkey).Msg("authorizing")

				username := req.ConnectOptions.Username
				password := req.ConnectOptions.Password

				userParts := strings.Split(username, "@")
				username = userParts[0]
				var account string
				if len(userParts) > 1 {
					account = userParts[1]
				}
				printer.Printf(cobra_utils.Blue, "username: %s, password: %s\n", username, password)
				authenticated := false
				var user *shared.User

				for _, item := range usersData.Users {
					if item.Username == username && item.Password == password {
						authenticated = true
						user = &item
						break
					}
				}
				if user == nil {
					printer.Printf(cobra_utils.Red, "UNAUTHORIZED: username: %s, password: %s\n, account: %s", username, password, account)
					return "", status.Error(codes.PermissionDenied, "permission denied")
				}
				// check if account is allowed.
				getRequestedAccount := func(account string) (string, bool) {
					if fluffycore_utils.IsEmptyOrNil(account) {
						if fluffycore_utils.IsEmptyOrNil(user.AllowedAccounts) {
							return "", false
						}
						first := user.AllowedAccounts[0]
						if first == "*" { // not allowed
							return "", false
						}
						return first, true
					}
					for _, item := range user.AllowedAccounts {
						if item == account || item == "*" {
							return account, true
						}
					}
					return "", false
				}
				account, ok := getRequestedAccount(account)
				if !ok {
					printer.Printf(cobra_utils.Red, "UNAUTHORIZED: account: %s\n", account)
					return "", status.Error(codes.PermissionDenied, "permission denied")
				}
				myAccount, err := getOrCreateAccount(account)
				if err != nil {
					log.Error().Err(err).Msg("error getting account")
					return "", status.Error(codes.Internal, "error getting account")
				}
				audience := myAccount.Audience
				if !authenticated {
					printer.Printf(cobra_utils.Red, "UNAUTHORIZED: username: %s, password: %s\n", username, password)
					return "", status.Error(codes.PermissionDenied, "permission denied")
				}
				printer.Println(cobra_utils.Blue, fluffycore_utils.PrettyJSON(myAccount))
				printer.Println(cobra_utils.Blue, fluffycore_utils.PrettyJSON(user))

				// use the server specified user nkey
				uc := jwt.NewUserClaims(req.UserNkey)
				// put the user in the global account
				uc.Audience = audience
				// add whatever permissions you need
				uc.Sub.Allow.Add(user.Sub.Allow...)
				uc.Pub.Allow.Add(user.Pub.Allow...)

				uc.Sub.Deny.Add(user.Sub.Deny...)
				uc.Pub.Deny.Add(user.Pub.Deny...)

				// perhaps add an expiration to the JWT
				uc.Expires = time.Now().Unix() + 90
				return uc.Encode(akp)
			}
			// start the service
			_, err = callout.NewAuthorizationService(nc, callout.Authorizer(authorizer), callout.ResponseSignerKey(akp))
			if err != nil {
				log.Error().Err(err).Msg("error starting service")
				return err
			}

			ncSys, err := nats.Connect(appInputs.NatsUrl,
				nats.UserInfo(sysAccount, sysAccount))
			if err != nil {
				log.Error().Err(err).Msg("error connecting to NATS")
				return err
			}
			defer func() {
				defer ncSys.Drain()
			}()
			sub, err := ncSys.Subscribe("$SYS.REQ.ACCOUNT.*.CLAIMS.LOOKUP", func(msg *nats.Msg) {
				accountId := strings.TrimSuffix(strings.TrimPrefix(msg.Subject, "$SYS.REQ.ACCOUNT."), ".CLAIMS.LOOKUP")

				friendlyName, ok := tryFetchFriendlyAccountPublicKey(accountId)
				if !ok {
					log.Error().Msgf("account not found: %s", accountId)
					return
				}
				createSimpleAccountResponse, err := getOrCreateAccount(friendlyName)
				if err != nil {
					log.Error().Err(err).Msg("error getting account")
					return
				}
				createSimpleAccountResponse, err = UpdateSimpleAccount(ctx,
					&UpdateSimpleAccountRequest{
						Original:      createSimpleAccountResponse,
						IssuerKeyPair: issuerKeyPair,
					})
				if err != nil {
					log.Error().Err(err).Msg("error getting account")
					return
				}
				jwt := createSimpleAccountResponse.JWT
				err = msg.Respond([]byte(jwt))
				if err != nil {
					log.Error().Err(err).Msg("error responding")
				}
				log.Info().Str("accountJWT", jwt).Msg("accountJWT")

			})
			if err != nil {
				log.Error().Err(err).Msg("error connecting to NATS")
				return err
			}
			defer func() {
				sub.Unsubscribe()
			}()
			// don't exit until sigterm
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
			<-quit
			return nil

		},
	}
	appInputs.NatsUser = "auth"
	appInputs.NatsPass = "auth"
	appInputs.IssuerSeed = "SAAEXFSYMLINXLKR2TG5FLHCJHLU62B3SK3ESZLGP4B4XGLUNXICW3LGAY"

	shared.InitCommonConnFlags(appInputs, command)

	flagName := "issuer.seed"
	defaultS := appInputs.IssuerSeed
	command.Flags().StringVar(&appInputs.IssuerSeed, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "users.file"
	defaultS = users
	command.Flags().StringVar(&users, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	parentCmd.AddCommand(command)

}

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

func randomInt64(min, max int64) int64 {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	// Generate a random number in the range [min, max]
	return min + rand.Int63n(max-min+1)
}

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
	return resp, nil
}

// a POC for updating an account
func UpdateSimpleAccount(ctx context.Context, request *UpdateSimpleAccountRequest) (*CreateSimpleAccountResponse, error) {
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
