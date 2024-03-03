package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

func (app *application) criarTransacao(w http.ResponseWriter, r *http.Request) {
	type transactionRequest = struct {
		Valor     string `json:"valor"`
		Tipo      string `json:"tipo"`
		Descricao string `json:"descricao"`
	}

	defer r.Body.Close()
	var tr transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&tr); err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	idClienteStr := r.PathValue("id")
	idCliente, err := strconv.Atoi(idClienteStr)

	if err != nil || idCliente < 1 || idCliente > 5 {
		http.NotFound(w, r)
		return
	}

	valorNumerico, err := strconv.Atoi(tr.Valor)
	if err != nil || valorNumerico < 1 {
		http.Error(w, "Erro: campo valor invalido", http.StatusUnprocessableEntity)
		return
	}

	if tr.Tipo != "c" && tr.Tipo != "d" {
		http.Error(w, "Erro: campo tipo invalido", http.StatusUnprocessableEntity)
		return
	}

	if utf8.RuneCountInString(tr.Descricao) <= 0 || utf8.RuneCountInString(tr.Descricao) > 10 {
		http.Error(w, "Erro: campo descricao invalido", http.StatusUnprocessableEntity)
		return
	}

	tx, err := app.db.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		fmt.Printf("Erro ao iniciar transaction: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	cliente, err := app.clientes.Obter(r.Context(), idCliente)
	if err != nil {
		tx.Rollback(r.Context())
		http.NotFound(w, r)
		return
	}

	err = app.transacoes.Inserir(r.Context(), idCliente, valorNumerico, tr.Tipo, tr.Descricao)
	if err != nil {
		tx.Rollback(r.Context())
		fmt.Printf("Erro ao inserir transação: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	saldoAtualizado := 0
	if tr.Tipo == "c" {
		saldoAtualizado = cliente.Saldo + valorNumerico
	} else {
		saldoAtualizado = cliente.Saldo - valorNumerico
		if saldoAtualizado < -cliente.Limite {
			tx.Rollback(r.Context())
			http.Error(w, "Limite excedido", http.StatusUnprocessableEntity)
			return
		}
	}

	_, err = app.clientes.Atualizar(r.Context(), cliente.ID, saldoAtualizado, cliente.Limite)
	if err != nil {
		tx.Rollback(r.Context())
		fmt.Printf("Erro ao atualizar cliente: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tx.Commit(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"limite":%d,"saldo":%d}`, cliente.Limite, saldoAtualizado)
	w.Write([]byte(response))
}

func (app *application) obterExtrato(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	idCliente := 0
	_, err := fmt.Sscanf(r.URL.Path, "/clientes/%d/extrato", &idCliente)

	if err != nil || idCliente < 1 {
		http.NotFound(w, r)
		return
	}

	tx, err := app.db.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		fmt.Printf("Erro ao iniciar transaction: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	cliente, err := app.clientes.Obter(r.Context(), idCliente)
	if err != nil {
		tx.Rollback(r.Context())
		http.NotFound(w, r)
		return
	}

	transacoes, err := app.transacoes.UltimasTransacoesCliente(r.Context(), idCliente)
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
