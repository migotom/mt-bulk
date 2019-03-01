package driver

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq" // load psql driver
	"github.com/migotom/mt-bulk/internal/schema"
)

const maxRetries = 3

type sqlDB struct {
	conn     *sql.DB
	dbConfig *schema.DBConfig
}

func (d *sqlDB) connect() (err error) {
	d.conn, err = sql.Open(d.dbConfig.Driver, d.dbConfig.Params)
	return
}

type retryFunc func() error

func (d *sqlDB) retry(retryFunc retryFunc) (err error) {
	for retries := 0; retries < maxRetries; retries++ {
		err = retryFunc()
		if err != nil {
			// cleanup
			d.conn.Close()

			// reconnect and retry
			time.Sleep(1000 * time.Millisecond)
			d.connect()
			continue
		}
	}
	return
}

func (d *sqlDB) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	err = d.retry(func() error {
		rows, err = d.conn.Query(query, args...)
		return err
	})
	return
}

func (d *sqlDB) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	err = d.retry(func() error {
		result, err = d.conn.Exec(query, args...)
		return err
	})
	return
}

func getDB(dbConfig *schema.DBConfig) *sqlDB {
	db, ok := dbConfig.Connection.(*sqlDB)
	if !ok {
		db = &sqlDB{}
		db.dbConfig = dbConfig
		dbConfig.Connection = db
	}
	return db
}

// DBCleaner closes DB connection.
func DBCleaner(dbConfig *schema.DBConfig) {
	if db, ok := dbConfig.Connection.(*sqlDB); ok {
		db.conn.Close()
	}
}

// DBSqlLoadHosts loads list of hosts from database.
func DBSqlLoadHosts(hostParser schema.HostParserFunc, dbConfig *schema.DBConfig) ([]schema.Host, error) {
	db := getDB(dbConfig)
	if err := db.connect(); err != nil {
		return nil, err
	}

	var hosts []schema.Host

	rows, err := db.Query(dbConfig.Queries.GetDevices, dbConfig.IDserver)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		host := schema.Host{}

		err = rows.Scan(&host.ID, &host.IP)
		if err != nil {
			return nil, err
		}
		host, err = hostParser(host)
		if err != nil {
			return nil, err
		}

		hosts = append(hosts, host)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return hosts, nil
}
