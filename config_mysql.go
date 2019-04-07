package dbutil

import (
	"fmt"
	"net/url"
	"strconv"
)

var mysqlDriverOptions = []string{
	"charset",
	"allowAllFiles",
	"allowCleartextPasswords",
	"allowNativePasswords",
	"allowOldPasswords",
	"clientFoundRows",
	"collation",
	"columnsWithAlias",
	"interpolateParams",
	"multiStatements",
	"readTimeout",
	"rejectReadOnly",
	"serverPubKey",
	"timeout",
	"tls",
	"writeTimeout",
	"maxAllowedPacket",
}

// mysqlDriverOption
func mysqlDriverOption(cc *DBConfig) Values {
	query := Values{}
	query.Add("parseTime", "true")
	query.Add("loc", url.QueryEscape(cc.Timezone))
	query.Add("time_zone", url.QueryEscape(fmt.Sprintf("'%s'", cc.Timezone)))

outer:
	for name, v := range cc.Parameters {
		for _, pname := range mysqlDriverOptions {
			if name == pname {
				switch x := v.(type) {
				case int:
					query.Add(name, strconv.FormatInt(int64(x), 10))
				case bool:
					if x {
						query.Add(name, "true")
					} else {
						query.Add(name, "false")
					}
				case string:
					query.Add(name, x)
				}
				continue outer
			}
		}

		switch x := v.(type) {
		case int:
			query.Add(name, strconv.FormatInt(int64(x), 10))
		case bool:
			if x {
				query.Add(name, "1")
			} else {
				query.Add(name, "0")
			}
		case string:
			query.Add(name, url.QueryEscape(fmt.Sprintf("'%s'", x)))
		}
	}
	return query
}
