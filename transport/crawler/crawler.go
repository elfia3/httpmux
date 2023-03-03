package crawler

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	TimeoutRequest = time.Second * 1
)

type Crawler struct {
	cli *http.Client
}

func NewCrawler(timeout time.Duration, totalConn int) Crawler {
	return Crawler{
		cli: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: totalConn,
			},
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Crawler) ParseURLS(ctx context.Context, jobQueue []string) (err error, response [][]string) {
	eg, ctx := errgroup.WithContext(ctx)
	results := make(chan []string, len(jobQueue))

	for _, job := range jobQueue {
		job := job

		eg.Go(func() error {
			ctxWithTimeout, cancel := context.WithTimeout(ctx, TimeoutRequest)
			defer cancel()

			req, _ := http.NewRequestWithContext(ctxWithTimeout, http.MethodGet, job, nil)
			if response, err := c.cli.Do(req); err == nil {
				response.Body.Close()
				results <- []string{job, response.Status}
				return nil
			} else {
				return err
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return err, nil
	}
	close(results)

	for result := range results {
		response = append(response, result)
	}

	return nil, response
}
