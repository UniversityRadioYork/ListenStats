package main

import (
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	lsgeoip "listenstats/utils/geoip"
	"log"
	"runtime"
)

type Result struct {
	ListenID      uint64 `sql:"listen_id"`
	GeoIPCountry  string `sql:"geoip_country"`
	GeoIPLocation string `sql:"geoip_location"`
}

type Row struct {
	ListenID  uint64 `sql:"listen_id"`
	Mount     string `sql:"mount"`
	ClientID  string `sql:"client_id"`
	IPAddr    string `sql:"ip_addr"`
	UserAgent string `sql:"user_agent"`
}

var geoip *lsgeoip.GeoIP
var tasks chan Row
var results = make(chan Result)
var done = make(chan struct{})

func work() {
	//fmt.Println("worker ready")
	for task := range tasks {
		//fmt.Printf("task %v\n", task)
		rez, err := geoip.Process(task.IPAddr)
		if err != nil {
			panic(err)
		}
		result := Result{
			ListenID:      task.ListenID,
			GeoIPCountry:  rez.GeoIPCountry,
			GeoIPLocation: rez.GeoIPLocation,
		}
		//fmt.Printf("pushing %+v\n", result)
		results <- result
	}
	done <- struct{}{}
}

func main() {
	geoipPath := flag.String("geoip-path", "", "Path to GeoIP2 database")

	lt := flag.Uint64("id-min", 0, "Lowest record ID to calculate for")
	gt := flag.Uint64("id-max", 9223372036854775807, "Highest record ID to calculate for")

	dbUrl := flag.String("db-url", "", "postgres:// URL to Postgres DB")

	numWorkers := flag.Int("workers", runtime.NumCPU(), "Number of numWorkers to use")

	flag.Parse()

	if *geoipPath == "" {
		log.Fatal("No geoip-path")
	}

	if *dbUrl == "" {
		log.Fatal("no db-url")
	}

	var err error
	geoip, err = lsgeoip.NewGeoIP(*geoipPath)
	if err != nil {
		panic(err)
	}
	defer geoip.Close()

	pg, err := sqlx.Open("postgres", *dbUrl)
	if err != nil {
		panic(err)
	}
	pg.SetMaxOpenConns(1)

	var data []Row

	rows, err := pg.Queryx(
		`SELECT listen_id, mount, client_id, host(ip_address) AS ip_addr, user_agent
		FROM listens.listen
		WHERE listen_id >= $1
		AND listen_id <= $2
		AND ip_address <> '127.0.0.1'::inet
		AND (geoip_country IS NULL OR geoip_location IS NULL)`,
		lt,
		gt,
	)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var row Row
		err = rows.Scan(
			&row.ListenID,
			&row.Mount,
			&row.ClientID,
			&row.IPAddr,
			&row.UserAgent,
		)
		if err != nil {
			panic(err)
		}
		data = append(data, row)
	}

	fmt.Printf("Processing %d records\n", len(data))

	tasks = make(chan Row, len(data))

	for i := 0; i < *numWorkers; i++ {
		go work()
	}

	fmt.Printf("Started %d workers\n", *numWorkers)

	for _, row := range data {
		tasks <- row
	}
	close(tasks)
	//fmt.Println("closed tasks")

	remainingWorkers := *numWorkers

work:
	for {
		select {
		case result := <-results:
			fmt.Printf("Result! %#v\n", result)
			_, err := pg.Exec(
				`UPDATE listens.listen
				SET geoip_country = $2, geoip_location = $3
				WHERE listen_id = $1`,
				result.ListenID,
				result.GeoIPCountry,
				result.GeoIPLocation,
			)
			if err != nil {
				panic(err)
			}
		case <-done:
			remainingWorkers--
			//fmt.Printf("worker done, %d remain\n", remainingWorkers)
			if remainingWorkers == 0 {
				break work
			}
		}
	}
}
