package spin

import (
	"strings"
	"testing"
)

func TestParseIPv4CIDR(t *testing.T) {
	prefix, err := parseIPv4CIDR("10.40.1.17/24", "subnet CIDR")
	if err != nil {
		t.Fatalf("parseIPv4CIDR returned error: %v", err)
	}
	if got, want := prefix.String(), "10.40.1.0/24"; got != want {
		t.Fatalf("parseIPv4CIDR returned %q, want %q", got, want)
	}

	if _, err := parseIPv4CIDR("2001:db8::/64", "subnet CIDR"); err == nil {
		t.Fatal("parseIPv4CIDR accepted an IPv6 CIDR")
	}
}

func TestPrefixesOverlap(t *testing.T) {
	parent, err := parseIPv4CIDR("10.40.0.0/16", "VPC CIDR")
	if err != nil {
		t.Fatal(err)
	}
	child, err := parseIPv4CIDR("10.40.1.0/24", "subnet CIDR")
	if err != nil {
		t.Fatal(err)
	}
	disjoint, err := parseIPv4CIDR("10.40.2.0/24", "subnet CIDR")
	if err != nil {
		t.Fatal(err)
	}

	if !prefixesOverlap(parent, child) {
		t.Fatal("expected nested CIDRs to overlap")
	}
	if prefixesOverlap(child, disjoint) {
		t.Fatal("expected disjoint subnet CIDRs not to overlap")
	}
}

func TestBuildVPCTemplateScalesSubnetResources(t *testing.T) {
	template := buildVPCTemplate(2)
	for _, expected := range []string{
		"Subnet01Cidr:",
		"Subnet02Cidr:",
		"Subnet02RouteTableAssociation:",
		"MapPublicIpOnLaunch: false",
		"TrafficType: REJECT",
		"CloseDefaultSecurityGroup:",
		"Ec2ApiEndpoint:",
		"EcrApiEndpoint:",
		"EcrDockerEndpoint:",
		"SystemsManagerEndpoint:",
		"S3GatewayEndpoint:",
		"Condition: CreateEc2ApiEndpoint",
		"Condition: CreateEcrApiEndpoint",
		"Condition: CreateEcrDockerEndpoint",
		"Condition: CreateSsmEndpoint",
		"Condition: CreateS3GatewayEndpoint",
		"Default: 'false'",
		"Subnet02Id:",
	} {
		if !strings.Contains(template, expected) {
			t.Errorf("generated template does not contain %q", expected)
		}
	}
	if strings.Contains(template, "Subnet03Cidr:") {
		t.Fatal("generated template contains a subnet parameter beyond the requested count")
	}
}

func TestBuildVPCParametersDefaultsEndpointsToFalse(t *testing.T) {
	parameters := buildVPCParameters("10.40.0.0/16", nil, vpcEndpointOptions{
		addEcrApiEndpoint: true,
	})
	values := make(map[string]string, len(parameters))
	for _, parameter := range parameters {
		values[parameter.ParameterKey] = parameter.ParameterValue
	}
	for key, want := range map[string]string{
		"AddEc2ApiEndpoint":    "false",
		"AddEcrApiEndpoint":    "true",
		"AddEcrDockerEndpoint": "false",
		"AddSsmEndpoint":       "false",
		"AddS3GatewayEndpoint": "false",
	} {
		if got := values[key]; got != want {
			t.Errorf("parameter %s = %q, want %q", key, got, want)
		}
	}
}
