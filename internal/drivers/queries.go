package drivers

const (
	QueryCreatePvz = `
	INSERT INTO pvz (id, registration_date, city)
	VALUES ($1, $2, $3)
`
	QueryCreateReception = `
	INSERT INTO receptions (id, reception_time, pvz_id, status)
	Values ($1, $2, $3, $4)
`
	QueryGetReception = `
	SELECT
	    reception_time, 
	    pvz_id, 
	    status
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
		p.id, 
		p.registration_date, 
		p.city,
		r.id,
		r.reception_time, 
		r.status, 
		pr.id, 
		pr.adding_time, 
		pr.product_type
	FROM pvz p
	LEFT JOIN receptions r ON p.id = r.pvz_id
	LEFT JOIN products pr ON r.id = pr.reception_id
`
	QueryGetPvzById = `
	SELECT registration_date, city
	FROM pvz
	WHERE id = $1
`
	QueryGetLastReceptionStatus = `
	SELECT status
	FROM receptions
	WHERE pvz_id = $1
	ORDER BY reception_time DESC
	LIMIT 1
`
	QueryCreateUser = `
	INSERT INTO users (id, email, password_hash, role)
	VALUES ($1, $2, $3, $4)
`
	QueryGetUserByEmail = `
	SELECT id, password_hash, role
	FROM users
	WHERE email = $1
`
	QueryGetUserById = `
	SELECT email, password_hash, role
	FROM users
	WHERE id = $1
`
)
