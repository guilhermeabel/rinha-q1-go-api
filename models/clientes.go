package models

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Cliente struct {
	ID     int
	Limite int
	Saldo  int
}

type ClienteModel struct {
	DB *pgxpool.Pool
}

func (m *ClienteModel) Atualizar(tx pgx.Tx, ctx context.Context, id int, saldo int) (int, error) {
	stmt := `UPDATE clientes SET saldo = $1 WHERE id = $2`

	result, err := tx.Exec(ctx, stmt, saldo, id)
	if err != nil {
		return 0, err
	}

	rows := result.RowsAffected()

	return int(rows), nil

}

func (m *ClienteModel) Obter(tx pgx.Tx, ctx context.Context, id int) (*Cliente, error) {
	stmt := `SELECT id, limite, saldo FROM clientes WHERE id = $1 LIMIT 1`

	row := tx.QueryRow(ctx, stmt, id)
	cliente := &Cliente{}

	err := row.Scan(&cliente.ID, &cliente.Limite, &cliente.Saldo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}

		return nil, err
	}

	return cliente, nil
}
