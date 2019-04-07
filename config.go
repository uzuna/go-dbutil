package dbutil

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	gorp "gopkg.in/gorp.v2"
	yaml "gopkg.in/yaml.v2"
)

// DBConfigure Database Configuration Interface
type DBConfigure interface {
	GetName() string                      // connection identity
	GetDSN() string                       // connection DSN
	GetDriver() string                    // driver
	GetLocation() (*time.Location, error) // DB system location for use timezone offset
	GetMaxOpenConns() int                 // DataBase max open connection
	GetMaxIdleConns() int                 // DataBase max idle connection
	GetMaxLifetime() time.Duration        // DataBase max lifetime of connection
	GenInstance() (*sql.DB, error)        // Generate DB Instance
	WrapDbMap(db *sql.DB) *gorp.DbMap     // Wrap to gorp.DbMap
}

// DBConfig is Dbへの接続情報 + リミット情報
type DBConfig struct {
	Name              string                 `yaml:"name"`
	Driver            string                 `yaml:"driver"`
	Hostname          string                 `yaml:"host"`
	Username          string                 `yaml:"user"`
	Password          string                 `yaml:"pass"`
	Database          string                 `yaml:"database"`
	Timezone          string                 `yaml:"tz"`
	MaxOpenConnection int                    `yaml:"max_open_conn"`
	MaxIdleConnection int                    `yaml:"max_idle_conn"`
	MaxLifetime       int                    `yaml:"max_lifetime"` // max lifetime(msec)
	Parameters        map[string]interface{} `yaml:"params"`
}

type Values map[string][]string

// Set sets the key to value. It replaces any existing
// values.
func (v Values) Set(key, value string) {
	v[key] = []string{value}
}

// Add adds the value to key. It appends to any existing
// values associated with key.
func (v Values) Add(key, value string) {
	v[key] = append(v[key], value)
}

func (v Values) Encode() string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		keyEscaped := k
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}

type ConnectionString struct {
	Scheme string
	User   *url.Userinfo
	Host   string
	Path   string
	Params Values
}

func (u *ConnectionString) String() string {
	var buf strings.Builder
	if u.Scheme != "" {
		buf.WriteString(u.Scheme)
		buf.WriteByte(':')
		buf.WriteString("//")
	}
	if ui := u.User; ui != nil {
		buf.WriteString(ui.String())
		buf.WriteByte('@')
	}
	if h := u.Host; h != "" {
		buf.WriteString(u.Host)
	}
	path := u.Path
	if path != "" && path[0] != '/' && u.Host != "" {
		buf.WriteByte('/')
	}
	buf.WriteString(path)
	param := u.Params.Encode()
	if len(param) > 0 {
		buf.WriteByte('?')
		buf.WriteString(param)
	}

	return buf.String()
}

// GetName is Parameterから任意の接続DSNを作成する
func (cc *DBConfig) GetName() string {
	return cc.Name
}

// GetDSN is Parameterから任意の接続DSNを作成する
func (cc *DBConfig) GetDSN() string {
	if cc.Driver == "mysql" {

		query := mysqlDriverOption(cc)

		u := &ConnectionString{
			User:   url.UserPassword(cc.Username, cc.Password),
			Host:   fmt.Sprintf("tcp(%s)", cc.Hostname),
			Path:   cc.Database,
			Params: query,
		}

		return u.String()
	} else if cc.Driver == "oracle" {
		u := &ConnectionString{
			Scheme: cc.Driver,
			User:   url.UserPassword(cc.Username, cc.Password),
			Host:   cc.Hostname,
		}
		return u.String()
	} else if cc.Driver == "sqlserver" {
		query := Values{}
		if len(cc.Database) > 0 {
			query.Add("database", cc.Database)
		}
		u := &ConnectionString{
			Scheme: cc.Driver,
			User:   url.UserPassword(cc.Username, cc.Password),
			Host:   cc.Hostname,
			Params: query,
		}
		return u.String()
	}
	panic(fmt.Sprintf("Unsupported driver [%s]", cc.Driver))
}

