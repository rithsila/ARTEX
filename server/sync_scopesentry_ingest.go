package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Autumn-27/artex/db"
)

// ssIngest preserves the UI importer mappings, counts, and partial-error behavior.
// Cancellation is checked before every write and after the final synchronous call.
func ssIngest(ctx context.Context, as ssAssetStore, assetType string, raw json.RawMessage, synced map[string]int) (err error) {
	defer func() {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	}()
	switch assetType {
	case "subdomain":
		var it struct {
			Host  string   `json:"host"`
			Type  string   `json:"type"`
			Value []string `json:"value"`
			IP    []string `json:"ip"`
		}
		if err := json.Unmarshal(raw, &it); err != nil {
			return fmt.Errorf("failed to parse subdomain: %w", err)
		}
		if it.Host == "" {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := as.UpsertSubdomain(db.UpsertSubdomainReq{Domain: it.Host, RecordType: it.Type, RecordValue: it.Value}); err != nil {
			return fmt.Errorf("subdomain %s: %w", it.Host, err)
		}
		synced["subdomain"]++
		for _, ip := range it.IP {
			if ip != "" {
				if err := ctx.Err(); err != nil {
					return err
				}
				if _, err := as.UpsertIP(db.UpsertIPReq{IP: ip, BoundDomains: []string{it.Host}}); err == nil {
					synced["ip"]++
				} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
			}
		}
	case "app":
		var it struct {
			Name        string `json:"name"`
			Category    string `json:"category"`
			Description string `json:"description"`
			ICP         string `json:"icp"`
		}
		if err := json.Unmarshal(raw, &it); err != nil {
			return fmt.Errorf("failed to parse app: %w", err)
		}
		if it.Name == "" {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := as.UpsertApp(db.UpsertAppReq{Name: it.Name, Category: it.Category, Description: it.Description, ICP: it.ICP}); err != nil {
			return fmt.Errorf("app %s: %w", it.Name, err)
		}
		synced["app"]++
	case "service":
		var it struct {
			Domain   string   `json:"domain"`
			IP       string   `json:"ip"`
			Port     string   `json:"port"`
			Service  string   `json:"service"`
			URL      string   `json:"url"`
			Title    string   `json:"title"`
			Status   *int     `json:"status"`
			Products []string `json:"products"`
			Icon     string   `json:"icon"`
		}
		if err := json.Unmarshal(raw, &it); err != nil {
			return fmt.Errorf("failed to parse service: %w", err)
		}
		if it.Service == "http" || it.URL != "" {
			if it.URL == "" {
				return nil
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, err := as.UpsertHTTPService(db.UpsertHTTPServiceReq{URL: it.URL, Technologies: it.Products, StatusCode: it.Status, PageTitle: it.Title, FaviconMMH3: it.Icon, IP: it.IP}); err != nil {
				return fmt.Errorf("service %s: %w", it.URL, err)
			}
		} else {
			port, _ := strconv.Atoi(it.Port)
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, err := as.UpsertOtherService(db.UpsertOtherServiceReq{Domain: it.Domain, IP: it.IP, Port: port, ServiceName: it.Service}); err != nil {
				return fmt.Errorf("service %s:%s: %w", it.IP, it.Port, err)
			}
		}
		synced["service"]++
	}
	return nil
}
