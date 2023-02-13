package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"log"
) //eliptical curve digital signing algorithm

// "golang.org/x/crypto/ripemd160"
const (
	ChecksumLength = 4
	version        = byte(0x00) //hexdecimal representation of 0
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	fmt.Printf("Versioned hash: %x\n", versionedHash)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	fmt.Printf("pub key: %x\n", w.PublicKey)
	fmt.Printf("pub hash: %x\n", pubHash)
	fmt.Printf("address: %s\n", address)

	return address

}

func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))                   //versioned hash + checksum
	actualChecksum := pubKeyHash[len(pubKeyHash)-ChecksumLength:] //dobije se checksum
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-ChecksumLength] //public key hash
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256() //output of the elliptic curve would be 256 bytes

	private, err := ecdsa.GenerateKey(curve.Params(), rand.Reader) //rand.Reader generates our private key
	if err != nil {
		print("ge")
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pub
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{
		PrivateKey: private,
		PublicKey:  public,
	}

	return &wallet
}

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil) //we don't need to concatanate any value
	fmt.Printf("Full hash %x\n", publicRipMD)

	return publicRipMD

}

func Checksum(payload []byte) []byte {
	fisrtHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(fisrtHash[:])
	fmt.Printf("full checksum %x\n", secondHash)
	fmt.Printf("part checksum %x\n", secondHash[:ChecksumLength])

	return secondHash[:ChecksumLength]
}
