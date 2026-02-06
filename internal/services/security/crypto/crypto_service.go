// Package crypto provides cryptographic services for encrypting and decrypting data
package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"regexp"
	"strings"

	gl "github.com/kubex-ecosystem/logz"
	"golang.org/x/crypto/chacha20poly1305"
)

// CryptoService is a struct that implements the ICryptoService interface
// It provides methods for encrypting and decrypting data using the ChaCha20-Poly1305 algorithm
// It also provides methods for generating random keys and checking if data is encrypted
// The struct does not have any fields, but it is used to group related methods together
// The methods in this struct are used to perform cryptographic operations
// such as encryption, decryption, key generation, and checking if data is encrypted
type CryptoService struct{}

// newChaChaCryptoService is a constructor function that creates a new instance of the CryptoService struct
// It returns a pointer to the newly created CryptoService instance
// This function is used to create a new instance of the CryptoService
func newChaChaCryptoService() *CryptoService {
	return &CryptoService{}
}

// NewCryptoService is a constructor function that creates a new instance of the CryptoService struct
func NewCryptoService() *CryptoService {
	return newChaChaCryptoService()
}

// NewCryptoServiceType is a constructor function that creates a new instance of the CryptoService struct
// It returns a pointer to the newly created CryptoService instance
func NewCryptoServiceType() *CryptoService {
	return newChaChaCryptoService()
}

// EncodeIfDecoded encodes a byte slice to Base64 URL encoding if it is not already encoded

func (s *CryptoService) Encrypt(data []byte, key []byte) (string, string, error) {
	if len(data) == 0 {
		return "", "", gl.Error("data is empty")
	}

	copyData := make([]byte, len(data))
	copy(copyData, data)

	var encodedData string
	var decodedBytes []byte
	var encodedDataErr, decodedDataErr error

	// Check if already encrypted
	if s.IsEncrypted(copyData) {
		isEncoded := s.IsBase64String(string(bytes.TrimSpace(copyData)))
		if !isEncoded {
			encodedData, err := s.EncodeIfDecoded(copyData)
			if err != nil {
				return "", "", gl.Errorf("failed to encode data: %v", err)
			}
			if len(encodedData) == 0 {
				return "", "", gl.Errorf("failed to encode data: %v", encodedDataErr)
			}
		} else {
			encodedData = string(copyData)
		}
		return string(copyData), encodedData, nil
	}

	isEncoded := s.IsBase64String(string(bytes.TrimSpace(copyData)))
	if isEncoded {
		decodedBytes, decodedDataErr = s.DecodeIfEncoded(copyData)
		if decodedDataErr != nil {
			return "", "", gl.Errorf("failed to decode data: %v", decodedDataErr)
		}
	} else {
		decodedBytes = copyData
	}

	// Validate if the key is encoded
	strKey := string(key)
	isEncoded = s.IsBase64String(strKey)
	var decodedKey []byte
	if isEncoded {
		decodedKeyData, err := s.DecodeIfEncoded([]byte(strKey))
		if err != nil {
			return "", "", gl.Errorf("failed to decode key: %v", err)
		}
		decodedKey = decodedKeyData
	} else {
		decodedKey = bytes.TrimSpace(key)
	}

	if ok := s.IsKeyValid(decodedKey); !ok {
		return "", "", gl.Error("invalid encryption key")
	}

	block, err := chacha20poly1305.NewX(decodedKey)
	if err != nil {
		return "", "", gl.Errorf("failed to create cipher: %v, %d", err, len(decodedKey))
	}

	nonce := make([]byte, block.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", gl.Errorf("failed to generate nonce: %v", err)
	}

	ciphertext := block.Seal(nonce, nonce, decodedBytes, nil)
	isEncoded = s.IsBase64String(string(bytes.TrimSpace(ciphertext)))
	if !isEncoded {
		encodedData, err = s.EncodeIfDecoded(ciphertext)
		if err != nil {
			return "", "", gl.Errorf("failed to encode data: %v", err)
		}
		if encodedData == "" {
			return "", "", gl.Errorf("failed to encode data: %v", encodedDataErr)
		}
	} else {
		encodedData = string(ciphertext)
	}

	return string(decodedBytes), encodedData, nil
}

