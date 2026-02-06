package certificates

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	crp "github.com/kubex-ecosystem/gnyx/internal/services/security/crypto"
	krs "github.com/kubex-ecosystem/gnyx/internal/services/security/external"
	sci "github.com/kubex-ecosystem/gnyx/internal/services/security/interfaces"
	gl "github.com/kubex-ecosystem/logz"
	"golang.org/x/crypto/chacha20poly1305"
)

// CertService provides methods for managing certificates and private keys.
// It supports generating, encrypting, decrypting, and verifying certificates.
type CertService struct {
	keyPath  string             // Path to the private key file.
	certPath string             // Path to the certificate file.
	security *crp.CryptoService // Service for cryptographic operations.
}

// ensureCrypto lazily initializes the crypto service to avoid nil dereferences.
func (c *CertService) ensureCrypto() {
	if c.security == nil {
		c.security = crp.NewCryptoService()
	}
}

// GenerateCertificate generates a self-signed certificate and encrypts the private key.
// Parameters:
// - certPath: Path to save the certificate.
// - keyPath: Path to save the private key.
// - password: Password used to encrypt the private key.
// Returns: The encrypted private key, the certificate bytes, and an error if any.
func (c *CertService) GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, []byte, error) {
	c.ensureCrypto()

	if certPath == "" {
		certPath = os.ExpandEnv(kbx.DefaultGNyxCertPath)
	} else {
		certPath = os.ExpandEnv(certPath)
	}
	if keyPath == "" {
		keyPath = os.ExpandEnv(kbx.DefaultGNyxKeyPath)
	} else {
		keyPath = os.ExpandEnv(keyPath)
	}

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return nil, nil, nil, gl.Errorf("error creating directory for certificate file: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0755); err != nil {
		return nil, nil, nil, gl.Errorf("error creating directory for key file: %v", err)
	}
	var certPathTmp string
	if strings.HasSuffix(certPath, ".crt") || strings.HasSuffix(certPath, ".pem") {
		certPathTmp = certPath
	} else {
		certPathTmp = certPath + ".crt"
	}
	var keyPathTmp string
	if strings.HasSuffix(keyPath, ".key") || strings.HasSuffix(keyPath, ".pem") {
		keyPathTmp = keyPath
	} else {
		keyPathTmp = keyPath + ".key"
	}

	priv, generateKeyErr := rsa.GenerateKey(rand.Reader, 4096)
	if generateKeyErr != nil {
		return nil, nil, nil, gl.Errorf("error generating private key: %v", generateKeyErr)
	}

	sn, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128)) // Serial number
	template := x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: "Kubex Self-Signed"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         true,
		// BasicConstraintsValid: true,
		// SubjectKeyId:          []byte{1, 2, 3, 4, 6},
		// DNSNames:              []string{"localhost", "gnyx.local"},
	}

	certDER, certDERErr := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if certDERErr != nil {
		return nil, nil, nil, gl.Errorf("error creating certificate: %v", certDERErr)
	}
	strPassword, err := c.GetPrivPwd(password)
	if err != nil {
		return nil, nil, nil, gl.Errorf("error getting certificate password: %v", err)
	}
	validStrPassword, ok := c.security.GetKeyIfValid([]byte(strPassword))
	if !ok {
		decodedPasswordBytes, err := c.security.DecodeIfEncoded([]byte(strPassword))
		if err != nil {
			return nil, nil, nil, gl.Errorf("error decoding password: %v", err)
		}
		validStrPassword, ok = c.security.GetKeyIfValid(decodedPasswordBytes)
		if !ok {
			return nil, nil, nil, gl.Error("invalid password for private key encryption")
		}
	}
	decodedPasswordBytes := bytes.TrimSpace([]byte(validStrPassword))
	block, err := chacha20poly1305.NewX(decodedPasswordBytes)
	if err != nil {
		return nil, nil, nil, gl.Errorf("error creating cipher: %v, %d", err, len(decodedPasswordBytes))
	}

	nonce := make([]byte, block.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, nil, gl.Errorf("error generating nonce: %v", err)
	}

	pkcs1PrivBytes := x509.MarshalPKCS1PrivateKey(priv)
	ciphertext := block.Seal(nonce, nonce, pkcs1PrivBytes, nil)

	certPemBlock := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEMBytes := pem.EncodeToMemory(&certPemBlock)
	if len(certPEMBytes) == 0 {
		return nil, nil, nil, gl.Error("error encoding certificate: empty buffer")
	}

	copyCertDER := make([]byte, len(certDER))
	copy(copyCertDER, certDER)

	copyPkcs1PrivBytes := make([]byte, len(ciphertext))
	copy(copyPkcs1PrivBytes, ciphertext)

	privPemBlock := pem.Block{Type: "RSA PRIVATE KEY", Bytes: ciphertext}
	keyPEMBytes := pem.EncodeToMemory(&privPemBlock)
	if len(keyPEMBytes) == 0 {
		return nil, nil, nil, gl.Error("error encoding private key: empty buffer")
	}

	if err := os.WriteFile(certPathTmp, certPEMBytes, 0644); err != nil {
		return nil, nil, nil, gl.Errorf("error writing certificate file: %v", err)
	}
	if certPath != certPathTmp {
		if err := os.WriteFile(certPath, certPEMBytes, 0644); err != nil {
			return nil, nil, nil, gl.Errorf("error writing certificate file: %v", err)
		}
	}

	if err := os.WriteFile(keyPathTmp, keyPEMBytes, 0600); err != nil {
		return nil, nil, nil, gl.Errorf("error writing key file: %v", err)
	}
	if keyPath != keyPathTmp {
		if err := os.WriteFile(keyPath, keyPEMBytes, 0600); err != nil {
			return nil, nil, nil, gl.Errorf("error writing key file: %v", err)
		}
	}

	_, encryptedKeyEncoded, err := c.security.Encrypt(copyPkcs1PrivBytes, decodedPasswordBytes)
	if err != nil {
		return nil, nil, nil, gl.Errorf("error encrypting private key: %v", err)
	}

	return []byte(encryptedKeyEncoded), copyCertDER, decodedPasswordBytes, nil
}

