package drivers

const (
	QueryCreatePvz = `
	INSERT INTO pvzs (id, registration_date, city)
	VALUES ($1, $2, $3)
`
	QueryCreateReception = `
	INSERT INTO receptions (id, reception_time, pvz_id, state)
	Values ($1, $2, $3, $4)
`
	QueryGetReception = `
	SELECT
	    reception_time, 
	    pvz_id, 
	    state
	FROM receptions
	WHERE id = $1
`
	QueryGetReceptionInProgressId = `
	SELECT id
	FROM receptions
	WHERE pvz_id = $1 AND status = 'in_progress'
	ORDER BY reception_time DESC
	LIMIT 1
	FOR UPDATE
`
	QueryCreateProduct = `
	INSERT INTO products (id, adding_time, product_type, reception_id)
	VALUES ($1, $2, $3, $4)
`
	QueryDeleteLastProduct = `
	DELETE FROM products
	WHERE id IN (
		SELECT id 
		FROM products
		WHERE reception_id = $1
		ORDER BY adding_time DESC
		LIMIT 1
	)
`
	QueryCloseReception = `
	UPDATE receptions
	SET status = 'close'
	WHERE id = $1
`
	QueryGetPvz = `
	SELECT
	    id, 
	    registration_date, 
	    city
	FROM pvzs
	ORDER BY id
	LIMIT $1 OFFSET $2
`
	QueryGetPvzWithReceptionInterval = `
	SELECT DISTINCT
	    p.id, 
	    p.registration_date, 
	    p.city
	FROM pvzs p
	    JOIN receptions r ON p.id = r.pvz_id
	WHERE r.reception_time BETWEEN $1 AND $2
	ORDER BY p.id
	LIMIT $3 OFFSET $4
`
)
