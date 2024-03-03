CREATE TABLE IF NOT EXISTS clientes (
	id INT AUTO_INCREMENT PRIMARY KEY,
	nome VARCHAR(255) NOT NULL,
	limite INT NOT NULL,
	saldo INT NOT NULL DEFAULT 0 
)

CREATE TABLE IF NOT EXISTS transacoes (
	id INT AUTO_INCREMENT PRIMARY KEY,
	idCliente INT NOT NULL,
	valor INT NOT NULL,
	tipo VARCHAR(1) NOT NULL,
	descricao VARCHAR(255) NOT NULL,
	dataCriacao DATETIME NOT NULL
)

INSERT INTO
  clientes (nome, limite)
VALUES
  ('o barato sai caro', 1000 * 100),
  ('zan corp ltda', 800 * 100),
  ('les cruders', 10000 * 100),
  ('padaria joia de cocaia', 100000 * 100),
  ('kid mais', 5000 * 100);

