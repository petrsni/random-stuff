package tsgw

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type AppParams struct {
	InfluxUrl   string
	InfluxToken string

	AppPort int

	AppUser string
	AppPass string
}

type InfluxGw struct {
	server *http.Server
	ctx    context.Context
	influx influxdb2.Client
	AppParams
}

func NewInfluxGw(params AppParams) (*InfluxGw, error) {
	client := influxdb2.NewClient(params.InfluxUrl, params.InfluxToken)
	return &InfluxGw{
		influx:    client,
		AppParams: params,
	}, nil
}

func (c *InfluxGw) Run() error {
	slog.Info("starting", slog.String("influx url", c.InfluxUrl), slog.Int("app port", c.AppPort))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /influx/{org}/{bucket}", c.handlePost)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.AppPort),
		Handler: mux,
	}
	go server.ListenAndServe()

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	<-term
	slog.Info("shutting down...")
	shutctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	server.Shutdown(shutctx)
	return nil

}

func (c *InfluxGw) handlePost(w http.ResponseWriter, req *http.Request) {
	if !c.isAuthenticated(req) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	org := req.PathValue("org")
	bucket := req.PathValue("bucket")

	payload, err := readPayload(req.Body)
	if err != nil {
		slog.Error("invalid request body", slog.String("err", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slog.Info("received data", slog.String("org", org), slog.String("bucket", bucket))
	err = writemsg(c.influx, org, bucket, payload)
	if err != nil {
		slog.Error("failed to write", slog.String("err", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (c *InfluxGw) isAuthenticated(req *http.Request) bool {
	user, pass, ok := req.BasicAuth()
	if !ok {
		return false
	}
	return user == c.AppUser && pass == c.AppPass
}

func readPayload(body io.ReadCloser) (InfluxMsg, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return InfluxMsg{}, err
	}
	slog.Debug("received payload", slog.String("payload", string(b)))
	ret := InfluxMsg{}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return InfluxMsg{}, err
	}
	return ret, nil
}

func writemsg(client influxdb2.Client, org string, bucket string, msg InfluxMsg) error {
	writeapi := client.WriteAPIBlocking(org, bucket)
	var timestamp time.Time
	if msg.Ts == 0 {
		timestamp = time.Now()
	} else {
		timestamp = time.Unix(msg.Ts, 0)
	}
	p := influxdb2.NewPoint(msg.Measurement, msg.Tags, msg.Fields, timestamp)
	err := writeapi.WritePoint(context.Background(), p) // TBD use deadline context
	if err != nil {
		return err
	}
	slog.Debug("wrote point", slog.String("measurement", msg.Measurement))
	return nil
}

type InfluxMsg struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	Ts          int64                  `json:"ts"`
}
type UnixTimestamp struct {
	Seconds     int64
	Nanoseconds int64
}
