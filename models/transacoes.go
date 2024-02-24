package models

import (
	"database/sql"
	"time"
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
	DB *sql.DB
}

func (m *TransacaoModel) Inserir(idCliente int, valor int, tipo string, descricao string) (int, error) {
	stmt := `INSERT INTO transacoes (idCliente, valor, tipo, descricao, dataCriacao) VALUES(?, ?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt, idCliente, valor, tipo, descricao, time.Now().Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *TransacaoModel) UltimasTransacoesCliente(idCliente int) ([]*Transacao, error) {
	stmt := `SELECT id, idCliente, valor, tipo, descricao, dataCriacao FROM transacoes
	WHERE idCliente = ? ORDER BY dataCriacao DESC LIMIT 10`

	rows, err := m.DB.Query(stmt, idCliente)
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
