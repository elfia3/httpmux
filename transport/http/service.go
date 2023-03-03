package transport

import (
	"context"
	"encoding/json"
	"httpmux/transport/crawler"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	MaxConnections = 100
	JsonDataLimit  = 20

	ErrNotValidJson  = "Error with JSON in request, not valid data"
	ErrJsonDataLimit = "Json data limit exceeded"
	ErrConnRateLimit = "Too many connections, try again later"
)

type Service struct {
	server    http.Server
	totalConn int // atomic
}

type UrlRequest struct {
	Urls []string `json:"urls"`
}
type UrlAnswer struct {
	Urls [][]string `json:"urls"`
}

func CreateService(size int, Addr string) Service {
	return Service{
		server: http.Server{Addr: Addr},
	}
}

func (s *Service) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", s.parseURLs)
	s.server.Handler = mux
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	return nil
}

func (s *Service) ShutdownGracefully() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *Service) parseURLs(w http.ResponseWriter, r *http.Request) {
	s.totalConn++
	defer func() {
		s.totalConn--
	}()

	if s.totalConn > MaxConnections {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrConnRateLimit))
		return
	}

	urls := UrlRequest{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := json.Unmarshal(body, &urls); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrNotValidJson))
		return
	}

	if len(urls.Urls) > JsonDataLimit {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		w.Write([]byte(ErrJsonDataLimit))
		return
	}

	crawler := crawler.NewCrawler(time.Second, len(urls.Urls))

	errs, resp := crawler.ParseURLS(r.Context(), urls.Urls)
	if errs != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errs.Error()))
		return
	}
	ans := UrlAnswer{resp}
	res, err := json.Marshal(ans)
	w.Write(res)
}
