package caddydnslinkup

import (
	"context"

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
	Logger    *zap.SugaredLogger `json:"-"`
	ctx       context.Context
	WorkerUrl string `json:"worker_url,omitempty"`
	Token     string `json:"token,omitempty"`
}

// TODO: Do we want to listen to these zone? I think so, but confirm

func (s *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	// DELETE /linkup/certificate-dns
	panic("unimplemented")
}

func (s *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	// POST /linkup/certificate-dns
	panic("unimplemented")
}

func (s *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	// PUT /linkup/certificate-dns
	panic("unimplemented")
}

func (s *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	// GET /linkup/certificate-dns
	panic("unimplemented")
}

func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "dns.providers.linkup",
		New: func() caddy.Module {
			return new(Provider)
		},
	}
}

func (s *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		key := d.Val()
		var value string
		if !d.Args(&value) {
			continue
		}

		switch key {
		case "worker_url":
			s.WorkerUrl = value
		case "token":
			s.Token = value
		}
	}
	return nil
}

func (s *Provider) Provision(ctx caddy.Context) error {
	// s.Logger = ctx.Logger(s).Sugar()
	// s.ctx = ctx.Context

	// This adds support to the documented Caddy way to get runtime environment variables.
	// Reference: https://caddyserver.com/docs/caddyfile/concepts#environment-variables
	//
	// So, with this, it should be able to do something like this:
	// ```
	// worker_url {env.LINKUP_WORKER_URL}
	// ```
	// which would replace `{env.LINKUP_WORKER_URL}` with the environemnt variable value
	// of LINKUP_WORKER_URL at runtime.
	s.WorkerUrl = caddy.NewReplacer().ReplaceAll(s.WorkerUrl, "")
	s.Token = caddy.NewReplacer().ReplaceAll(s.Token, "")

	return nil
}
