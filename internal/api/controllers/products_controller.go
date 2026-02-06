// Package products provides the controller for managing products in the application.
package controllers

import (
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	// mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type ProductController struct {
	productService svc.Service[any]
	APIWrapper     *t.APIWrapper[any]
}

func NewProductController(bridge svc.Service[any]) *ProductController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &ProductController{
		productService: nil,
		APIWrapper:     t.NewAPIWrapper[any](),
	}
}

// // GetAllProducts retorna todos os produtos.
// //
// // @Summary     Listar produtos
// // @Description Recupera a lista de produtos.
// // @Tags        products
// // @Security    BearerAuth
// // @Produce     json
// // @Success     200 {array} mdl.Products
// // @Failure     401 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/products [get]
// func (pc *ProductController) GetAllProducts(c *gin.Context) {
// 	items, err := pc.productService.List()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, items)
// }

// // GetProductByID retorna um produto pelo ID.
// //
// // @Summary     Obter produto
// // @Description Busca um produto específico pelo ID.
// // @Tags        products
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "ID do Produto"
// // @Success     200 {object} mdl.Products
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Router      /api/v1/products/{id} [get]
// func (pc *ProductController) GetProductByID(c *gin.Context) {
// 	id := c.Param("id")
// 	item, err := pc.productService.GetByID(id)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, item)
// }

// // CreateProduct cria um novo produto.
// //
// // @Summary     Criar produto
// // @Description Adiciona um novo produto.
// // @Tags        products
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       payload body mdl.Products true "Dados do Produto"
// // @Success     201 {object} mdl.Products
// // @Failure     400 {object} ErrorResponse
// // @Failure     401 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/products [post]
// func (pc *ProductController) CreateProduct(c *gin.Context) {
// 	var request mdl.Products
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	createdItem, err := pc.productService.Create(&request)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusCreated, createdItem)
// }

// // UpdateProduct atualiza um produto.
// //
// // @Summary     Atualizar produto
// // @Description Atualiza os dados de um produto.
// // @Tags        products
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       id      path string            true "ID do Produto"
// // @Param       payload body mdl.Products true "Dados atualizados"
// // @Success     200 {object} mdl.Products
// // @Failure     400 {object} ErrorResponse
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/products/{id} [put]
// func (pc *ProductController) UpdateProduct(c *gin.Context) {
// 	var request mdl.Products
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	updatedItem, err := pc.productService.Update(&request)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, updatedItem)
// }

// // DeleteProduct remove um produto.
// //
// // @Summary     Remover produto
// // @Description Remove um produto.
// // @Tags        products
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "ID do Produto"
// // @Success     204
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/products/{id} [delete]
// func (pc *ProductController) DeleteProduct(c *gin.Context) {
// 	id := c.Param("id")
// 	if err := pc.productService.Delete(id); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.Status(http.StatusNoContent)
// }
