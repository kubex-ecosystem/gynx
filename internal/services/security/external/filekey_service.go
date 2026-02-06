// Package external implements a file-based, AES-GCM encrypted replacement for go-keyring.
package external

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	crp "github.com/kubex-ecosystem/gnyx/internal/services/security/crypto"
	sci "github.com/kubex-ecosystem/gnyx/internal/services/security/interfaces"
	kbx "github.com/kubex-ecosystem/kbx"

	defaults "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

var (
	kbxFileKeyService    *FileKeyService
	kbxFileCryptoService sci.ICryptoService
	globalMasterKey      []byte
	globalBaseDir        string
)

// FileKeyService is a drop-in replacement for KeyService,
// maintaining the same contract and method signatures.
type FileKeyService struct {
	KeyService    kbx.GlobalRef
	keyringName   kbx.GlobalRef
	masterKey     []byte
	baseDir       string
	cryptoService sci.ICryptoService
}

// NewFileKeyService creates a new encrypted file-based secret store.
func NewFileKeyService(service, name string) sci.IKeyService {
	return newFileKeyService(service, name)
}

// NewFileKeyServiceType creates a new encrypted file-based secret store and returns the concrete type.
func NewFileKeyServiceType(service, name string) *FileKeyService {
	return newFileKeyService(service, name)
}

func newFileKeyService(service, name string) *FileKeyService {
	// Inicializa apenas uma vez as configurações globais
	if globalMasterKey == nil {
		// Use kbxGet.EnvOr para resiliência na obtenção do diretório de secrets
		dir := kbxGet.EnvOr("APP_SECRETS_DIR", os.ExpandEnv(defaults.DefaultVaultDir))

		// Tenta criar o diretório
		if err := os.MkdirAll(dir, 0o700); err != nil {
			gl.Log("warn", fmt.Sprintf("Failed to create secrets directory %s: %v, trying fallback", dir, err))

			// Fallback para $HOME/.gnyx/secrets se falhar
			homeDir, homeErr := os.UserHomeDir()
			if homeErr == nil {
				dir = filepath.Join(homeDir, ".gnyxsecrets")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					gl.Log("warn", fmt.Sprintf("Failed to create fallback secrets directory %s: %v", dir, err))
				} else {
					gl.Log("info", fmt.Sprintf("Using fallback secrets directory: %s", dir))
				}
			}
		}
		gl.Log("debug", fmt.Sprintf("Secrets directory ready: %s", dir))

		globalBaseDir = dir
		masterKeyPath := filepath.Join(dir, "kubex_kubex-jwt_secret.secret")
		ephemeralKeyPath := filepath.Join(dir, "kubex_ephemeral_jwt_secret.secret")

		// 1) Se APP_MASTER_KEY estiver set, usa ela.
		mk := strings.TrimSpace(kbxGet.EnvOr("APP_MASTER_KEY", ""))

		// 2) Caso contrário, tenta ler chave persistida (principal ou a efêmera gerada antes)
		if mk == "" {
			if data, err := os.ReadFile(masterKeyPath); err == nil && len(data) > 0 {
				mk = strings.TrimSpace(string(data))
				gl.Debugf("Loaded master key from %s", masterKeyPath)
			} else if data, err := os.ReadFile(ephemeralKeyPath); err == nil && len(data) > 0 {
				mk = strings.TrimSpace(string(data))
				gl.Debugf("Loaded ephemeral master key from %s", ephemeralKeyPath)
			}
		}

		// 3) Se ainda vazio, gera e persiste (no caminho principal para reutilizar em restarts)
		if mk == "" {
			gl.Debugf("APP_MASTER_KEY not set; generating persistent master key")
			tmp := make([]byte, 32)
			_, _ = rand.Read(tmp)
			mk = base64.StdEncoding.EncodeToString(tmp)

			if err := os.WriteFile(masterKeyPath, []byte(mk), 0o600); err != nil {
				gl.Warnf("Failed to persist master key: %v", err)
			} else {
				gl.Debugf("Master key persisted to: %s", masterKeyPath)
			}
		}

		raw, err := base64.StdEncoding.DecodeString(mk)
		if err != nil || len(raw) == 0 {
			gl.Log("fatal", fmt.Sprintf("failed to decode master key: %v", err))
		}
		globalMasterKey = raw
	}

	// Inicializa o crypto service se necessário
	if kbxFileCryptoService == nil {
		kbxFileCryptoService = crp.NewCryptoService()
	}

	return &FileKeyService{
		KeyService:    kbx.NewGlobalRef(service),
		keyringName:   kbx.NewGlobalRef(name),
		masterKey:     globalMasterKey,
		baseDir:       globalBaseDir,
		cryptoService: kbxFileCryptoService,
	}
}

