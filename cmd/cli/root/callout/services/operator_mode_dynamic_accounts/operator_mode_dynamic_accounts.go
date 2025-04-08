package operator_mode_dynamic_accounts

import (
	"bytes"
	"errors"
	"fmt"
	cobra_utils "natsauth/internal/cobra_utils"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	contracts_nats "natsauth/internal/contracts/nats"
	shared "natsauth/internal/shared"

	callout_shared "natsauth/cmd/cli/root/callout/shared"

	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	status "github.com/gogo/status"
	jwt "github.com/nats-io/jwt/v2"
	nats "github.com/nats-io/nats.go"
	nkeys "github.com/nats-io/nkeys"
	zerolog "github.com/rs/zerolog"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
	callout "github.com/synadia-io/callout.go"
	codes "google.golang.org/grpc/codes"
)

const use = "operator_mode_dynamic_accounts"

var (
	appInputs    = shared.NewInputs()
	users        = "./configs/users.json"
	accountMutex = sync.Mutex{}
)

type (
	AccountUser struct {
		AccountName  string `json:"accountName"`
		UserName     string `json:"userName"`
		UserPassword string `json:"userPassword"`
	}
)

// will be in a persistent store
var accountFriendlyNameToAccountInfo = make(map[string]*contracts_nats.CreateSimpleAccountResponse)
var accountPubKeyToAccountInfo = make(map[string]*contracts_nats.CreateSimpleAccountResponse)

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

			if !shared.FileExists(users) {
				log.Error().Str("file", users).Msg("file does not exist")
				return status.Error(codes.Internal, "file does not exist")
			}
			usersData, err := shared.LoadUsersData(users)
			if err != nil {
				log.Error().Err(err).Msg("error loading users data")
				return status.Error(codes.Internal, "error loading users data")
			}
			for idx := range usersData.Users {
				usersData.Users[idx].Username = strings.ToLower(usersData.Users[idx].Username)
				for aid := range usersData.Users[idx].AllowedAccounts {
					usersData.Users[idx].AllowedAccounts[aid] = strings.ToLower(usersData.Users[idx].AllowedAccounts[aid])
				}
			}
			printer.Println(cobra_utils.Blue, fluffycore_utils.PrettyJSON(usersData))

			authAccountJWT, err := os.ReadFile(appInputs.AuthAccountJWTFile)
			if err != nil {
				return err
			}
			subI, err := shared.ExtractClaimFromJWTNoValidation(string(authAccountJWT), "sub")
			if err != nil {
				log.Error().Err(err).Msg("error extracting sub from jwt")
				return err
			}
			authAudience := subI.(string)
			accountFriendlyNameToAccountInfo["AUTH"] = &contracts_nats.CreateSimpleAccountResponse{
				CommonAccountData: contracts_nats.CommonAccountData{
					Name:     "AUTH",
					JWT:      string(authAccountJWT),
					Audience: authAudience,
				},
			}
			accountPubKeyToAccountInfo[authAudience] = accountFriendlyNameToAccountInfo["AUTH"]

			okp, err := loadAndParseKeys(appInputs.OperatorNKeyFile, 'O')
			if err != nil {
				log.Error().Err(err).Msg("error loading operator key")
				return err
			}

			sysOpts, err := getConnectionOptions(appInputs.SysCreds)
			if err != nil {
				log.Error().Err(err).Msg("error loading creds")
				return err
			}

			sysConn, err := nats.Connect(appInputs.NatsUrl, sysOpts...)
			if err != nil {
				log.Error().Err(err).Msg("error connecting to nats")
				return err
			}
			_, err = sysConn.Subscribe("$SYS.REQ.ACCOUNT.*.CLAIMS.LOOKUP",
				func(m *nats.Msg) {
					chunks := strings.Split(m.Subject, ".")
					id := chunks[3]
					fmt.Println(id)
					accountInfo, ok := accountPubKeyToAccountInfo[id]
					if ok {
						err = m.Respond([]byte(accountInfo.JWT))
						if err != nil {
							log.Error().Err(err).Msg("error responding")
						}
						fmt.Println(accountInfo.JWT)

					} else {
						_ = m.Respond(nil)
					}

				})
			if err != nil {
				log.Error().Err(err).Msg("error subscribing to claims lookup")
				return err
			}

			// this creates a new account named as specified returning
			// the key used to sign users
			getOrCreateAccount := func(name string) (*contracts_nats.CreateSimpleAccountResponse, error) {
				//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
				accountMutex.Lock()
				defer accountMutex.Unlock()
				//--~--~--~--~--~-- BARBED WIRE --~--~--~--~--~--~--
				createSimpleAccountResponse, ok := accountFriendlyNameToAccountInfo[name]
				if ok {
					return createSimpleAccountResponse, nil
				}
				var err error
				createSimpleAccountResponse, err = callout_shared.CreateSimpleAccount(ctx,
					&contracts_nats.CreateSimpleAccountRequest{
						Name:          name,
						IssuerKeyPair: okp,
					})
				if err != nil {
					return nil, err
				}
				accountFriendlyNameToAccountInfo[name] = createSimpleAccountResponse
				accountPubKeyToAccountInfo[createSimpleAccountResponse.Audience] = createSimpleAccountResponse

				return createSimpleAccountResponse, err
			}
			// load the callout key
			cKP, err := loadAndParseKeys(appInputs.CalloutIssuerNKeyFile, 'A')
			if err != nil {
				panic(fmt.Errorf("error loading callout issuer: %w", err))
			}

			// the authorizer function
			authorizer := func(req *jwt.AuthorizationRequest) (string, error) {
				// reading the account name from the token, likely this will be
				// encoded string with more information

				username := req.ConnectOptions.Username
				password := req.ConnectOptions.Password

				userParts := strings.Split(username, "@")
				username = userParts[0]
				var account string
				if len(userParts) > 1 {
					account = userParts[1]
				}
				accountUser := &AccountUser{
					AccountName:  strings.ToLower(account),
					UserName:     strings.ToLower(username),
					UserPassword: password,
				}
				log := log.With().Interface("accountUser", accountUser).Logger()
				isUserAllowed := func() (*shared.User, error) {
					for _, user := range usersData.Users {
						if user.Username == username {
							for _, account := range user.AllowedAccounts {
								if account == "*" {
									return user, nil
								}
								if account == accountUser.AccountName {
									return user, nil
								}
							}
						}
					}
					return nil, status.Error(codes.Unauthenticated, "user not found or not allowed")
				}
				user, err := isUserAllowed()
				if err != nil {
					log.Error().Err(err).Msg("user not allowed")
					return "", err
				}

				// see if we have this account
				createSimpleAccountResponse, err := getOrCreateAccount(accountUser.AccountName)
				if err != nil {
					return "", err
				}

				// issue the user
				uc := jwt.NewUserClaims(req.UserNkey)
				// put the user in the global account
				uc.Audience = createSimpleAccountResponse.Audience
				// add whatever permissions you need
				uc.Sub.Allow.Add(user.Sub.Allow...)
				uc.Pub.Allow.Add(user.Pub.Allow...)

				uc.Sub.Deny.Add(user.Sub.Deny...)
				uc.Pub.Deny.Add(user.Pub.Deny...)

				// perhaps add an expiration to the JWT
				uc.Expires = time.Now().Unix() + 90
				kp, err := nkeys.FromSeed(createSimpleAccountResponse.KeyPair.Seed)
				if err != nil {
					return "", err
				}
				return uc.Encode(kp)
			}
			// connect the service with the creds
			opts, err := getConnectionOptions(appInputs.CalloutCreds)
			if err != nil {
				log.Error().Err(err).Msg("error loading creds")
				return err
			}
			nc, err := nats.Connect(appInputs.NatsUrl, opts...)
			if err != nil {
				log.Error().Err(err).Msg("error connecting")
				return err
			}
			defer nc.Close()
			// start the service
			_, err = callout.NewAuthorizationService(nc, callout.Authorizer(authorizer), callout.ResponseSignerKey(cKP))

			// don't exit until sigterm
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
			<-quit
			return nil
		},
	}
	shared.InitCommonConnFlags(appInputs, command)

	appInputs.CalloutCreds = "./herb/herb.auth.creds"
	appInputs.SysCreds = "./sys.creds"
	flagName := "users.file"
	defaultS := users
	command.Flags().StringVar(&users, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "sys.creds"
	defaultS = appInputs.SysCreds
	command.Flags().StringVar(&appInputs.SysCreds, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "auth.account.jwt"
	defaultS = "auth.account.jwt"
	command.Flags().StringVar(&appInputs.AuthAccountJWTFile, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "operator.nk"
	defaultS = "operator.nk"
	command.Flags().StringVar(&appInputs.OperatorNKeyFile, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "callout.issuer.nk"
	defaultS = "C.nk"
	command.Flags().StringVar(&appInputs.CalloutIssuerNKeyFile, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	flagName = "callout.creds"
	defaultS = "service.creds"
	command.Flags().StringVar(&appInputs.CalloutCreds, flagName, defaultS, fmt.Sprintf("[required] i.e. --%s=%s", flagName, defaultS))
	viper.BindPFlag(flagName, command.PersistentFlags().Lookup(flagName))

	parentCmd.AddCommand(command)

}

func loadAndParseKeys(fp string, kind byte) (nkeys.KeyPair, error) {
	if fp == "" {
		return nil, errors.New("key file required")
	}
	seed, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("error reading key file: %w", err)
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

func getConnectionOptions(fp string) ([]nats.Option, error) {
	if fp == "" {
		return nil, errors.New("creds file required")
	}
	return []nats.Option{nats.UserCredentials(fp)}, nil
}
