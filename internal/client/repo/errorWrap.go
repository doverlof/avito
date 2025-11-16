package repo

import "fmt"

func ErrToCreateToCreateSql(err error) error {
	return fmt.Errorf("err to create sql: %w", err)
}
