// Package cookies testa utilitários de manipulação de cookies HTTP.
//
// Estes testes validam:
// - Resolução de opções por ambiente
// - Obtenção de valores de cookies
// - Configurações de segurança (Secure, HTTPOnly, SameSite)
//
// Tipo: Testes unitários
// Previne: Cookies inseguros em produção, falhas na leitura de cookies
package cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- Testes de GetCookieValue ---
// Previne: Falhas na leitura de cookies válidos

func TestGetCookieValue(t *testing.T) {
	t.Run("returns value for existing cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "test_cookie",
			Value: "test_value",
		})

		value, ok := GetCookieValue(req, "test_cookie")
		if !ok {
			t.Error("expected ok to be true for existing cookie")
		}
		if value != "test_value" {
			t.Errorf("value = %q, want %q", value, "test_value")
		}
	})

	t.Run("returns false for non-existent cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		_, ok := GetCookieValue(req, "non_existent")
		if ok {
			t.Error("expected ok to be false for non-existent cookie")
		}
	})

	t.Run("returns false for empty cookie value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "empty_cookie",
			Value: "",
		})

		_, ok := GetCookieValue(req, "empty_cookie")
		if ok {
			t.Error("expected ok to be false for empty cookie value")
		}
	})

	t.Run("handles multiple cookies", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "cookie1", Value: "value1"})
		req.AddCookie(&http.Cookie{Name: "cookie2", Value: "value2"})
		req.AddCookie(&http.Cookie{Name: "cookie3", Value: "value3"})

		value, ok := GetCookieValue(req, "cookie2")
		if !ok {
			t.Error("expected ok to be true")
		}
		if value != "value2" {
			t.Errorf("value = %q, want %q", value, "value2")
		}
	})
}

// --- Testes de ResolveCookieOptions ---
// NOTA: Este teste depende de variáveis de ambiente que podem variar.
// Testamos a estrutura retornada, não os valores específicos.

func TestResolveCookieOptions(t *testing.T) {
	t.Run("returns valid CookieOptions struct", func(t *testing.T) {
		opts := ResolveCookieOptions()

		// Path deve ser sempre "/"
		if opts.Path != "/" {
			t.Errorf("Path = %q, want %q", opts.Path, "/")
		}

		// HTTPOnly deve ser sempre true para cookies de autenticação
		if !opts.HTTPOnly {
			t.Error("HTTPOnly should always be true for security")
		}

		// SameSite deve ser um valor válido
		validSameSite := opts.SameSite == http.SameSiteLaxMode ||
			opts.SameSite == http.SameSiteStrictMode ||
			opts.SameSite == http.SameSiteNoneMode ||
			opts.SameSite == http.SameSiteDefaultMode
		if !validSameSite {
			t.Errorf("SameSite = %v, not a valid http.SameSite value", opts.SameSite)
		}
	})
}

// --- Testes de SetAuthCookie ---
// Previne: Cookies com configurações incorretas

func TestSetAuthCookie(t *testing.T) {
	t.Run("sets cookie with correct name and value", func(t *testing.T) {
		w := httptest.NewRecorder()

		SetAuthCookie(w, "test_auth", "token123", time.Now().Add(time.Hour))

		cookies := w.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("expected at least one cookie to be set")
		}

		var found bool
		for _, c := range cookies {
			if c.Name == "test_auth" {
				found = true
				if c.Value != "token123" {
					t.Errorf("cookie value = %q, want %q", c.Value, "token123")
				}
				if !c.HttpOnly {
					t.Error("cookie should have HttpOnly flag")
				}
				break
			}
		}
		if !found {
			t.Error("cookie 'test_auth' not found")
		}
	})
}

// --- Testes de ClearAuthCookie ---
// Previne: Cookies não sendo limpos corretamente

func TestClearAuthCookie(t *testing.T) {
	t.Run("clears cookie by setting MaxAge to -1", func(t *testing.T) {
		w := httptest.NewRecorder()

		ClearAuthCookie(w, "auth_cookie")

		cookies := w.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("expected cookie to be set for clearing")
		}

		c := cookies[0]
		if c.Name != "auth_cookie" {
			t.Errorf("cookie name = %q, want %q", c.Name, "auth_cookie")
		}
		if c.MaxAge != -1 {
			t.Errorf("MaxAge = %d, want -1", c.MaxAge)
		}
		if c.Value != "" {
			t.Errorf("Value = %q, want empty string", c.Value)
		}
	})
}

// --- Testes de constantes de cookie ---

func TestCookieConstants(t *testing.T) {
	if CookieAccessToken == "" {
		t.Error("CookieAccessToken should not be empty")
	}
	if CookieRefreshToken == "" {
		t.Error("CookieRefreshToken should not be empty")
	}
	if CookieAccessToken == CookieRefreshToken {
		t.Error("CookieAccessToken and CookieRefreshToken should be different")
	}
}
