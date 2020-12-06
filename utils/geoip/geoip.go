package geoip

import (
	"github.com/oschwald/geoip2-golang"
	"net"
)

type Result struct {
	GeoIPCountry  string `sql:"geoip_country"`
	GeoIPLocation string `sql:"geoip_location"`
}

type GeoIP struct {
	geoipDb *geoip2.Reader
}

func (g *GeoIP) Process(ipaddr string) (*Result, error) {
	ip := net.ParseIP(ipaddr)
	country, err := g.geoipDb.Country(ip)
	if err != nil {
		return nil, err
	}
	loc, err := g.geoipDb.City(ip)
	if err != nil {
		return nil, err
	}
	result := Result{
		GeoIPCountry:  country.Country.IsoCode,
		GeoIPLocation: loc.City.Names["en"],
	}
	return &result, nil
}

func (g *GeoIP) Close() error {
	return g.geoipDb.Close()
}

func NewGeoIP(dbPath string) (*GeoIP, error) {
	var err error
	result := &GeoIP{}
	result.geoipDb, err = geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return result, nil
}