func (c *CertService) GetPrivPwd(password []byte) (string, error) {
	c.ensureCrypto()
	// Using FileKeyring instead of DBUS-based keyring //kbx.KeyService
	KeyService := krs.NewFileKeyServiceType(
		"CertSvc", fmt.Sprintf("gnyx, "+"%s", "jwt_secret"),
	)
	if KeyService == nil {
		return "", gl.Error("file keyring service is nil")
	}
	var strPassword string
	var passwordErr error
	if password == nil {
		strPassword, passwordErr = KeyService.RetrieveOrCreatePassword()
		if passwordErr != nil && !os.IsNotExist(passwordErr) {
			return "", gl.Errorf("error retrieving password: %v", passwordErr)
		}
		if len(strPassword) == 0 {
			bytesPassword, passwordErr := c.security.GenerateKeyWithLength(32)
			if passwordErr != nil {
				return "", gl.Errorf("error generating password: %v", passwordErr)
			}
			passwordErr = KeyService.StorePassword(string(bytesPassword))
			if passwordErr != nil {
				return "", gl.Errorf("error storing password: %v", passwordErr)
			}
			strPassword = string(bytesPassword)
		}
	} else {
		strPassword = string(password)
		currentPassword, passwordErr := KeyService.RetrievePassword()
		if passwordErr != nil && !os.IsNotExist(passwordErr) {
			return "", gl.Errorf("error retrieving password: %v", passwordErr)
		}
		if currentPassword == strPassword {
			return strPassword, nil
		}
		passwordErr = KeyService.StorePassword(string(password))
		if passwordErr != nil {
			return "", gl.Errorf("error storing password: %v", passwordErr)
		}
	}
	return strPassword, nil
}

// GenSelfCert generates a self-signed certificate and stores it in the configured paths.
// Returns: The encrypted private key, the certificate bytes, and an error if any.
func (c *CertService) GenSelfCert(password []byte) ([]byte, []byte, []byte, error) {
	// HERE WE ARE USING THE KEYRING TO STORE THE PASSWORD
	// FOR THE CERTIFICATE AND PRIVATE KEY!!! THE NAME GIVEN
	// TO THE SECRET IS "jwt_secret" AND IT WILL BE USED TO
	// ENCRYPT THE PRIVATE KEY AND STORE IT IN THE KEYRING

	strPassword, err := c.GetPrivPwd(password)
	if err != nil {
		return nil, nil, nil, gl.Errorf("error getting certificate password: %v", err)
	}

	return c.GenerateCertificate(c.certPath, c.keyPath, []byte(strPassword))
}

