package dbutil

/*
timezone package
OracleのTimestampやMySQLのDatetimeのようにTZを持たないものはgolang time.Localのtimezoneになる
SQLServerの場合はUTCで出てくる
それらの日付はServerのTimezoneと合わせるべきなので、設定にあるtimezoneへと書き変える
*/
import (
	"reflect"
	"time"
)

// TimeISODB SQLとのやり取りに使う標準的なフォーマット
// Timezone情報が省略されているため、接続時の補正をかける
var TimeISODB = "2006-01-02 15:04:05"

// GetTimeField 任意の構造体におけるtime.Timeカラムを抽出
func GetTimeField(t interface{}) ([]string, error) {

	var timeFields []string

	// Reflectで構造を取得
	r := reflect.ValueOf(t)
	// struct field名の取得用
	ri := reflect.Indirect(r)

	for i := 0; i < r.NumField(); i++ {
		if r.Field(i).Type() == reflect.TypeOf(time.Time{}) {
			timeFields = append(timeFields, ri.Type().Field(i).Name)
		}
	}
	return timeFields, nil
}

// ParseISOTime is 日付データの変換。UTCとして扱われる
func ParseISOTime(t string) (time.Time, error) {
	return time.Parse(TimeISODB, t)
}

// FormatISOTime is stiringに変換。timezone情報は消失する
func FormatISOTime(t time.Time) string {
	return t.Format(TimeISODB)
}

// RewriteTimezone ServerのTimezone情報をもとに出力されたデータのTimeZoneの書き換えを行う
// 時間を変えずにLocationを変更する(=UNIXTimeから見るとオフセット分変化する)
func RewriteTimezone(pointer interface{}, fields []string, tzAfter *time.Location) error {
	rp := reflect.ValueOf(pointer)
	for _, vField := range fields {
		pt, err := time.ParseInLocation(TimeISODB, rp.Elem().FieldByName(vField).Interface().(time.Time).Format(TimeISODB), tzAfter)
		if err != nil {
			return err
		}
		rp.Elem().FieldByName(vField).Set(reflect.ValueOf(pt))
	}
	return nil
}

// LocaleIn 渡された構造体内の任意のFieldのlocaleを変更する
// Locale変更に伴い時間表示が変わる(UTC上は同じ)
func LocaleIn(pointer interface{}, fields []string, tzAfter *time.Location) error {
	rp := reflect.ValueOf(pointer)
	for _, vField := range fields {
		pt := rp.Elem().FieldByName(vField).Interface().(time.Time)
		rp.Elem().FieldByName(vField).Set(reflect.ValueOf(pt.In(tzAfter)))
	}
	return nil
}
