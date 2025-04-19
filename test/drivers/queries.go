package drivers

const (
	createSchema = `
	CREATE TYPE city AS enum (
		'Москва',
		'Санкт-Петербург',
		'Казань'
		);
	
	CREATE TYPE reception_status AS enum (
		'in_progress',
		'close'
		);
	
	CREATE TYPE product_type AS enum (
		'электроника',
		'одежда',
		'обувь'
		);
	
	CREATE TYPE user_role AS enum (
		'employee',
		'moderator'
		);

	CREATE TABLE IF NOT EXISTS pvz
	(
		id                UUID PRIMARY KEY,
		registration_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		city              city      NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS receptions
	(
		id             UUID PRIMARY KEY,
		reception_time TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
		pvz_id         UUID             NOT NULL,
		status         reception_status NOT NULL,
		FOREIGN KEY (pvz_id) REFERENCES pvz (id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS products
	(
		id           UUID PRIMARY KEY,
		adding_time  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
		product_type product_type NOT NULL,
		reception_id UUID         NOT NULL,
		FOREIGN KEY (reception_id) REFERENCES receptions (id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS  users
	(
		id UUID PRIMARY KEY,
		email VARCHAR(254) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role user_role NOT NULL
	);
`
	queryCreatePvz = `
	INSERT INTO pvz (id, registration_date, city) 
	VALUES ($1, $2, $3)
`
	queryCreateReception = `
	INSERT INTO receptions (id, reception_time, pvz_id, status) 
	VALUES ($1, $2, $3, $4)
`
	queryCreateProduct = `
	INSERT INTO products (id, adding_time, product_type, reception_id) 
	VALUES ($1, $2, $3, $4)
`
	queryGetPvz = `
	SELECT registration_date, city 
	FROM pvz 
	WHERE id = $1
`
	queryGetReception = `
	SELECT reception_time, pvz_id, status
	FROM receptions 
	WHERE id = $1
`
)
