package interfaces

import "crypto/rsa"

type ICertManager interface {
	GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, []byte, error)
	VerifyCert() error
	GetCertAndKeyFromFile() ([]byte, []byte, error)
}

type ICertService interface {
	GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, []byte, error)
	GenSelfCert(password []byte) ([]byte, []byte, []byte, error)
	DecryptPrivateKey(password []byte) (*rsa.PrivateKey, error)
	VerifyCert() error
	GetCertAndKeyFromFile() ([]byte, []byte, error)
	GetPublicKey() (*rsa.PublicKey, error)
	GetPrivateKey() (*rsa.PrivateKey, error)
	GetPrivPwd(key []byte) (string, error)
}
