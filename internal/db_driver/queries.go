package db_driver

const (
	queryCreatePvz = `
	INSERT INTO pvzs (id, register_date, city)
	VALUES ($1, $2, $3)
`
	queryCreateReception = `
	INSERT INTO receptions (id, reception_time, pvz_id, state)
	Values ($1, $2, $3, $4)
`
	queryGetReception = `
	SELECT
	    reception_time, 
	    pvz_id, 
	    state
	FROM receptions
	WHERE id = $1
`
	queryGetReceptionInProgressId = `
	SELECT id
	FROM receptions
	WHERE pvz_id = $1 AND state = 'in_progress'
	ORDER BY reception_time DESC
	LIMIT 1
	FOR UPDATE
`
	queryCreateProduct = `
	INSERT INTO products (id, adding_time, product_type, reception_id)
	VALUES ($1, $2, $3, $4)
`
)