func (k *FileKeyService) StorePassword(password string) error {
	if password == "" {
		gl.Log("error", "key cannot be empty")
		return gl.Errorf("key cannot be empty")
	}

	// Garante que o diretório existe antes de escrever
	if err := os.MkdirAll(k.baseDir, 0o700); err != nil {
		return gl.Errorf("error creating secrets directory: %v", err)
	}

	enc, err := k.encrypt([]byte(password))
	if err != nil {
		return gl.Errorf("error encrypting password: %v", err)
	}
	path := filepath.Join(k.baseDir, fmt.Sprintf("%s_%s.secret", k.KeyService.GetName(), k.keyringName.GetName()))
	if err := os.WriteFile(path, []byte(enc), 0o600); err != nil {
		return gl.Errorf("error storing key: %v", err)
	}
	gl.Log("debug", fmt.Sprintf("key stored successfully: %s", k.keyringName.GetName()))
	return nil
}

func (k *FileKeyService) RetrievePassword() (string, error) {
	path := filepath.Join(k.baseDir, fmt.Sprintf("%s_%s.secret", k.KeyService.GetName(), k.keyringName.GetName()))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", os.ErrNotExist
		}
		gl.Log("debug", fmt.Sprintf("error reading key: %v", err))
		return "", gl.Errorf("error retrieving key: %v", err)
	}
	plain, err := k.decrypt(string(data))
	if err != nil {
		return "", gl.Errorf("error decrypting key: %v", err)
	}
	return string(plain), nil
}

func (k *FileKeyService) DeletePassword() error {
	path := filepath.Join(k.baseDir, fmt.Sprintf("%s_%s.secret", k.KeyService.GetName(), k.keyringName.GetName()))
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			gl.Log("debug", fmt.Sprintf("key not found for deletion: %s", k.keyringName.GetName()))
			return os.ErrNotExist
		}
		gl.Log("debug", fmt.Sprintf("error deleting key: %v", err))
		return gl.Errorf("error deleting key: %v", err)
	}
	gl.Log("debug", fmt.Sprintf("key deleted successfully: %s", k.keyringName.GetName()))
	return nil
}

func (k *FileKeyService) RetrieveOrCreatePassword() (string, error) {
	if k == nil {
		gl.Log("fatal", "FileKeyService is nil")
		return "", errors.New("FileKeyService is nil")
	}
	password, err := k.RetrievePassword()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || len(password) == 0 {
			gl.Log("debug", "password not found, generating a new one")
			newPasswordBytes, err := k.cryptoService.GenerateKeyWithLength(32)
			if err != nil {
				return "", gl.Errorf("error generating new password: %v", err)
			}
			newPassword := k.cryptoService.EncodeBase64(newPasswordBytes)
			if err := k.StorePassword(newPassword); err != nil {
				return "", gl.Errorf("error storing new password: %v", err)
			}
			return newPassword, nil
		}
		return "", gl.Errorf("error retrieving password: %v", err)
	}
	return password, nil
}

// --- internal helpers ---

func (k *FileKeyService) encrypt(plain []byte) (string, error) {
	hash := sha256.Sum256(k.masterKey)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := aead.Seal(nil, nonce, plain, nil)
	full := append(nonce, ct...)
	return base64.StdEncoding.EncodeToString(full), nil
}

func (k *FileKeyService) decrypt(ciphertext string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(k.masterKey)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < aead.NonceSize() {
		return nil, errors.New("invalid ciphertext")
	}
	nonce, ct := data[:aead.NonceSize()], data[aead.NonceSize():]
	return aead.Open(nil, nonce, ct, nil)
}
