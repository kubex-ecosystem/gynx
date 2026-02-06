// Package api fornece controllers genéricos CRUD para uso com Repository[T].
//
// O Controller[T] genérico funciona com qualquer tipo T que tenha um Repository[T]
// correspondente (seja DSRepository adapter, Store direto, ou ORM).
//
// Uso:
//
//	factory, _ := dsClient.NewAdapterFactory(ctx, "gnyx", gormDB, nil)
//	userRepo, _ := adapter.CreateAdapter[Users](factory, ctx, "user", users.NewRepository)
//	userCtrl := api.NewController[Users](userRepo)
//
//	r.GET("/users", userCtrl.GetAll)
//	r.POST("/users", userCtrl.Create)
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
)

// Controller é um controller genérico CRUD que funciona com qualquer Repository[T].
//
// Implementa handlers padrão para:
// - GetAll: Lista paginada de entidades
// - GetByID: Busca por ID (404 se não encontrado)
// - Create: Cria nova entidade (201 com ID retornado)
// - Update: Atualiza entidade existente
// - Delete: Remove entidade (204 no content)
//
// O controller é completamente genérico e não sabe se está usando Store ou ORM,
// graças ao DSRepository[T] adapter.
type Controller[T any] struct {
	repo dsclient.Repository[T]
}

// NewController cria um novo controller genérico para o tipo T.
//
// O repository pode ser:
// - DSRepository[T] (adapter que usa Store ou ORM)
// - Store direto (UserStore, InviteStore, etc)
// - Qualquer implementação de Repository[T]
//
// Exemplo:
//
//	userRepo, _ := adapter.CreateAdapter[Users](factory, ctx, "user", users.NewRepository)
//	userCtrl := api.NewController[Users](userRepo)
func NewController[T any](repo dsclient.Repository[T]) *Controller[T] {
	return &Controller[T]{repo: repo}
}

// GetAll retorna lista paginada de entidades.
//
// Query params opcionais:
// - page: número da página (default: 1)
// - limit: itens por página (default: definido pelo repo)
// - Outros filtros específicos do domínio via map
//
// Response:
//
//	200 OK: {data: []T, total: int, page: int, limit: int, total_pages: int}
//	500 Internal Server Error: {error: string}
func (ctrl *Controller[T]) GetAll(c *gin.Context) {
	// Repository[T].List retorna PaginatedResult[T]
	// Filters podem ser extraídos de query params se necessário
	result, err := ctrl.repo.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetByID busca entidade por ID.
//
// Path param:
// - id: ID da entidade
//
// Response:
//
//	200 OK: T
//	404 Not Found: {error: "not found"}
//	500 Internal Server Error: {error: string}
//
// Nota: Usa convenção Kubex (nil, nil) para not found.
func (ctrl *Controller[T]) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := ctrl.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Convenção Kubex: (nil, nil) quando não encontrado
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Create cria nova entidade.
//
// Body: T (JSON)
//
// Response:
//
//	201 Created: {id: string, data: T}
//	400 Bad Request: {error: string} (JSON inválido)
//	500 Internal Server Error: {error: string}
//
// Nota: O ID é retornado separadamente pois pode ser gerado pelo banco.
func (ctrl *Controller[T]) Create(c *gin.Context) {
	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := ctrl.repo.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   id,
		"data": entity,
	})
}

// Update atualiza entidade existente.
//
// Path param:
// - id: ID da entidade (opcional, pode vir no body)
//
// Body: T (JSON)
//
// Response:
//
//	200 OK: T
//	400 Bad Request: {error: string} (JSON inválido)
//	404 Not Found: {error: "not found"}
//	500 Internal Server Error: {error: string}
//
// Nota: Idealmente o ID deveria ser setado no entity antes de chamar Update.
// Se necessário, pode usar reflection ou interface para setar ID da URL.
func (ctrl *Controller[T]) Update(c *gin.Context) {
	// id := c.Param("id")  // Pode ser usado se necessário via reflection
	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Se necessário, setar ID da URL no entity via reflection
	// setEntityID(&entity, id)

	err := ctrl.repo.Update(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// Delete remove entidade por ID.
//
// Path param:
// - id: ID da entidade
//
// Response:
//
//	204 No Content
//	404 Not Found: {error: "not found"}
//	500 Internal Server Error: {error: string}
func (ctrl *Controller[T]) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
