package models

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Cliente struct {
	ID     int
	Nome   string
	Limite int
	Saldo  int
}

type ClienteModel struct {
	DB *pgxpool.Pool
}

func (m *ClienteModel) Atualizar(ctx context.Context, id int, saldo int, limite int) (int, error) {
	stmt := `UPDATE clientes SET saldo = $1, limite = $2 WHERE id = $3 LIMIT 1`

	result, err := m.DB.Exec(ctx, stmt, saldo, limite, id)
	if err != nil {
		return 0, err
	}

	rows := result.RowsAffected()

	return int(rows), nil

}

func (m *ClienteModel) Obter(ctx context.Context, id int) (*Cliente, error) {
	stmt := `SELECT id, nome, limite, saldo FROM clientes WHERE id = $1`

	row := m.DB.QueryRow(ctx, stmt, id)

	cliente := &Cliente{}

	err := row.Scan(&cliente.ID, &cliente.Nome, &cliente.Limite, &cliente.Saldo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}

		return nil, err
	}

	return cliente, nil
}
