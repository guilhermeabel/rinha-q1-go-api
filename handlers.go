package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"
)

func (app *application) criarTransacao(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	valor := r.PostForm.Get("valor")
	tipo := r.PostForm.Get("tipo")
	descricao := r.PostForm.Get("descricao")

	// VALIDAÇÃO
	// [id] (na URL) deve ser um número inteiro representando a identificação do cliente.
	idCliente := 0
	_, err = fmt.Sscanf(r.URL.Path, "/clientes/%d/transacoes", &idCliente)

	if err != nil || idCliente < 1 {
		http.NotFound(w, r)
		return
	}

	// valor deve ser um número inteiro positivo que representa centavos (não vamos trabalhar com frações de centavos). Por exemplo, R$ 10 são 1000 centavos.
	valorNumerico, err := strconv.Atoi(valor)
	if err != nil || valorNumerico < 1 {
		http.Error(w, "Erro: campo valor invalido", http.StatusBadRequest)
		return
	}
	// tipo deve ser apenas c para crédito ou d para débito.
	if tipo != "c" && tipo != "d" {
		http.Error(w, "Erro: campo tipo invalido", http.StatusBadRequest)
		return
	}
	// descricao deve ser uma string de 1 a 10 caracteres.
	if utf8.RuneCountInString(descricao) <= 0 || utf8.RuneCountInString(descricao) > 10 {
		http.Error(w, "Erro: campo descricao invalido", http.StatusBadRequest)
		return
	}
	// Todos os campos são obrigatórios.

	tx, err := app.db.Begin()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	cliente, err := app.clientes.Obter(idCliente)
	if err != nil {
		tx.Rollback()
		http.NotFound(w, r)
		return
	}

	_, err = app.transacoes.Inserir(idCliente, valorNumerico, tipo, descricao)
	if err != nil {
		tx.Rollback()
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if tipo == "d" && cliente.Saldo-valorNumerico < -cliente.Limite {
		tx.Rollback()
		http.Error(w, "Limite excedido", http.StatusUnprocessableEntity)
		return
	}

	saldoAtualizado := 0
	if tipo == "c" {
		saldoAtualizado = cliente.Saldo + valorNumerico
	}

	if tipo == "d" {
		saldoAtualizado = cliente.Saldo - valorNumerico
	}

	_, err = app.clientes.Atualizar(cliente.ID, saldoAtualizado, cliente.Limite)
	if err != nil {
		tx.Rollback()
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"limite":%d,"saldo":%d}`, cliente.Limite, saldoAtualizado)
	w.Write([]byte(response))
}

func (app *application) obterExtrato(w http.ResponseWriter, r *http.Request) {
	idCliente := 0
	_, err := fmt.Sscanf(r.URL.Path, "/clientes/%d/extrato", &idCliente)

	if err != nil || idCliente < 1 {
		http.NotFound(w, r)
		return
	}

	tx, err := app.db.Begin()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	cliente, err := app.clientes.Obter(idCliente)
	if err != nil {
		tx.Rollback()
		http.NotFound(w, r)
		return
	}

	transacoes, err := app.transacoes.UltimasTransacoesCliente(idCliente)
	if err != nil {
		tx.Rollback()
		fmt.Printf("Erro: %v\n", err)

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tx.Commit()

	dataExtrato := time.Now().Format(time.RFC3339Nano)

	saldo := map[string]interface{}{
		"total":        cliente.Saldo,
		"data_extrato": dataExtrato,
		"limite":       cliente.Limite,
	}

	ultimasTransacoes := []map[string]interface{}{}
	for _, t := range transacoes {
		transacao := map[string]interface{}{
			"valor":        t.Valor,
			"tipo":         t.Tipo,
			"descricao":    t.Descricao,
			"realizada_em": t.DataCriacao.Format(time.RFC3339Nano),
		}
		ultimasTransacoes = append(ultimasTransacoes, transacao)
	}

	responseMap := map[string]interface{}{
		"saldo":              saldo,
		"ultimas_transacoes": ultimasTransacoes,
	}

	responseJSON, err := json.Marshal(responseMap)
	if err != nil {
		// Handle error
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)

	// RESPOSTA
	// {
	// 	"saldo": {
	// 	  "total": -9098,
	// 	  "data_extrato": "2024-01-17T02:34:41.217753Z",
	// 	  "limite": 100000
	// 	},
	// 	"ultimas_transacoes": [
	// 	  {
	// 		"valor": 10,
	// 		"tipo": "c",
	// 		"descricao": "descricao",
	// 		"realizada_em": "2024-01-17T02:34:38.543030Z"
	// 	  },
	// 	  {
	// 		"valor": 90000,
	// 		"tipo": "d",
	// 		"descricao": "descricao",
	// 		"realizada_em": "2024-01-17T02:34:38.543030Z"
	// 	  }
	// 	]
	//   }

	// saldo
	// total deve ser o saldo total atual do cliente (não apenas das últimas transações seguintes exibidas).
	// data_extrato deve ser a data/hora da consulta do extrato.
	// limite deve ser o limite cadastrado do cliente.
	// ultimas_transacoes é uma lista ordenada por data/hora das transações de forma decrescente contendo até as 10 últimas transações com o seguinte:
	// valor deve ser o valor da transação.
	// tipo deve ser c para crédito e d para débito.
	// descricao deve ser a descrição informada durante a transação.
	// realizada_em deve ser a data/hora da realização da transação.

	// Regras Se o atributo [id] da URL for de uma identificação não existente de cliente, a API deve retornar HTTP Status Code 404. O corpo da resposta nesse caso não será testado e você pode escolher como o representar. Já sabe o que acontece se sua API retornar algo na faixa 2XX, né? Agradecido.

}
