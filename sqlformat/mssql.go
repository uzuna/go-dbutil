package sqlformat

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

// AtmarkFormat this is place hosder for mssql driver
// MSSQLDriverは`@p1`という形のplaceholderのため
var AtmarkFormat = atmarkFormat{}

type atmarkFormat struct{}

func (_ atmarkFormat) ReplacePlaceholders(sql string) (string, error) {
	buf := &bytes.Buffer{}
	i := 0
	for {
		p := strings.Index(sql, "?")
		if p == -1 {
			break
		}

		if len(sql[p:]) > 1 && sql[p:p+2] == "??" { // escape ?? => ?
			buf.WriteString(sql[:p])
			buf.WriteString("?")
			if len(sql[p:]) == 1 {
				break
			}
			sql = sql[p+2:]
		} else {
			i++
			buf.WriteString(sql[:p])
			fmt.Fprintf(buf, "@p%d", i)
			sql = sql[p+1:]
		}
	}

	buf.WriteString(sql)
	return buf.String(), nil
}

//TrimSpace CHARでspace paddingされているカラムのトリムを行う
func TrimSpace(t interface{}) error {

	// 書き変え用のpointerを取得
	r := reflect.ValueOf(t)
	// loop用のstruct構造を取得
	ri := reflect.Indirect(r)

	// string型の検知用
	rstr := reflect.TypeOf("")

	// structのパラメータ分だけループ
	for i := 0; i < ri.NumField(); i++ {
		switch ri.Field(i).Type() {
		case rstr:
			fieldName := ri.Type().Field(i).Name
			ts := r.Elem().FieldByName(fieldName).Interface().(string)
			r.Elem().FieldByName(fieldName).Set(reflect.ValueOf(strings.TrimSpace(ts)))
		}
	}
	return nil
}
