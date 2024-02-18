package main

import (
	"net/http"
)

func (app *application) routes() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("POST /clientes/{id}/transacoes", app.criarTransacao)
	router.HandleFunc("GET /clientes/{id}/extrato", app.obterExtrato)

	return router
}
