package auth

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func pgxErrNoRows() error { return pgx.ErrNoRows }

func fmtInt(v int64) string { return strconv.FormatInt(v, 10) }

func parseSubject(sub string) (int64, error) {
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("auth: bad subject: %w", err)
	}
	return id, nil
}
