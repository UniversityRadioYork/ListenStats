package main

import (
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
	"runtime"
)

type Row struct {
	ListenID  uint64 `sql:"listen_id"`
	Mount     string `sql:"mount"`
	ClientID  string `sql:"client_id"`
	IPAddr    string `sql:"ip_addr"`
	UserAgent string `sql:"user_agent"`
}

type Result struct {
	ListenID      uint64 `sql:"listen_id"`
	GeoIPCountry  string `sql:"geoip_country"`
	GeoIPLocation string `sql:"geoip_location"`
}

var geoipDb *geoip2.Reader
var tasks chan Row
var results = make(chan Result)
var done = make(chan struct{})

func work() {
	//fmt.Println("worker ready")
	for task := range tasks {
		//fmt.Printf("task %v\n", task)
		ip := net.ParseIP(task.IPAddr)
		country, err := geoipDb.Country(ip)
		if err != nil {
			panic(err)
		}
		loc, err := geoipDb.City(ip)
		if err != nil {
			panic(err)
		}
		result := Result{
			ListenID:      task.ListenID,
			GeoIPCountry:  country.Country.IsoCode,
			GeoIPLocation: loc.City.Names["en"],
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
	geoipDb, err = geoip2.Open(*geoipPath)
	if err != nil {
		panic(err)
	}
	defer geoipDb.Close()

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