// GetDriver is driver nameから実際にライブラリで使うdriver名に変換
func (cc *DBConfig) GetDriver() string {
	if cc.Driver == "mysql" {
		return "mysql"
	} else if cc.Driver == "oracle" {
		return "goracle"
	} else if cc.Driver == "sqlserver" {
		return "sqlserver"
	}
	panic(fmt.Sprintf("Unsupported driver [%s]", cc.Driver))
}

// WrapDbMap is gorp diarect wrapper
func (cc *DBConfig) WrapDbMap(db *sql.DB) *gorp.DbMap {
	return Wrap2Diarect(cc.Driver, db)
}

// GenInstance is generate instance
func (cc *DBConfig) GenInstance() (*sql.DB, error) {
	connectStr := cc.GetDSN()
	db, err := sql.Open(cc.GetDriver(), connectStr)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cc.GetMaxOpenConns())
	db.SetMaxIdleConns(cc.GetMaxIdleConns())
	db.SetConnMaxLifetime(cc.GetMaxLifetime())
	return db, err
}

// GetLocation is 対象のTimezoneを取得
func (cc *DBConfig) GetLocation() (*time.Location, error) {
	return time.LoadLocation(cc.Timezone)
}

// GetMaxOpenConns returns number of max_open_connection
func (cc *DBConfig) GetMaxOpenConns() int {
	return cc.MaxOpenConnection
}

// GetMaxIdleConns returns number of max_idle_connection
func (cc *DBConfig) GetMaxIdleConns() int {
	return cc.MaxIdleConnection
}

// GetMaxLifetime returns number of connection timeout duration
func (cc *DBConfig) GetMaxLifetime() time.Duration {
	return time.Millisecond * time.Duration(cc.MaxLifetime)
}

// ParseConfig dbconfig
func ParseConfig(b []byte) (m *DBConfig, err error) {
	m = &DBConfig{}
	r := bytes.NewReader(b)
	dec := yaml.NewDecoder(r)
	err = dec.Decode(m)
	return
}

// DecodeConfig dbconfig
func DecodeConfig(r io.Reader) (m []*DBConfig, err error) {
	dec := yaml.NewDecoder(r)
	err = dec.Decode(&m)
	if len(m) < 1 {
		err = errors.Errorf("Not found db setting")
		return
	}
	for _, mi := range m {
		err = verify(mi)
		if err != nil {
			return
		}
	}
	return
}

// verify facilitates entry of unset parameters
func verify(c *DBConfig) (err error) {
	if c.MaxOpenConnection == 0 {
		err = errors.Errorf("[Error] %s Max Open Conns is 0(unlimited). Recommended:[Measured or less than the number of CPU cores] forgot set?", c.Name)
	}
	if c.MaxIdleConnection <= 0 {
		fmt.Printf("[WARN] %s Max Idle Conns is 0(not retaind). Recommended:[>=Max Open Conns] forgot set?", c.Name)
	}
	if c.MaxLifetime <= 0 {
		fmt.Printf("[WARN] %s Max Lifetime is 0(connections are reused forever). Recommended:[<=Max Open Conns * 1 second] forgot set?", c.Name)
	}

	return err
}

// Wrap2Diarect is gorp diarectでwrapする
func Wrap2Diarect(driver string, db *sql.DB) *gorp.DbMap {
	if driver == "mysql" {
		return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{
			Engine: "InnoDB", Encoding: "UTF8"}}
	} else if driver == "oracle" {
		return &gorp.DbMap{Db: db, Dialect: gorp.OracleDialect{}}
	} else if driver == "sqlserver" {
		return &gorp.DbMap{Db: db, Dialect: gorp.SqlServerDialect{}}
	}
	panic(fmt.Sprintf("Unsupported driver [%s]", driver))
}
