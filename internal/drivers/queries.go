package drivers

const (
	QueryCreatePvz = `
	INSERT INTO pvz_driver (id, registration_date, city)
	VALUES ($1, $2, $3)
`
	QueryCreateReception = `
	INSERT INTO reception_driver (id, reception_time, pvz_id, state)
	Values ($1, $2, $3, $4)
`
	QueryGetReception = `
	SELECT
	    reception_time, 
	    pvz_id, 
	    state
	FROM reception_driver
	WHERE id = $1
`
	QueryGetReceptionInProgressId = `
	SELECT id
	FROM reception_driver
	WHERE pvz_id = $1 AND status = 'in_progress'
	ORDER BY reception_time DESC
	LIMIT 1
	FOR UPDATE
`
	QueryCreateProduct = `
	INSERT INTO product_driver (id, adding_time, product_type, reception_id)
	VALUES ($1, $2, $3, $4)
`
	QueryDeleteLastProduct = `
	DELETE FROM product_driver
	WHERE id IN (
		SELECT id 
		FROM product_driver
		WHERE reception_id = $1
		ORDER BY adding_time DESC
		LIMIT 1
	)
`
	QueryCloseReception = `
	UPDATE reception_driver
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
	LEFT JOIN reception_driver r ON p.id = r.pvz_id
	LEFT JOIN product_driver pr ON r.id = pr.reception_id
`
	QueryGetLastReceptionStatus = `
	SELECT status
	FROM reception_driver
	WHERE pvz_id = $1
	ORDER BY reception_time DESC
	LIMIT 1
`
	QueryCreateUser = `
	INSERT INTO user (id, email, password_hash, user_role)
	VALUES ($1, $2, $3, $4)
`
	QueryGetUserByEmail = `
	SELECT id, email, password_hash, user_role
	FROM user
	WHERE email = $1
`
)
