// Package cookies fornece utilitários para manipulação segura de cookies HTTP
// com base no ambiente de implantação e nas melhores práticas de segurança.
package cookies

import (
	"net/http"
	"strings"
	"time"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
)

// Cookie names (padrão Kubex/Kubex)
const (
	CookieAccessToken  = "kubex_at"
	CookieRefreshToken = "kubex_rt"
)

// CookieOptions representa configurações finais usadas no Set-Cookie
type CookieOptions struct {
	Domain   string
	Path     string
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

// ResolveCookieOptions determina automaticamente as configurações corretas
// com base no ambiente, domínio público e requisitos de segurança.
func ResolveCookieOptions() CookieOptions {
	env := strings.ToLower(
		kbxGet.ValueOrIf(
			kbxGet.EnvOr("GIN_MODE", "debug") == "release", // Confere se está em modo release
			kbxGet.EnvOr("KUBEX_GNYX_ENV", "production"),   // Se estiver em release, prioriza produção caso variável não esteja setada
			kbxGet.EnvOr("KUBEX_GNYX_ENV", "development"),  // Senão, assume development como padrão, caso variável não esteja setada
		),
	)
	publicDomain := strings.TrimSpace(
		kbxGet.ValueOrIf(
			env == "production", // Confere se está em produção
			kbxGet.EnvOr("PUBLIC_DOMAIN", ".gnyx.kubex.world"), // Se estiver em produção, prioriza domínio oficial caso variável não esteja setada
			kbxGet.EnvOr("PUBLIC_DOMAIN", "localhost"),         // Senão, assume localhost como padrão, caso variável não esteja setada
		),
	)

	cookie := CookieOptions{}

	switch env {
	case "staging", "stage", "testing", "test":
		// homologação/testes
		cookie.Domain = publicDomain
		cookie.Secure = false // ajustar conforme necessário
		cookie.SameSite = http.SameSiteLaxMode
		cookie.HTTPOnly = true
		cookie.Path = "/"
		return cookie
	case "development", "dev", "local":
		// local/development
		// cookie.Domain = "localhost"
		cookie.Secure = false
		cookie.SameSite = http.SameSiteLaxMode
		cookie.HTTPOnly = true
		cookie.Path = "/"
		return cookie
	default:
		// produção
		cookie.Domain = ".gnyx.kubex.world"
		cookie.Secure = true
		cookie.SameSite = http.SameSiteNoneMode
		cookie.HTTPOnly = true
		cookie.Path = "/"
		return cookie
	}
}

// SetAuthCookie escreve um cookie seguro no cliente
func SetAuthCookie(w http.ResponseWriter, name, value string, expires time.Time) {
	opts := ResolveCookieOptions()

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     opts.Path,
		Domain:   opts.Domain,
		Expires:  expires,
		Secure:   opts.Secure,
		HttpOnly: opts.HTTPOnly,
		SameSite: opts.SameSite,
	})
}

// ClearAuthCookie limpa um cookie autenticador
func ClearAuthCookie(w http.ResponseWriter, name string) {
	opts := ResolveCookieOptions()

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     opts.Path,
		Domain:   opts.Domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   opts.Secure,
		HttpOnly: opts.HTTPOnly,
		SameSite: opts.SameSite,
	})
}

// GetCookieValue ajuda handlers a obter valor de cookie
func GetCookieValue(r *http.Request, name string) (string, bool) {
	c, err := r.Cookie(name)
	if err != nil || c == nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}
