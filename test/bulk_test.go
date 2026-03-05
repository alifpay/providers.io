package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
)

func Test_BulkSend(t *testing.T) {
	const (
		url   = "http://localhost:8080/payments"
		total = 50
	)

	var wg sync.WaitGroup
	wg.Add(total)

	for i := range total {
		go func(i int) {
			defer wg.Done()

			body, _ := json.Marshal(map[string]any{
				"provider_id": (i % 3) + 1,
				"agent_id":    "agent1",
				"ref_id":      fmt.Sprintf("ref-6321%d", i),
				"amount":      90.00,
			})

			resp, err := http.Post(url, "application/json", bytes.NewReader(body))
			if err != nil {
				t.Errorf("request %d: %v", i, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("request %d: status %d body: %s", i, resp.StatusCode, body)
			}
		}(i)
	}

	wg.Wait()
}
