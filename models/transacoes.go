package models

import (
	"context"
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

func (m *TransacaoModel) Inserir(tx pgx.Tx, ctx context.Context, idCliente int, valor int, tipo string, descricao string) error {
	stmt := `INSERT INTO transacoes (idCliente, valor, tipo, descricao, dataCriacao) VALUES($1, $2, $3, $4, $5)`

	_, err := tx.Exec(ctx, stmt, idCliente, valor, tipo, descricao, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (m *TransacaoModel) UltimasTransacoesCliente(tx pgx.Tx, ctx context.Context, idCliente int) ([]*Transacao, error) {
	stmt := `SELECT id, idCliente, valor, tipo, descricao, dataCriacao FROM transacoes
	WHERE idCliente = $1 ORDER BY dataCriacao DESC LIMIT 10`

	rows, err := tx.Query(ctx, stmt, idCliente)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	transacoes := []*Transacao{}

	for rows.Next() {
		transacao := &Transacao{}

		err = rows.Scan(&transacao.ID, &transacao.IdCliente, &transacao.Valor, &transacao.Tipo, &transacao.Descricao, &transacao.DataCriacao)
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
