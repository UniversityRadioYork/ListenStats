package reporters

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	cfg "listenstats/config"
	"listenstats/utils/geoip"
	"log"
	"strings"
	"time"
)

type PostgresReporter struct {
	db            *sqlx.DB
	queuedGeoip   *geoip.QueuedGeoIP
	minListenTime time.Duration
}

func NewPostgresReporter(globalConfig *cfg.Config) (*PostgresReporter, error) {
	config := globalConfig.Postgres
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
	result := &PostgresReporter{db: db, minListenTime: config.MinListenTime.Duration}
	if globalConfig.GeoIPPath != "" {
		result.queuedGeoip, err = geoip.NewQueuedGeoIP(globalConfig.GeoIPPath)
		if err != nil {
			return nil, err
		}
		go result.queuedGeoip.Work()
		go result.handleGeoIpResults()
	}
	return result, nil
}

func (r *PostgresReporter) handleGeoIpResults() {
	for {
		select {
		// We can do the "ok; return" trick on both, because we know they'll get closed simultaneously
		case result, ok := <-r.queuedGeoip.Results:
			if !ok {
				return
			}
			_, err := r.db.Exec(
				`UPDATE listen
			SET geoip_country = $2, geoip_location = $3
			WHERE client_id = $1`,
				result.ClientID,
				result.GeoIPCountry,
				result.GeoIPLocation,
			)
			if err != nil {
				panic(err)
			}
		case err, ok := <-r.queuedGeoip.Errors:
			if !ok {
				return
			}
			log.Printf("WARN: geoIP error %v", err)
		}
	}
}

func (r *PostgresReporter) Close() error {
	if r.queuedGeoip != nil {
		r.queuedGeoip.Close()
	}
	return r.db.Close()
}

func (r *PostgresReporter) ReportListenStart(clientId string, info *ListenerInfo) error {
	mount := info.ServerURL.Path
	ua := info.Headers.Get("User-Agent")
	referrer := info.Headers.Get("Referer")
	_, err := r.db.Exec(
		`INSERT INTO listens.listen (mount, client_id, ip_address, user_agent, referrer, time_start, time_end)
        VALUES ($1, $2, $3, $4, $5, NOW(), NULL)`,
		mount,
		clientId,
		info.IP,
		ua,
		referrer,
	)
	return err
}

func (r *PostgresReporter) ReportGeoIP(clientId string, info *ListenerInfo) {
	r.queuedGeoip.Request(clientId, info.IP)
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
