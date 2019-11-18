package driver

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq" // load psql driver
	"github.com/migotom/mt-bulk/internal/entities"
)

const maxRetries = 3

type sqlDB struct {
	conn     *sql.DB
	dbConfig *DBConfig
}

func (d *sqlDB) connect() (err error) {
	d.conn, err = sql.Open(d.dbConfig.Driver, d.dbConfig.Params)
	return
}

type retryFunc func() error

// retry tries to process operation and reconnects if got mysql error.
func (d *sqlDB) retry(operation retryFunc) (err error) {
	for retries := 0; retries < maxRetries; retries++ {
		if err = operation(); err != nil {
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

// Query database and return rows as result of query.
func (d *sqlDB) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	err = d.retry(func() error {
		rows, err = d.conn.Query(query, args...)
		return err
	})
	return
}

// Execute query on database and return result.
func (d *sqlDB) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	err = d.retry(func() error {
		result, err = d.conn.Exec(query, args...)
		return err
	})
	return
}

func getDB(dbConfig *DBConfig) *sqlDB {
	db, ok := dbConfig.Connection.(*sqlDB)
	if !ok {
		db = &sqlDB{}
		db.dbConfig = dbConfig
		dbConfig.Connection = db
	}
	return db
}

// DBCleaner closes DB connection.
func DBCleaner(dbConfig *DBConfig) {
	if db, ok := dbConfig.Connection.(*sqlDB); ok {
		db.conn.Close()
	}
}

// DBSqlLoadJobs loads list of jobs from database.
func DBSqlLoadJobs(ctx context.Context, jobTemplate entities.Job, dbConfig *DBConfig) ([]entities.Job, error) {
	db := getDB(dbConfig)
	if err := db.connect(); err != nil {
		return nil, err
	}

	var jobs []entities.Job

	rows, err := db.Query(dbConfig.Queries.GetDevices, dbConfig.IDserver)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// TODO add ctx.Done check for very long running queries
	for rows.Next() {
		job := jobTemplate
		job.Host = entities.Host{}

		if err = rows.Scan(&job.Host.ID, &job.Host.IP); err != nil {
			return nil, err
		}

		if err := job.Host.Parse(); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

// DBConfig defines database connection settings.
type DBConfig struct {
	Driver     string      `toml:"driver" yaml:"driver"`
	Params     string      `toml:"params" yaml:"params"`
	Connection interface{} `toml:"-" yaml:"-"`
	IDserver   int         `toml:"id_server" yaml:"id_server"`
	Queries    DBQueries   `toml:"queries" yaml:"queries"`
}

// DBQueries defines list of database queries.
type DBQueries struct {
	GetDevices   string `toml:"get_devices" yaml:"get_devices"`
	UpdateDevice string `toml:"update_device" yaml:"update_device"`
}
