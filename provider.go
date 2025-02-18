package caddydnslinkup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/libdns"
	"go.uber.org/zap"
)

// Interface guards
var (
	_ caddy.Module          = (*Provider)(nil)
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)

func init() {
	caddy.RegisterModule(Provider{})
}

type Provider struct {
	client    *http.Client       `json:"-"`
	Logger    *zap.SugaredLogger `json:"-"`
	ctx       context.Context
	WorkerUrl string `json:"worker_url,omitempty"`
	Token     string `json:"token,omitempty"`
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	body := map[string]interface{}{"zone": zone, "records": recs}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return []libdns.Record{}, err
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/linkup/certificate-dns", p.WorkerUrl), bytes.NewBuffer(jsonBody))
	if err != nil {
		return []libdns.Record{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	return sendLibDnsLinkupRequest(p.Logger, p.client, req, zone)
}

func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	body := map[string]interface{}{"zone": zone, "records": recs}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return []libdns.Record{}, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/linkup/certificate-dns", p.WorkerUrl), bytes.NewBuffer(jsonBody))
	if err != nil {
		return []libdns.Record{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	return sendLibDnsLinkupRequest(p.Logger, p.client, req, zone)
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	body := map[string]interface{}{"zone": zone, "records": recs}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return []libdns.Record{}, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/linkup/certificate-dns", p.WorkerUrl), bytes.NewBuffer(jsonBody))
	if err != nil {
		return []libdns.Record{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	return sendLibDnsLinkupRequest(p.Logger, p.client, req, zone)
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/linkup/certificate-dns", p.WorkerUrl), nil)
	if err != nil {
		return []libdns.Record{}, err
	}

	q := req.URL.Query()
	q.Add("zone", zone)
	req.URL.RawQuery = q.Encode()

	return sendLibDnsLinkupRequest(p.Logger, p.client, req, zone)
}

func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "dns.providers.linkup",
		New: func() caddy.Module {
			return new(Provider)
		},
	}
}

func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		key := d.Val()
		var value string
		if !d.Args(&value) {
			continue
		}

		switch key {
		case "worker_url":
			p.WorkerUrl = value
		case "token":
			p.Token = value
		}
	}
	return nil
}

func (p *Provider) Provision(ctx caddy.Context) error {
	p.Logger = ctx.Logger(p).Sugar()
	p.ctx = ctx.Context

	// This adds support to the documented Caddy way to get runtime environment variables.
	// Reference: https://caddyserver.com/docs/caddyfile/concepts#environment-variables
	//
	// So, with this, it should be able to do something like this:
	// ```
	// worker_url {env.LINKUP_WORKER_URL}
	// ```
	// which would replace `{env.LINKUP_WORKER_URL}` with the environemnt variable value
	// of LINKUP_WORKER_URL at runtime.
	p.WorkerUrl = strings.TrimRight(caddy.NewReplacer().ReplaceAll(p.WorkerUrl, ""), "/")
	p.Token = caddy.NewReplacer().ReplaceAll(p.Token, "")

	p.client = http.DefaultClient

	return nil
}

func sendLibDnsLinkupRequest(logger *zap.SugaredLogger, client *http.Client, req *http.Request, zone string) ([]libdns.Record, error) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		return []libdns.Record{}, err
	}

	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	logger.Infow("Sending request to worker.",
		"method", req.Method,
		"url", req.URL.String(),
		"body", string(reqBody),
	)

	resp, err := client.Do(req)
	if err != nil {
		return []libdns.Record{}, err
	}
	defer resp.Body.Close()

	// TODO: Handle not 2xx response

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []libdns.Record{}, err
	}

	logger.Infow("Got response from worker.",
		"body", string(body),
	)

	var records []libdns.Record
	err = json.Unmarshal(body, &records)
	if err != nil {
		return []libdns.Record{}, err
	}

	logger.Infow("Formatted records", "records", records)

	return records, nil
}
