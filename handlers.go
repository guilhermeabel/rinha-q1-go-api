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
	type transactionRequest = struct {
		Valor     int    `json:"valor"`
		Tipo      string `json:"tipo"`
		Descricao string `json:"descricao"`
	}

	defer r.Body.Close()
	var tr transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&tr); err != nil {
		http.Error(w, "Erro ao decodificar body", http.StatusUnprocessableEntity)
		return
	}

	idClienteStr := r.PathValue("id")
	idCliente, err := strconv.Atoi(idClienteStr)

	if err != nil || idCliente < 1 || idCliente > 5 {
		http.NotFound(w, r)
		return
	}

	if tr.Valor < 1 {
		http.Error(w, "Erro: campo valor invalido", http.StatusUnprocessableEntity)
		return
	}

	if tr.Tipo != "c" && tr.Tipo != "d" {
		http.Error(w, "Erro: campo tipo invalido", http.StatusUnprocessableEntity)
		return
	}

	var comprimentoDescricao = utf8.RuneCountInString(tr.Descricao)
	if comprimentoDescricao < 1 || comprimentoDescricao > 10 {
		http.Error(w, "Erro: campo descricao invalido", http.StatusUnprocessableEntity)
		return
	}

	var saldoAtualizado int
	var success bool
	var limite int

	if tr.Tipo == "c" {
		err = app.db.QueryRow(r.Context(), "SELECT * FROM creditar($1, $2, $3)", idCliente, tr.Valor, tr.Descricao).Scan(&saldoAtualizado, &success, &limite)
	} else {
		err = app.db.QueryRow(r.Context(), "SELECT * FROM debitar($1, $2, $3)", idCliente, tr.Valor, tr.Descricao).Scan(&saldoAtualizado, &success, &limite)
	}
	if err != nil || !success {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"limite":%d,"saldo":%d}`, limite, saldoAtualizado)
	w.Write([]byte(response))
}

func (app *application) obterExtrato(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	idClienteStr := r.PathValue("id")
	idCliente, err := strconv.Atoi(idClienteStr)

	if err != nil || idCliente < 1 {
		http.NotFound(w, r)
		return
	}

	tx, err := app.db.Begin(r.Context())
	if err != nil {
		fmt.Printf("Erro begin transaction [get]: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	cliente, err := app.clientes.Obter(tx, r.Context(), idCliente)
	if err != nil {
		tx.Rollback(r.Context())
		http.NotFound(w, r)
		return
	}

	transacoes, err := app.transacoes.UltimasTransacoesCliente(tx, r.Context(), idCliente)
	if err != nil {
		tx.Rollback(r.Context())
		fmt.Printf("Erro ao obter transações: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tx.Commit(r.Context())

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
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)

}