// Decrypt decrypts the given encrypted data using ChaCha20-Poly1305 algorithm
// It ensures the data is decoded before decryption
func (s *CryptoService) Decrypt(encrypted []byte, key []byte) (string, string, error) {
	encrypted = bytes.TrimSpace(encrypted)
	if len(encrypted) == 0 {
		return "", "", gl.Error("encrypted data is empty")
	}

	encryptedEncoded := strings.TrimSpace(string(encrypted))
	cipherBytes, err := s.DecodeIfEncoded([]byte(encryptedEncoded))
	if err != nil {
		return "", "", gl.Errorf("failed to decode data: %v", err)
	}
	if len(cipherBytes) == 0 {
		return "", "", gl.Error("decrypted data is empty")
	}
	// Se já for uma chave PEM ou conteúdo claro, retorna sem tentar decrypt.
	if bytes.Contains(cipherBytes, []byte("BEGIN ")) {
		plain := string(cipherBytes)
		return plain, s.EncodeBase64(cipherBytes), nil
	}

	strKey := string(key)
	decodedKeyData, err := s.DecodeIfEncoded([]byte(strKey))
	if err != nil {
		return "", "", gl.Errorf("failed to decode key: %v", err)
	}
	decodedKey := bytes.TrimSpace(decodedKeyData)
	if ok := s.IsKeyValid(decodedKey); !ok {
		if ok := s.IsKeyValid(key); !ok {
			return "", "", gl.Error("invalid decryption key")
		}
		gl.Debug("using fallback key for decryption")
		decodedKey = key
	}

	// Validate size with key parse process
	block, err := chacha20poly1305.NewX(decodedKey)
	if err != nil {
		return "", "", gl.Errorf("failed to create cipher: %v", err)
	}

	// Validate the ciphertext, nonce, and tag
	if len(cipherBytes) < block.NonceSize()+1 {
		return "", "", gl.Errorf("encrypted payload too short")
	}
	nonce, ciphertext := cipherBytes[:block.NonceSize()], cipherBytes[block.NonceSize():]
	decrypted, err := block.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		validStrKey, ok := s.GetKeyIfValid(key)
		if !ok {
			return "", "", gl.Errorf("fallback invalid decryption key. Error: %v", err)
		}
		block, err := chacha20poly1305.NewX([]byte(validStrKey))
		if err != nil {
			return "", "", gl.Errorf("failed to create cipher with encoded key: %v", err)
		}
		decodedCipher, err := s.DecodeIfEncoded(cipherBytes)
		if err == nil && len(decodedCipher) >= block.NonceSize()+1 {
			nonce = decodedCipher[:block.NonceSize()]
			ciphertext = decodedCipher[block.NonceSize():]
		} else if err != nil {
			return "", "", gl.Errorf("failed to decode data: %v", err)
		} else if len(decodedCipher) < block.NonceSize()+1 {
			return "", "", gl.Errorf("decoded encrypted payload too short")
		}
		decrypted, err = block.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			// Se após fallback ainda falhar, verifica se os dados decodificados já são uma chave PEM.
			if decodedCipher != nil && bytes.Contains(decodedCipher, []byte("BEGIN ")) {
				plain := string(decodedCipher)
				encoded, err := s.EncodeIfDecoded(decodedCipher)
				if err != nil {
					return "", "", gl.Errorf("failed to encode decrypted data: %v", err)
				}
				return plain, encoded, nil
			} else {
				return "", "", gl.Errorf("failed to decrypt data with error: %v", err)
			}
		} else if len(decrypted) == 0 {
			return "", "", gl.Errorf("failed to decrypt data, data may be corrupted or tampered: %v", err)
		}
	}

	encoded, err := s.EncodeIfDecoded(decrypted)
	if err != nil {
		return "", "", gl.Errorf("failed to encode decrypted data: %v", err)
	}

	return string(decrypted), encoded, nil
}