// DecryptPrivateKey decrypts an encrypted private key using the provided password.
// Parameters:
// - ciphertext: The encrypted private key.
// - password: The password used for decryption.
// Returns: The decrypted private key and an error if any.
func (c *CertService) DecryptPrivateKey(password []byte) (*rsa.PrivateKey, error) {
	if c == nil {
		gl.Fatal("CertService is nil, trying to create a new one")
	}
	c.ensureCrypto()
	certBytes, privKeyBytes, err := c.GetCertAndKeyFromFile()
	if err != nil {
		return nil, gl.Errorf("error getting certificate and key from file: %v", err)
	}
	if len(privKeyBytes) == 0 {
		return nil, gl.Error("private key bytes are empty")
	}
	if len(certBytes) == 0 {
		return nil, gl.Error("certificate bytes are empty")
	}

	if password == nil {
		// Get the password for the private key
		strPassword, err := c.GetPrivPwd(password)
		if err != nil {
			return nil, gl.Errorf("error getting certificate password: %v", err)
		}
		password = []byte(strPassword)
	}
	privKeyDecrypted, privKeyDecryptedEncoded, err := c.security.Decrypt(privKeyBytes, password)
	if err != nil {
		return nil, gl.Errorf("error decrypting private key: %v", err)
	}
	if len(privKeyDecrypted) == 0 {
		return nil, gl.Error("decrypted private key is empty")
	}
	if len(privKeyDecryptedEncoded) == 0 {
		return nil, gl.Error("decrypted encoded private key is empty")
	}
	privKeyParsed, err := x509.ParsePKCS1PrivateKey([]byte(privKeyDecrypted))
	if err != nil {
		return nil, gl.Errorf("error parsing private key: %v", err)
	}
	if privKeyParsed == nil {
		return nil, gl.Error("parsed private key is nil")
	}
	derBytes := x509.MarshalPKCS1PrivateKey(privKeyParsed)
	pob := &rsa.PrivateKey{}
	pob, err = x509.ParsePKCS1PrivateKey(derBytes)
	if err != nil {
		return nil, gl.Errorf("error parsing PKCS1 private key: %v", err)
	}
	// TODO: Store the parsed private key in a secure location if needed
	return pob, nil
}

// GetCertAndKeyFromFile reads the certificate and private key from their respective files.
// Returns: The certificate bytes, the private key bytes, and an error if any.
func (c *CertService) GetCertAndKeyFromFile() ([]byte, []byte, error) {
	if c == nil {
		gl.Warn("CertService is nil, trying to create a new one")
		c = new(CertService)
	}
	if c.certPath == "" {
		c.certPath = os.ExpandEnv(kbx.DefaultGNyxCertPath)
	} else {
		c.certPath = os.ExpandEnv(c.certPath)
	}
	if c.keyPath == "" {
		c.keyPath = os.ExpandEnv(kbx.DefaultGNyxKeyPath)
	} else {
		c.keyPath = os.ExpandEnv(c.keyPath)
	}

	certPath := resolvePathWithFallbacks(c.certPath, ".crt", ".pem")
	keyPath := resolvePathWithFallbacks(c.keyPath, ".key", ".pem")

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, gl.Errorf("error opening certificate file: %v", err)
	}
	if len(certBytes) == 0 {
		return nil, nil, gl.Errorf("certificate file is empty at path: %s", certPath)
	}
	if block, _ := pem.Decode(certBytes); block != nil && len(block.Bytes) > 0 {
		certBytes = block.Bytes
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, gl.Errorf("error opening key file: %v", err)
	}
	if len(keyBytes) == 0 {
		return nil, nil, gl.Errorf("key file is empty at path: %s", keyPath)
	}
	if block, _ := pem.Decode(keyBytes); block != nil && len(block.Bytes) > 0 {
		keyBytes = block.Bytes
	}

	copyCert := make([]byte, len(certBytes))
	copy(copyCert, certBytes)

	copyKey := make([]byte, len(keyBytes))
	copy(copyKey, keyBytes)

	return copyCert, copyKey, nil
}

