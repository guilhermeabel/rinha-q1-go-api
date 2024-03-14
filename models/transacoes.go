package models

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transacao struct {
	ID          int
	IdCliente   int
	Valor       int
	Tipo        string
	Descricao   string
	DataCriacao time.Time
}

type TransacaoModel struct {
	DB *pgxpool.Pool
}

func (m *TransacaoModel) LockTable(tx pgx.Tx, ctx context.Context, idCliente int) error {
	_, err := tx.Exec(ctx, "LOCK TABLE transacoes_"+strconv.Itoa(idCliente)+" IN EXCLUSIVE MODE")
	if err != nil {
		return err
	}

	return nil
}

func (m *TransacaoModel) Inserir(tx pgx.Tx, ctx context.Context, idCliente int, valor int, tipo string, descricao string) error {
	stmt := `INSERT INTO transacoes_` + strconv.Itoa(idCliente)
	stmt = stmt + ` (valor, tipo, descricao, dataCriacao) VALUES($1, $2, $3, $4)`

	_, err := tx.Exec(ctx, stmt, valor, tipo, descricao, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (m *TransacaoModel) UltimasTransacoesCliente(tx pgx.Tx, ctx context.Context, idCliente int) ([]*Transacao, error) {
	stmt := `SELECT id, valor, tipo, descricao, dataCriacao FROM transacoes_` + strconv.Itoa(idCliente)
	stmt = stmt + ` ORDER BY dataCriacao DESC LIMIT 10`

	rows, err := tx.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	transacoes := []*Transacao{}

	for rows.Next() {
		transacao := &Transacao{}
		transacao.IdCliente = idCliente

		err = rows.Scan(&transacao.ID, &transacao.Valor, &transacao.Tipo, &transacao.Descricao, &transacao.DataCriacao)
		if err != nil {
			return nil, err
		}

		transacoes = append(transacoes, transacao)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transacoes, nil
}

func (m *TransacaoModel) LimparTransacoesAntigas(ctx context.Context) error {
	stmt := `CALL delete_old_transacoes()`

	_, err := m.DB.Exec(ctx, stmt)
	if err != nil {
		return err
	}

	return nil
}
