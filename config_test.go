package dbutil_test

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	dbutil "github.com/uzuna/go-dbutil"
	// drivers
	// _ "github.com/denisenkom/go-mssqldb"
	// _ "github.com/go-sql-driver/mysql"
	// _ "gopkg.in/goracle.v2"
)

func TestReadConfig(t *testing.T) {
	f, err := os.Open("./testdata/config_sample.yml")
	defer f.Close()
	checkError(t, err)
	dbcs, err := dbutil.DecodeConfig(f)
	checkError(t, err)

	type Expect struct {
		ConnectionString string
	}
	table := []Expect{
		Expect{
			ConnectionString: "oracle://user:pass@localhost/oracle",
		},
		Expect{
			ConnectionString: "sqlserver://user:pass@localhost/sqlserver",
		},
		Expect{
			ConnectionString: "user:pass@tcp(localhost:3306)",
		},
	}

	for i, v := range dbcs {
		td := table[i]
		assert.Contains(t, v.GetDSN(), td.ConnectionString)
	}
}
func TestConnection(t *testing.T) {
	t.SkipNow()
	f, err := os.Open("./testdata/config.yml")
	defer f.Close()
	checkError(t, err)
	dbcs, err := dbutil.DecodeConfig(f)
	checkError(t, err)

	type Expect struct {
		Query  string
		Expect interface{}
	}
	table := []Expect{
		Expect{
			Query:  "SELECT 1 FROM DUAL",
			Expect: 1,
		},
		Expect{
			Query:  "SELECT 1",
			Expect: 1,
		},
		Expect{
			Query:  "SELECT 1",
			Expect: 1,
		},
	}

	for i, v := range dbcs {
		td := table[i]
		log.Println(v.GetDSN())
		_ = td
		db, err := v.GenInstance()
		checkError(t, err)
		rows, err := db.Query(td.Query)
		checkError(t, err)

		// scan
		var x int
		for rows.Next() {
			err = rows.Scan(&x)
			checkError(t, err)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		err = rows.Close()
		checkError(t, err)
		// log.Println(x)
		assert.Equal(t, td.Expect, x)

		err = db.Close()
		checkError(t, err)
	}
}

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Logf("%#v", err)
		t.FailNow()
	}
}