// VerifyCert verifies the validity of the certificate stored in the configured path.
// Returns: An error if the certificate is invalid or cannot be read.
func (c *CertService) VerifyCert() error {
	if c == nil {
		gl.Warn("CertService is nil, trying to create a new one")
		c = new(CertService)
	}
	certPath := c.certPath
	if certPath == "" {
		certPath = os.ExpandEnv(kbx.DefaultGNyxCertPath)
	}
	certPath = resolvePathWithFallbacks(certPath, ".crt", ".pem")

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return gl.Errorf("error opening certificate file: %v", err)
	}
	if len(certBytes) == 0 {
		return gl.Errorf("certificate file is empty at path: %s", certPath)
	}

	block, rest := pem.Decode(certBytes)
	if block == nil {
		// fallback: try to parse as DER
		if _, err := x509.ParseCertificate(certBytes); err != nil {
			return gl.Errorf("error decoding certificate")
		}
		return nil
	}
	if len(rest) > 0 {
		gl.Warn("extra data found after first PEM block in certificate file")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return gl.Errorf("error parsing certificate: %v", err)
	}
	if cert == nil {
		return gl.Errorf("parsed certificate is nil")
	}
	return nil
}

// GetPublicKey retrieves the public key from the certificate file.
// Returns: The public key and an error if any.
func (c *CertService) GetPublicKey() (*rsa.PublicKey, error) {
	if c == nil {
		gl.Warn("CertService is nil, trying to create a new one")
		c = new(CertService)
	}
	certPath := c.certPath
	if certPath == "" {
		certPath = os.ExpandEnv(kbx.DefaultGNyxCertPath)
	}
	certPath = resolvePathWithFallbacks(certPath, ".crt", ".pem")

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, gl.Errorf("error reading certificate file: %v", err)
	}

	block, _ := pem.Decode(certBytes)
	var cert *x509.Certificate
	if block != nil {
		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, gl.Errorf("error parsing certificate: %v", err)
		}
	} else {
		cert, err = x509.ParseCertificate(certBytes)
		if err != nil {
			return nil, gl.Errorf("error decoding certificate")
		}
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, gl.Errorf("error asserting public key type")
	}

	return pubKey, nil
}

// GetPrivateKey retrieves the private key from the key file.
// Returns: The private key and an error if any.
func (c *CertService) GetPrivateKey() (*rsa.PrivateKey, error) {
	return c.DecryptPrivateKey(nil)
}

// newCertService creates a new instance of CertService with the provided paths.
// Parameters:
// - keyPath: Path to the private key file.
// - certPath: Path to the certificate file.
// Returns: A pointer to a CertService instance.
func newCertService(keyPath, certPath string) *CertService {
	if keyPath == "" {
		keyPath = os.ExpandEnv(kbx.DefaultGNyxKeyPath)
	}
	if certPath == "" {
		certPath = os.ExpandEnv(kbx.DefaultGNyxCertPath)
	}
	crtService := &CertService{
		keyPath:  os.ExpandEnv(keyPath),
		certPath: os.ExpandEnv(certPath),
		security: crp.NewCryptoService(),
	}
	return crtService
}

func resolvePathWithFallbacks(path string, suffixes ...string) string {
	if path == "" {
		return path
	}
	expanded := os.ExpandEnv(path)
	candidates := []string{expanded}
	for _, s := range suffixes {
		if s == "" {
			continue
		}
		if strings.HasSuffix(expanded, s) {
			continue
		}
		candidates = append(candidates, expanded+s)
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return expanded
}

// NewCertService creates a new CertService and returns it as an interface.
// Parameters:
// - keyPath: Path to the private key file.
// - certPath: Path to the certificate file.
// Returns: An implementation of sci.ICertService.
func NewCertService(keyPath, certPath string) sci.ICertService {
	return newCertService(keyPath, certPath)
}

// NewCertServiceType creates a new CertService and returns it as a concrete type.
// Parameters:
// - keyPath: Path to the private key file.
// - certPath: Path to the certificate file.
// Returns: A pointer to a CertService instance.
func NewCertServiceType(keyPath, certPath string) *CertService {
	return newCertService(keyPath, certPath)
}