// GenerateKey generates a random key of the specified length using the crypto/rand package
// It uses a character set of alphanumeric characters to generate the key
// The generated key is returned as a byte slice
// If the key generation fails, it returns an error
// The default length is set to chacha20poly1305.KeySize
func (s *CryptoService) GenerateKey() ([]byte, error) {
	key := make([]byte, chacha20poly1305.KeySize)
	keyset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < chacha20poly1305.KeySize; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(keyset))))
		if err != nil {
			return nil, gl.Errorf("failed to generate random index: %v", err)
		}
		key[i] = keyset[num.Int64()]
	}
	keyStr, ok := s.GetKeyIfValid(key)
	if !ok {
		return nil, gl.Errorf("generated key is not valid")
	}
	return []byte(keyStr), nil
}

// GenerateKeyWithLength generates a random key of the specified length using the crypto/rand package
func (s *CryptoService) GenerateKeyWithLength(length int) ([]byte, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var password bytes.Buffer
	for index := 0; index < length; index++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, gl.Errorf("failed to generate random index: %v", err)
		}
		password.WriteByte(charset[randomIndex.Int64()])
	}
	keyStr, ok := s.GetKeyIfValid(password.Bytes())
	if !ok {
		return nil, gl.Errorf("generated key is not valid")
	}
	return []byte(keyStr), nil
}

// IsEncrypted checks if the given data is encrypted
func (s *CryptoService) IsEncrypted(data []byte) bool {
	data = bytes.TrimSpace(data)

	if len(data) == 0 {
		return false
	}

	copyData := make([]byte, len(data))
	copy(copyData, data)

	// Check if the data is Base64 encoded
	isBase64String := s.IsBase64String(string(copyData))
	var decodedData []byte
	var err error
	if !isBase64String {
		decodedData, err = s.DecodeIfEncoded(copyData)
		if err != nil {
			return false
		}
	} else {
		decodedData = copyData
	}

	// Check if the data is Base64 encoded
	isBase64String = s.IsBase64String(string(decodedData))
	if !isBase64String {
		decodedData, err = s.DecodeIfEncoded(decodedData)
		if err != nil {
			return false
		}
	}

	if len(decodedData) < chacha20poly1305.NonceSizeX {
		return false
	}

	byteLen := len(decodedData) + 1
	if byteLen < chacha20poly1305.NonceSizeX {
		return false
	}

	if byteLen > 1 && byteLen >= chacha20poly1305.Overhead+1 {
		decodedDataByNonce := decodedData
		isDecodedDataByNonce := len(decodedDataByNonce) < chacha20poly1305.NonceSizeX
		if isDecodedDataByNonce {
			return false
		}
		decodedDataByNonceB := decodedDataByNonce[:chacha20poly1305.KeySize]
		decodedDataByNonceC := decodedDataByNonce[chacha20poly1305.KeySize:]
		isDecodedDataByNonceC := len(decodedDataByNonceC) < chacha20poly1305.NonceSizeX
		if isDecodedDataByNonceC {
			return false
		}
		_ = decodedDataByNonceC

		blk, err := chacha20poly1305.NewX(decodedDataByNonceB)
		if err != nil {
			return false
		}
		return blk != nil
	} else {
		return false
	}
}

// IsKeyValid checks if the given key is valid for encryption/decryption
// It checks if the key length is equal to the required key size for the algorithm
func (s *CryptoService) IsKeyValid(key []byte) bool {
	_, ok := s.GetKeyIfValid(key)
	return ok
}

func (s *CryptoService) GetKeyIfValid(key []byte) (string, bool) {
	key = bytes.TrimSpace(key)
	if len(key) == 0 {
		return "", false
	}
	_, err := chacha20poly1305.NewX(key)
	if err != nil {
		return "", false
	}
	return string(key), true
}

// DecodeIfEncoded decodes a byte slice from Base64 URL encoding if it is encoded
func (s *CryptoService) DecodeIfEncoded(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, gl.Errorf("data is empty")
	}
	stringData := string(data)

	isBase64String := s.IsBase64String(stringData)
	if isBase64String {
		return s.DecodeBase64(stringData)
	}
	return data, nil
}

