package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

func (p *KeyPair) SignPKCS1v15(data []byte) ([]byte, error) {
	// Compute SHA256 hash of the data
	h := sha256.New()
	h.Write(data)
	d := h.Sum(nil)
	// Sign the hash using RSA PKCS1v15
	return rsa.SignPKCS1v15(rand.Reader, p.PrivateKey, crypto.SHA256, d)
}

func (p *KeyPair) VerifyPKCS1v15(data, signature []byte) error {
	// Compute SHA256 hash of the data
	h := sha256.New()
	h.Write(data)
	d := h.Sum(nil)
	// Verify the signature using RSA PKCS1v15
	return rsa.VerifyPKCS1v15(&p.PrivateKey.PublicKey, crypto.SHA256, d, signature)
}
