package external

import (
	sci "github.com/kubex-ecosystem/gnyx/internal/services/security/interfaces"
	"github.com/kubex-ecosystem/gnyx/internal/services/security/storage"
)

type SecretStorageAdapter struct {
	backend storage.ISecretStorage
}

func NewSecretStorageAdapter(backend storage.ISecretStorage) sci.IKeyService {
	return &SecretStorageAdapter{backend: backend}
}

func (a *SecretStorageAdapter) StorePassword(password string) error {
	return a.backend.StorePassword(password)
}

func (a *SecretStorageAdapter) RetrievePassword() (string, error) {
	return a.backend.RetrievePassword()
}