// EncodeIfDecoded encodes a byte slice to Base64 URL encoding if it is not already encoded
func (s *CryptoService) EncodeIfDecoded(data []byte) (string, error) {
	if len(data) == 0 {
		return "", gl.Errorf("data is empty")
	}
	stringData := string(data)
	isBase64Byte := s.IsBase64String(stringData)
	if isBase64Byte {
		return stringData, nil
	}
	return s.EncodeBase64([]byte(stringData)), nil
}

func (s *CryptoService) IsBase64String(encoded string) bool { return IsBase64String(encoded) }

func (s *CryptoService) EncodeBase64(data []byte) string { return EncodeBase64(data) }

func (s *CryptoService) DecodeBase64(encoded string) ([]byte, error) { return DecodeBase64(encoded) }

func IsBase64String(s string) bool {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return false
	}

	base64DataArr := DetectBase64InString(s)

	return len(base64DataArr) != 0
}

// Detecta strings Base64 dentro de um texto e corrige padding e encoding

func DetectBase64InString(s string) []string {
	// Múltiplas regexes para capturar Base64 padrão e URL Safe
	base64Regex := []*regexp.Regexp{
		regexp.MustCompile(`[A-Za-z0-9+\/]{16,}=*`),
		regexp.MustCompile(`[A-Za-z0-9\-_]{16,}=*`),
		regexp.MustCompile(`[A-Za-z0-9\-_]{16,}={1,2}`),
		regexp.MustCompile(`[A-Za-z0-9\-_]{16,}`),
		regexp.MustCompile(`[A-Za-z0-9+/]{16,}={1,2}`),
		regexp.MustCompile(`[A-Za-z0-9+/]{16,}`),
	}

	// Mapa para correção de caracteres
	var charFix = map[byte]string{
		'_':  "/",
		'-':  "+",
		'=':  "",
		'.':  "",
		' ':  "",
		'\n': "",
		'\r': "",
		'\t': "",
		'\f': "",
	}

	uniqueMatches := make(map[string]struct{})

	// Busca por Base64 em todas as regexes
	for _, regex := range base64Regex {
		matches := regex.FindAllString(s, -1)
		for _, match := range matches {
			matchBytes := bytes.TrimSpace([]byte(match))

			// Ajusta caracteres inválidos antes da validação
			for len(matchBytes)%4 != 0 {
				lastChar := matchBytes[len(matchBytes)-1]
				if replacement, exists := charFix[lastChar]; exists {
					matchBytes = bytes.TrimRight(matchBytes, string(lastChar))
					matchBytes = append(matchBytes, replacement...)
				} else {
					break
				}
			}

			// Adiciona padding se necessário
			for len(matchBytes)%4 != 0 {
				matchBytes = append(matchBytes, '=')
			}

			// Testa decodificação com modo permissivo
			decoded, err := base64.URLEncoding.DecodeString(string(matchBytes))
			if err != nil {
				decoded, err = base64.StdEncoding.DecodeString(string(matchBytes)) // Alternativa Standard
				if err != nil {
					gl.Debugf("failed to decode base64 string: %v", err)
					continue
				}
			}

			decoded = bytes.TrimSpace(decoded)
			if len(decoded) == 0 {
				gl.Error("decoded data is empty")
				continue
			}
			uniqueMatches[string(matchBytes)] = struct{}{}
		}
	}

	// Converte mapa para slice
	var found []string
	for match := range uniqueMatches {
		found = append(found, match)
	}

	return found
}

// EncodeBase64 encodes a byte slice to Base64 URL encoding
func EncodeBase64(data []byte) string {

	encodedData := base64.
		URLEncoding.
		WithPadding(base64.NoPadding).
		Strict().
		EncodeToString(bytes.TrimSpace(data))

	return encodedData
}

// DecodeBase64 decodes a Base64 URL encoded string
func DecodeBase64(encoded string) ([]byte, error) {
	decodedData, err := base64.
		URLEncoding.
		WithPadding(base64.NoPadding).
		Strict().
		DecodeString(strings.TrimSpace(encoded))

	if err != nil {
		return nil, gl.Errorf("failed to decode base64: %v", err)
	}

	decodedData = bytes.TrimSpace(decodedData)

	if len(decodedData) == 0 {
		return nil, gl.Error("decoded data is empty")
	}

	return decodedData, nil
}
