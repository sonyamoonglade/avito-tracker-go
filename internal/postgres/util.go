package postgres

import (
	"errors"

	"github.com/jackc/pgx/v4"
)

func CheckEmptyRows(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	return err
}
