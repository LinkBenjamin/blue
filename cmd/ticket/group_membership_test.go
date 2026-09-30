package ticket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGroupMembershipRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", request.Method)
		}
		if request.URL.Path != "/api/sn_sc/servicecatalog/items/catalog-item-id/order_now" {
			t.Errorf("path = %q, want catalog order endpoint", request.URL.Path)
		}
		username, password, ok := request.BasicAuth()
		if !ok || username != "api-user" || password != "api-password" {
			t.Errorf("unexpected basic auth credentials")
		}

		var payload struct {
			Quantity  int               `json:"sysparm_quantity"`
			Variables map[string]string `json:"variables"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		if payload.Quantity != 1 || payload.Variables["requested_group"] != "Engineering" || payload.Variables["requested_user"] != "alex" {
			t.Errorf("unexpected request payload: %+v", payload)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"result":{"request_number":"REQ0012345"}}`))
	}))
	defer server.Close()

	config := serviceNowConfig{
		baseURL:           server.URL,
		username:          "api-user",
		password:          "api-password",
		catalogItemSysID:  "catalog-item-id",
		groupVariableName: "requested_group",
		userVariableName:  "requested_user",
	}
	requestID, err := createGroupMembershipRequest(context.Background(), server.Client(), config, "Engineering", "alex")
	if err != nil {
		t.Fatalf("createGroupMembershipRequest returned error: %v", err)
	}
	if requestID != "REQ0012345" {
		t.Fatalf("request ID = %q, want REQ0012345", requestID)
	}
}

func TestCreateGroupMembershipRequestRejectsServiceNowError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "invalid catalog item", http.StatusNotFound)
	}))
	defer server.Close()

	config := serviceNowConfig{
		baseURL:           server.URL,
		username:          "api-user",
		password:          "api-password",
		catalogItemSysID:  "catalog-item-id",
		groupVariableName: "group",
		userVariableName:  "user",
	}
	if _, err := createGroupMembershipRequest(context.Background(), server.Client(), config, "Engineering", "alex"); err == nil {
		t.Fatal("createGroupMembershipRequest succeeded for a ServiceNow error response")
	}
}
