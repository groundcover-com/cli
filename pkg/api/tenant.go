package api

import (
	"encoding/json"
)

const (
	TenantUUIDHeader     = "X-Tenant-UUID"
	TENANT_LIST_ENDPOINT = "rbac/member/tenants"
)

type TenantListResponse struct {
	Tenants []*TenantInfo `json:"tenants"`
}

type TenantInfo struct {
	UUID       string `json:"UUID"`
	OrgName    string `json:"OrgName"`
	TenantName string `json:"TenantName"`
}

func (client *Client) TenantList() ([]*TenantInfo, error) {
	var err error

	var body []byte
	if body, err = client.get(TENANT_LIST_ENDPOINT); err != nil {
		return nil, err
	}

	var tenantListResponse TenantListResponse
	if err = json.Unmarshal(body, &tenantListResponse); err != nil {
		return nil, err
	}

	return tenantListResponse.Tenants, nil
}
