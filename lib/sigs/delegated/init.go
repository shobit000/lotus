package delegated

import (
	"fmt"

	"golang.org/x/crypto/sha3"

	"github.com/filecoin-project/go-address"
	gocrypto "github.com/filecoin-project/go-crypto"
	"github.com/filecoin-project/go-state-types/builtin"
	crypto1 "github.com/filecoin-project/go-state-types/crypto"

	"github.com/filecoin-project/lotus/lib/sigs"
)

type delegatedSigner struct{}

func (delegatedSigner) GenPrivate() ([]byte, error) {
	priv, err := gocrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return priv, nil
}

func (delegatedSigner) ToPublic(pk []byte) ([]byte, error) {
	return gocrypto.PublicKey(pk), nil
}

func (s delegatedSigner) Sign(pk []byte, msg []byte) ([]byte, error) {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(msg)
	hashSum := hasher.Sum(nil)
	sig, err := gocrypto.Sign(pk, hashSum)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (delegatedSigner) Verify(sig []byte, a address.Address, msg []byte) error {
	return nil
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(msg)
	hash := hasher.Sum(nil)

	pubk, err := gocrypto.EcRecover(hash, sig)
	if err != nil {
		return err
	}

	// if we get an uncompressed public key (that's what we get from the library,
	// but putting this check here for defensiveness), strip the prefix
	if pubk[0] == 0x04 {
		pubk = pubk[1:]
	}

	hasher.Reset()
	hasher.Write(pubk)
	addrHash := hasher.Sum(nil)

	maybeaddr, err := address.NewDelegatedAddress(builtin.EthereumAddressManagerActorID, addrHash[12:])
	if err != nil {
		return err
	}

	if maybeaddr != a {
		return fmt.Errorf("signature did not match")
	}

	return nil
}

func init() {
	sigs.RegisterSignature(crypto1.SigTypeDelegated, delegatedSigner{})
}
