package gossip_common

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"golang.org/x/crypto/argon2"
)

var (
	entity *openpgp.Entity
)

/**
 * GenerateKeys generates an elliptic curve public/private key pair.
 */
func GenerateKeys() {
	var err error
	// Generate a new entity with default settings, which includes ECC keys
	entity, err = openpgp.NewEntity("gossip", "tH1s1$n0tAs3CuR3p4s$w0rD", "gossip@gossip.io", nil)
	if err != nil {
		Err("Failed to generate keys: %v", err)
		return
	}
}

/**
 * RetrievePublicKey retrieves the generated public key.
 * @return The public key.
 */
func RetrievePublicKey() []byte {
	buf := bytes.NewBuffer(nil)
	w, err := armor.Encode(buf, openpgp.PublicKeyType, nil)
	if err != nil {
		Err("Failed to encode public key: %v", err)
		return nil
	}
	entity.Serialize(w)
	w.Close()
	return buf.Bytes()
}

/**
 * RetrievePrivateKey retrieves the generated private key.
 * @return The private key.
 */
func RetrievePrivateKey() []byte {
	buf := bytes.NewBuffer(nil)
	w, err := armor.Encode(buf, openpgp.PrivateKeyType, nil)
	if err != nil {
		Err("Failed to encode private key: %v", err)
		return nil
	}
	entity.SerializePrivate(w, nil)
	w.Close()
	return buf.Bytes()
}

/**
 * GWEncrypt encrypts a message for a recipient using their public key.
 * @param plaintext The plaintext message to encrypt.
 * @param recipientPublicKey The recipient's public key.
 * @return The encrypted message.
 */
func GWEncrypt(plaintext []byte, recipientPublicKey []byte) ([]byte, error) {
	// Decode the recipient's public key
	recipientEntityList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(recipientPublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to read recipient public key: %w", err)
	}

	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, recipientEntityList, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}
	_, err = w.Write(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to write plaintext to encrypted message: %w", err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close WriteCloser: %w", err)
	}
	return ioutil.ReadAll(buf)
}

/**
 * GWEncryptToMultiple encrypts a message for multiple recipients using their public keys.
 * @param plaintext The plaintext message to encrypt.
 * @param recipientPublicKeys A slice of recipient public keys.
 * @return The encrypted message.
 */
func GWEncryptToMultiple(plaintext []byte, recipientPublicKeys [][]byte) ([]byte, error) {
	// Convert all recipient public keys from bytes to entities
	var recipientEntities []*openpgp.Entity
	for _, pk := range recipientPublicKeys {
		entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(pk))
		if err != nil {
			return nil, fmt.Errorf("failed to read recipient public key: %w", err)
		}
		recipientEntities = append(recipientEntities, entityList...)
	}

	// Encrypt the message for all recipients
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, recipientEntities, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}
	_, err = w.Write(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to write plaintext to encrypted message: %w", err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close WriteCloser: %w", err)
	}

	// Return the encrypted message
	return ioutil.ReadAll(buf)
}

/**
 * GWDecrypt decrypts an encrypted message.
 * @param ciphertext The encrypted message to decrypt.
 * @return The decrypted message.
 */
func GWDecrypt(ciphertext []byte) ([]byte, error) {
	md, err := openpgp.ReadMessage(bytes.NewBuffer(ciphertext), openpgp.EntityList([]*openpgp.Entity{entity}), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted message: %w", err)
	}
	plaintext, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read plaintext from encrypted message: %w", err)
	}
	return plaintext, nil
}

/**
 * HashPassword hashes a password using Argon2.
 * @param password The password to hash.
 * @return The hashed password.
 */
func HashPassword(password string) string {
	salt := []byte("Th1S1$nOt4sEcuR3sALt") // In a real application, use a random, unique salt for each password.
	hP := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%x", hP)
}
