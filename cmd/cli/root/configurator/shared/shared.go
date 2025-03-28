package shared

import (
	"fmt"
	"os"

	"github.com/nats-io/nkeys"
)

type (
	JWTParams struct {
		IssuerSeed string
	}
)

var (
	CurrentJWTParams = &JWTParams{}
)

// fromSeedOrFile will attempt to load a keypair from a seed file or
// or a literal seed string.
func FromSeedOrFile(seedOrFile string) (nkeys.KeyPair, error) {
	contents, err := os.ReadFile(seedOrFile)
	if err != nil {
		if os.IsNotExist(err) {
			contents = []byte(seedOrFile)
		} else {
			return nil, fmt.Errorf("failed to read file: %s", err)
		}
	}

	return nkeys.FromSeed(contents)
}
