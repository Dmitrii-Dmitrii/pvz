package db_driver

const (
	queryCreatePvz = `
	INSERT INTO pvz (id, register_date, city)
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
	WHERE pvz_id = $1 AND status = 'in_progress'
	ORDER BY reception_time DESC
	LIMIT 1
	FOR UPDATE
`
	queryCreateProduct = `
	INSERT INTO products (id, adding_time, product_type, reception_id)
	VALUES ($1, $2, $3, $4)
`
	queryDeleteLastProduct = `
	DELETE FROM products
	WHERE id IN (
		SELECT id 
		FROM products
		WHERE reception_id = $1
		ORDER BY adding_time DESC
		LIMIT 1
	)
`
	queryCloseReception = `
	UPDATE receptions
	SET status = 'close'
	WHERE id = $1
`
	queryGetPvz = `
	SELECT
	    id, 
	    register_date, 
	    city
	FROM pvz
	ORDER BY id
	LIMIT $1 OFFSET $2
`
	queryGetPvzWithReceptionInterval = `
	SELECT DISTINCT
	    p.id, 
	    p.register_date, 
	    p.city
	FROM pvz p
	    JOIN receptions r ON p.id = r.pvz_id
	WHERE r.reception_time BETWEEN $1 AND $2
	ORDER BY p.id
	LIMIT $3 OFFSET $4
`
)
