package reporters

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"listenstats/config"
	"strings"
	"time"
)

type PostgresReporter struct {
	db            *sqlx.DB
	minListenTime time.Duration
}

func NewPostgresReporter(config *config.PostgresReporter) (*PostgresReporter, error) {
	var pwd string
	if strings.ContainsAny(config.Password, " ") || len(config.Password) == 0 {
		pwd = "'" + config.Password + "'"
	} else {
		pwd = config.Password
	}
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		pwd,
		config.Database,
	)
	if config.Schema != "" {
		connString += fmt.Sprintf(" search_path=%s", config.Schema)
	}
	db, err := sqlx.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	if config.ResetOnStart {
		_, err = db.Exec("UPDATE listen SET time_end = NOW() WHERE time_start < NOW() AND time_end IS NULL")
		if err != nil {
			return nil, err
		}
	}
	return &PostgresReporter{db: db, minListenTime: config.MinListenTime.Duration}, nil
}

func (r *PostgresReporter) Close() error {
	return r.db.Close()
}

func (r *PostgresReporter) ReportListenStart(clientId string, info *ListenerInfo) error {
	mount := info.ServerURL.Path
	ua := info.Headers.Get("User-Agent")
	_, err := r.db.Exec(
		`INSERT INTO listens.listen (mount, client_id, ip_address, user_agent, time_start, time_end)
        VALUES ($1, $2, $3, $4, NOW(), NULL)`,
		mount,
		clientId,
		info.IP,
		ua,
	)
	return err
}

func (r *PostgresReporter) ReportListenEnd(clientId string, time time.Duration) error {
	var err error
	if time < r.minListenTime {
		_, err = r.db.Exec(
			`DELETE FROM listen
        WHERE client_id = $1`,
			clientId,
		)
	} else {
	}
	_, err = r.db.Exec(
		`UPDATE listen
        SET time_end = time_start + $2
        WHERE client_id = $1`,
		clientId,
		time.String(),
	)
	return err
}
