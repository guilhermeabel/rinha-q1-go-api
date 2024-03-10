CREATE TABLE IF NOT EXISTS clientes (
    id SERIAL PRIMARY KEY,
    limite INTEGER NOT NULL,
    saldo INTEGER NOT NULL DEFAULT 0
);

INSERT INTO clientes (limite)
VALUES
  (100000),
  (80000),
  (1000000),
  (10000000),
  (500000);

CREATE INDEX idCliente_idx ON clientes (id);

DO $$ 
DECLARE
    idCliente INT;
BEGIN
    FOR idCliente IN (select distinct clientes.id from clientes) LOOP
        EXECUTE FORMAT('
			CREATE UNLOGGED TABLE "transacoes_%s" (
				id serial,
				valor int NOT NULL,
				tipo varchar(1) NOT NULL,
				descricao varchar(10) NOT NULL,
				dataCriacao timestamp DEFAULT now()
			);
			CREATE INDEX transacoes_cliente_%s_idx ON "transacoes_%s" (id DESC);
        ', idCliente, idCliente, idCliente);
    END LOOP;
END $$;

CREATE OR REPLACE
PROCEDURE delete_old_transacoes()
LANGUAGE plpgsql
AS $$
DECLARE
	idCliente INT;
	tableName TEXT;
BEGIN
	FOR idCliente IN (select distinct clientes.id from clientes) LOOP
		tableName := format('transacoes_%s', idCliente);
		EXECUTE format('
			DELETE FROM %s
			WHERE id < (
				SELECT id
				FROM %s
				ORDER BY id DESC
				LIMIT 1 OFFSET 11
			);
		', tableName, tableName);
	END LOOP;
END $$;
