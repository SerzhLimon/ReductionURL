package repository

const (
	querySetURL = `
		INSERT INTO reductionurl (originalURL, shortURL)
		VALUES ($1, $2)
		ON CONFLICT (originalURL) DO NOTHING;
	`
	queryGetURL = `
		SELECT originalURL
		FROM reductionurl
		WHERE shortURL = $1
	`
)