package repository

import (
	"context"
	"errors"
	"osush/internal/db"
)

func CreateDictionary(ctx context.Context, DB *db.LoggedDB, nameTm, nameRu string) (*int, error) {
	var id int

	insertDict := `
				INSERT INTO dictionary (tm, ru)
				VALUES ($1,$2)
				RETURNING id
			`

	err := DB.QueryRowContext(ctx, insertDict, nameTm, nameRu).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func GetDictionaryByID(ctx context.Context, DB *db.LoggedDB, id int) (*string, *string, error) {
	var tm, ru string

	insertDict := `
				SELECT 
					tm, ru 
				FROM dictionary
				WHERE id = $1
			`

	err := DB.QueryRowContext(ctx, insertDict, id).Scan(&tm, &ru)
	if err != nil {
		return nil, nil, err
	}

	return &tm, &ru, nil
}

func UpdateDictionary(ctx context.Context, DB *db.LoggedDB, id int, nameTm, nameRu string) error {

	updateDict := `
		UPDATE dictionary
		SET
			tm = $1,
			ru = $2
		WHERE id = $3
	`

	_, err := DB.ExecContext(ctx, updateDict, nameTm, nameRu, id)

	if err != nil {
		return errors.New("failed update dictionary data")
	}

	return nil
}

func DeleteDictionary(ctx context.Context, DB *db.LoggedDB, id int) error {

	deleteDict := `
		DELETE FROM dictionary
		WHERE id = $1
	`

	_, err := DB.ExecContext(
		ctx,
		deleteDict,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}
