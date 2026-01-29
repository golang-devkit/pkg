package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

const (
	RSAKeySize = 2048
)

func GenerateKeyPair() (*KeyPair, error) {
	// Generate RSA key pair
	key, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	return &KeyPair{PrivateKey: key}, nil
}

func KeyPairFromSecret(secret string) (*KeyPair, error) {
	der, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode secret: %w", err)
	}

	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key")
	}
	return &KeyPair{
		PrivateKey: privateKey,
	}, nil
}

func KeyPairFromPEM(privatePEM []byte) (*KeyPair, error) {
	block, _ := pem.Decode(privatePEM)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	var (
		keyParsed *rsa.PrivateKey
		ok        bool
		err       error
	)

	switch block.Type {
	case "PRIVATE KEY":
		// PKCS8 format
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		if keyParsed, ok = parsed.(*rsa.PrivateKey); !ok {
			return nil, fmt.Errorf("invalid private key")
		}
	case "RSA PRIVATE KEY":
		// PKCS8 format
		keyParsed, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	return &KeyPair{
		PrivateKey: keyParsed,
	}, nil
}

type KeyPair struct {
	PrivateKey *rsa.PrivateKey
}

func (p *KeyPair) PKIXPublicKey() (string, error) {
	if p.PrivateKey == nil {
		return "", fmt.Errorf("private key is nil")
	}
	pubKey := p.PrivateKey.Public()
	bin, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bin), nil
}

func (p *KeyPair) PKCS8PrivateKey() (string, error) {
	bin, err := x509.MarshalPKCS8PrivateKey(p.PrivateKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bin), nil
}

func (p *KeyPair) PEM() (public []byte, private []byte, err error) {
	// PKIX marshal public key
	pkix, err := x509.MarshalPKIXPublicKey(p.PrivateKey.Public())
	if err != nil {
		return nil, nil, err
	}
	// Encode result
	o1 := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkix,
	})
	// PKCS8 marshal private key
	bin, err := x509.MarshalPKCS8PrivateKey(p.PrivateKey)
	if err != nil {
		return nil, nil, err
	}
	o2 := pem.EncodeToMemory(&pem.Block{
		// This kind of key is commonly encoded in PEM blocks of type "RSA PRIVATE KEY".
		// PKCS8 kind of key is commonly encoded in PEM blocks of type "PRIVATE KEY".
		Type:  "PRIVATE KEY",
		Bytes: bin,
	})
	return o1, o2, nil
}
