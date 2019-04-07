package dbutil

import (
	"database/sql"
	"time"

	gorp "gopkg.in/gorp.v2"
)

// New Create repository from DBConfigure
func New(conf DBConfigure) (*Repository, error) {
	db, err := conf.GenInstance()
	if err != nil {
		return nil, err
	}
	// Must has location
	_, err = conf.GetLocation()
	if err != nil {
		return nil, err
	}
	return &Repository{
		db:   db,
		conf: conf,
	}, nil
}

// Repository is DatabaseConnection
type Repository struct {
	db   *sql.DB // 接続instance
	conf DBConfigure
}

// Db is instanceへのポインタを取得
func (r *Repository) Db() *sql.DB {
	return r.db
}

// Dbm is gorpでwrapしたpointerを取得
func (r *Repository) Dbm() *gorp.DbMap {
	return r.conf.WrapDbMap(r.db)
}

// Driver driverの表示 / sqlbuilderに使う
func (r *Repository) Driver() string {
	return r.conf.GetDriver()
}

// Location RepositoryのTimezone情報。データの補正に使う
func (r *Repository) Location() *time.Location {
	tz, _ := r.conf.GetLocation()
	return tz
}

// Close is DBのGraceful Shutdownのため
func (r *Repository) Close() error {
	return r.db.Close()
}
