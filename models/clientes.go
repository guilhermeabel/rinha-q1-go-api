package models

import (
	"database/sql"
	"errors"
)

type Cliente struct {
	ID     int
	Nome   string
	Limite int
	Saldo  int
}

type ClienteModel struct {
	DB *sql.DB
}

func (m *ClienteModel) Atualizar(id int, saldo int, limite int) (int, error) {
	stmt := `UPDATE clientes SET saldo = ?, limite = ? WHERE id = ? LIMIT 1`

	result, err := m.DB.Exec(stmt, saldo, limite, id)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rows), nil

}

func (m *ClienteModel) Obter(id int) (*Cliente, error) {
	stmt := `SELECT id, nome, limite, saldo FROM clientes WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

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
