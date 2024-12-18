package cloudflare

import (
	"context"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// customHostnameRow is a wrapper struct that includes the zone ID and flattens the CustomHostname fields
type customHostnameRow struct {
	ID                 string                        `json:"id"`
	ZoneID             string                        `json:"zone_id"`
	Hostname           string                        `json:"hostname"`
	CustomOriginServer string                        `json:"custom_origin_server"`
	Status             string                        `json:"status"`
	CreatedAt          *time.Time                    `json:"created_at"`
	SSLConfig          *cloudflare.CustomHostnameSSL `json:"-"` // Internal use only for field mapping
}

func tableCloudflareCustomHostname(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "cloudflare_custom_hostname",
		Description: "Custom Hostnames allow you to use your own domain name with SSL/TLS protection for your zone.",
		List: &plugin.ListConfig{
			ParentHydrate: listZones,
			Hydrate:       listCustomHostnames,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"zone_id", "id"}),
			Hydrate:    getCustomHostname,
		},
		Columns: commonColumns([]*plugin.Column{
			// Top columns
			{Name: "id", Type: proto.ColumnType_STRING, Description: "Custom hostname identifier."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "The custom hostname that will point to your zone."},
			{Name: "zone_id", Type: proto.ColumnType_STRING, Description: "Zone identifier."},

			// Other columns
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "When the custom hostname was created."},
			{Name: "custom_origin_server", Type: proto.ColumnType_STRING, Description: "Custom origin server address for the hostname."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Status of the custom hostname."},

			// SSL fields
			{Name: "certificate_authority", Type: proto.ColumnType_STRING, Transform: transform.FromField("SSLConfig.CertificateAuthority"), Description: "Certificate authority used for the SSL certificate."},
			{Name: "ssl_id", Type: proto.ColumnType_STRING, Transform: transform.FromField("SSLConfig.ID"), Description: "Identifier of the SSL certificate."},
			{Name: "ssl_method", Type: proto.ColumnType_STRING, Transform: transform.FromField("SSLConfig.Method"), Description: "Method used for SSL verification."},
			{Name: "ssl_settings", Type: proto.ColumnType_JSON, Transform: transform.FromField("SSLConfig.Settings"), Description: "Additional SSL settings for the hostname."},
			{Name: "ssl_status", Type: proto.ColumnType_STRING, Transform: transform.FromField("SSLConfig.Status"), Description: "Status of the SSL certificate."},
			{Name: "ssl_type", Type: proto.ColumnType_STRING, Transform: transform.FromField("SSLConfig.Type"), Description: "Type of SSL certificate."},
			{Name: "ssl_wildcard", Type: proto.ColumnType_BOOL, Transform: transform.FromField("SSLConfig.Wildcard"), Description: "Whether the SSL certificate is a wildcard certificate."},
		}),
	}
}

func listCustomHostnames(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	zone := h.Item.(cloudflare.Zone)
	logger := plugin.Logger(ctx)

	conn, err := connect(ctx, d)
	if err != nil {
		logger.Error("listCustomHostnames", "connection_error", err)
		return nil, err
	}

	// Initialize page number
	page := 1

	for {
		// List custom hostnames for the zone with pagination
		customHostnames, info, err := conn.CustomHostnames(ctx, zone.ID, page, cloudflare.CustomHostname{})
		if err != nil {
			logger.Error("listCustomHostnames", "api_error", err)
			return nil, err
		}

		for _, hostname := range customHostnames {
			row := customHostnameRow{
				ID:                 hostname.ID,
				ZoneID:             zone.ID,
				Hostname:           hostname.Hostname,
				CustomOriginServer: hostname.CustomOriginServer,
				Status:             string(hostname.Status),
				CreatedAt:          hostname.CreatedAt,
				SSLConfig:          hostname.SSL,
			}
			d.StreamListItem(ctx, row)
		}

		// Check if we've reached the last page
		if page >= info.TotalPages {
			break
		}

		// Move to next page
		page++
	}

	return nil, nil
}

func getCustomHostname(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	conn, err := connect(ctx, d)
	if err != nil {
		logger.Error("getCustomHostname", "connection_error", err)
		return nil, err
	}

	quals := d.EqualsQuals
	zoneID := quals["zone_id"].GetStringValue()
	hostnameID := quals["id"].GetStringValue()

	hostname, err := conn.CustomHostname(ctx, zoneID, hostnameID)
	if err != nil {
		logger.Error("getCustomHostname", "api_error", err)
		return nil, err
	}

	row := customHostnameRow{
		ID:                 hostname.ID,
		ZoneID:             zoneID,
		Hostname:           hostname.Hostname,
		CustomOriginServer: hostname.CustomOriginServer,
		Status:             string(hostname.Status),
		CreatedAt:          hostname.CreatedAt,
		SSLConfig:          hostname.SSL,
	}

	return row, nil
}
