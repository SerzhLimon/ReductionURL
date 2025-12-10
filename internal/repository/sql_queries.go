package repository

const (
	querySetURL = `
        INSERT INTO reductionurl (originalURL, shortURL) 
        VALUES ($1, $2)
        ON CONFLICT (originalURL) 
        DO NOTHING
    `
	queryGetURL = `
		SELECT originalURL, isDeleted
		FROM reductionurl
		WHERE shortURL = $1
	`
	queryGetArrayURL = `
		SELECT originalURL, shortURL 
		FROM reductionurl;
	`

	queryDeleteURL = `
		UPDATE reductionurl 
		SET isDeleted = true 
		WHERE shortURL = $1;
	`
)
