package reporters

import (
	"log"
	"time"
)

type LogReporter struct {
}

func (l *LogReporter) ReportListenStart(clientId string, info *ListenerInfo) error {
	log.Printf("LISTENER [%s] from %s to %s, %v", clientId, info.IP, info.ServerURL, info.QueryParams)
	return nil
}

func (l *LogReporter) ReportListenEnd(clientId string, time time.Duration) error {
	log.Printf("LISTENER [%s] over, took %v", clientId, time)
	return nil
}
