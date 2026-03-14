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
			// Verifico se o roteador está em DEBUG, porque se estiver, nem vou considerar nada além da variável própria da kubex e fallback em dev direto.
			kbxGet.EnvOr("GIN_MODE", "debug") != "release",

			// Se estiver em release, prioriza produção caso variável não esteja setada
			kbxGet.EnvOr("KUBEX_GNYX_ENV", "development"),

			// Caso contrário, iremos considerar qualquer cenário das variáveis abaixo como fonte do valor e usar production como fallback.
			kbxGet.EnvOr("BUILD_MODE",
				kbxGet.EnvOr("KUBEX_ENV",
					kbxGet.EnvOr("ENV",
						kbxGet.EnvOr("KUBEX_GNYX_ENV", "production"),
					),
				),
			),
		),
	)
	publicDomain := strings.TrimSpace(
		kbxGet.ValueOrIf(
			// Confere se não está em desenvolvimento
			env != "development",
			// Se não estiver em desenvolvimento, insere o domínio oficial como fallback
			kbxGet.EnvOr("PUBLIC_DOMAIN", ".kubex.world"),
			// Se estiver em desenvolvimento, insere localhost como fallback,mas ainda permite valor vindo da variável caso preenchida
			kbxGet.EnvOr("PUBLIC_DOMAIN", "localhost"),
		),
	)

	cookie := CookieOptions{}

	switch env {
	case "staging", "stage", "testing", "test":
		// homologação/testes
		cookie.Domain = publicDomain
		cookie.Secure = false // ajustar se necessário
		cookie.SameSite = http.SameSiteLaxMode
		cookie.HTTPOnly = true
		cookie.Path = "/"
		return cookie
	case "development", "dev", "local":
		// local/development
		cookie.Domain = publicDomain
		cookie.Secure = false
		cookie.SameSite = http.SameSiteLaxMode
		cookie.HTTPOnly = true
		cookie.Path = "/"
		return cookie
	default:
		// produção
		// Na produção DE VERDADE, o domínio sempre será .kubex.world
		cookie.Domain = ".kubex.world"
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
