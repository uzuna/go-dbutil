package dbutil

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/uzuna/go-dbutil/sqlformat"
)

// GetSelectBuilder squirrelで方言別のPlaceholderを用意
func GetSelectBuilder(driver string) (sq.StatementBuilderType, error) {
	var sbpf sq.StatementBuilderType
	switch driver {
	case "mysql":
		sbpf = sq.StatementBuilder
	case "sqlserver":
		sbpf = sq.StatementBuilder.PlaceholderFormat(sqlformat.AtmarkFormat)
	case "oracle":
		sbpf = sq.StatementBuilder.PlaceholderFormat(sqlformat.CollonFormat)
	default:
		return sbpf, errors.Errorf("Un-supported driver [%s]", driver)
	}
	return sbpf, nil
}
