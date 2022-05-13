package main

import (
	"os"
	"fmt"
	"time"
	"crypto/tls"
	"github.com/arangodb/go-driver/http"
	"github.com/arangodb/go-driver"
	"context"
	"sync"
	"strings"
	http3 "net/http"
)

func main() {
	ip := os.Args[1]
	id := os.Args[2]

	endpointMap := map[string]string{
		"eu": "34.116.157.50",
		"us": "34.136.136.255",
	}

	endpoints := fmt.Sprintf("https://%s:8529", endpointMap[ip])

	token := "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmFuZ29kYiIsInNlcnZlcl9pZCI6ImZvbyJ9.JL0AS55Y5_Iyexn8BwcBNbRuZoazVEHs4sZc-SN4H-I"

	ticker := time.NewTicker(2 * time.Second)

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{endpoints},
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DisableKeepAlives: true,
		},
		ContentType: driver.ContentTypeJSON,
	})
	if err != nil {
		panic(err)
	}

	if nConn, err := conn.SetAuthentication(driver.RawAuthentication(token)); err != nil {
		panic(err)
	} else {
		conn = nConn
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		panic(err)
	}

	names := []string{
		"Version check",
	}
	size := len(names)

	requests := make([]time.Duration, size)
	resp := make([]string, size)

	for range ticker.C {
		var wg sync.WaitGroup

		wg.Add(size)

		go func() {
			defer wg.Done()

			s := time.Now()
			if _, err := client.Version(context.Background()); err != nil {
				println(err.Error())
			} else {
				requests[0] = time.Since(s)
			}
		}()

		wg.Wait()

		for n := range names {
			resp[n] = fmt.Sprintf("%s: %s", names[n], requests[n].String())
		}

		println(fmt.Sprintf("%s: %s", id, strings.Join(resp, ", ")))
	}
}
