package ticket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type serviceNowConfig struct {
	baseURL           string
	username          string
	password          string
	catalogItemSysID  string
	groupVariableName string
	userVariableName  string
}

var serviceNowTenant = "your-instance"
var serviceNowUsername = "your-username"
var serviceNowPassword = "your-password"
var groupMembershipCatalogItemSysID = "replace-with-group-membership-catalog-item-sys-id"
var groupMembershipGroupVariable = "group"
var groupMembershipUserVariable = "user"

var serviceNowSettings = serviceNowConfig{
	baseURL:           "https://" + serviceNowTenant + ".service-now.com",
	username:          serviceNowUsername,
	password:          serviceNowPassword,
	catalogItemSysID:  groupMembershipCatalogItemSysID,
	groupVariableName: groupMembershipGroupVariable,
	userVariableName:  groupMembershipUserVariable,
}

var serviceNowHTTPClient = &http.Client{Timeout: 30 * time.Second}

func createGroupMembershipRequest(ctx context.Context, client *http.Client, config serviceNowConfig, groupName, userName string) (string, error) {
	if strings.Contains(config.baseURL, "your-instance") || config.baseURL == "" ||
		config.username == "" || config.password == "" ||
		strings.Contains(config.catalogItemSysID, "replace-with") || config.catalogItemSysID == "" ||
		config.groupVariableName == "" || config.userVariableName == "" {
		return "", fmt.Errorf("configure the ServiceNow settings in group_membership.go before creating a request")
	}

	endpoint := fmt.Sprintf("%s/api/sn_sc/servicecatalog/items/%s/order_now", strings.TrimRight(config.baseURL, "/"), config.catalogItemSysID)
	payload := struct {
		Quantity  int               `json:"sysparm_quantity"`
		Variables map[string]string `json:"variables"`
	}{
		Quantity: 1,
		Variables: map[string]string{
			config.groupVariableName: groupName,
			config.userVariableName:  userName,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode ServiceNow request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build ServiceNow request: %w", err)
	}
	request.SetBasicAuth(config.username, config.password)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("send ServiceNow request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read ServiceNow response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("ServiceNow returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	requestID, err := extractServiceNowRequestID(responseBody)
	if err != nil {
		return "", err
	}
	return requestID, nil
}

func extractServiceNowRequestID(responseBody []byte) (string, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return "", fmt.Errorf("decode ServiceNow response: %w", err)
	}

	result := envelope["result"]
	if len(result) == 0 {
		result = responseBody
	}
	var resultFields map[string]json.RawMessage
	if err := json.Unmarshal(result, &resultFields); err != nil {
		return "", fmt.Errorf("decode ServiceNow result: %w", err)
	}

	for _, field := range []string{"request_number", "request_id", "requestId", "number", "sys_id", "sysId"} {
		if value, ok := resultFields[field]; ok {
			var requestID string
			if err := json.Unmarshal(value, &requestID); err == nil && requestID != "" {
				return requestID, nil
			}
		}
	}
	return "", fmt.Errorf("ServiceNow response did not contain a request ID")
}
