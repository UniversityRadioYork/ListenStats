package geoip

type Request struct {
	ClientID string
	IPAddr   string
}

type QueuedResult struct {
	Result
	ClientID string
}

type QueuedGeoIP struct {
	GeoIP
	requests chan Request
	Results  chan QueuedResult
	Errors   chan error
}

func NewQueuedGeoIP(dbPath string) (*QueuedGeoIP, error) {
	result := &QueuedGeoIP{
		requests: make(chan Request, 32),
		Results:  make(chan QueuedResult, 32),
		Errors:   make(chan error, 32),
	}
	geoip, err := NewGeoIP(dbPath)
	if err != nil {
		return nil, err
	}
	result.GeoIP = *geoip
	return result, nil
}

func (q *QueuedGeoIP) Request(clientId, ipAddr string) {
	q.requests <- Request{
		ClientID: clientId,
		IPAddr:   ipAddr,
	}
}

func (q *QueuedGeoIP) Close() {
	close(q.requests)
	close(q.Results)
	close(q.Errors)
}

func (q *QueuedGeoIP) Work() {
	for task := range q.requests {
		result, err := q.GeoIP.Process(task.IPAddr)
		if err != nil {
			q.Errors <- err
		} else {
			q.Results <- QueuedResult{
				Result:   *result,
				ClientID: task.ClientID,
			}
		}
	}
}
