package svc

import "testing"

func TestEndpointToHost_AddressRecordTypes(t *testing.T) {
	s := &pfsenseService{}

	cases := []struct {
		name       string
		recordType string
		targets    []string
		wantIP     string
		wantErr    bool
	}{
		{name: "A record", recordType: "A", targets: []string{"10.10.0.130"}, wantIP: "10.10.0.130"},
		{name: "AAAA record", recordType: "AAAA", targets: []string{"2a00:6020:503d:b902::100:1000"}, wantIP: "2a00:6020:503d:b902::100:1000"},
		{name: "TXT record uses placeholder IP", recordType: "TXT", targets: []string{"some text"}, wantIP: "127.0.0.1"},
		{name: "unsupported record type rejected", recordType: "CNAME", targets: []string{"example.com"}, wantErr: true},
		{name: "A record with multiple targets rejected", recordType: "A", targets: []string{"10.0.0.1", "10.0.0.2"}, wantErr: true},
		{name: "AAAA record with multiple targets rejected", recordType: "AAAA", targets: []string{"::1", "::2"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := s.endpointToHost(UnboundEndpoint{
				DNSName:    "test.example.com",
				RecordType: tc.recordType,
				Targets:    tc.targets,
			})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none (host=%+v)", h)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if h.Ip != tc.wantIP {
				t.Errorf("Ip = %q, want %q", h.Ip, tc.wantIP)
			}
		})
	}
}

func TestHostToEndpoint_InfersRecordTypeFromAddressFamily(t *testing.T) {
	s := &pfsenseService{}

	cases := []struct {
		name           string
		ip             string
		wantRecordType string
	}{
		{name: "IPv4 address infers A", ip: "10.10.0.130", wantRecordType: "A"},
		{name: "IPv6 address infers AAAA", ip: "2a00:6020:503d:b902::100:1000", wantRecordType: "AAAA"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			endpoint, err := s.hostToEndpoint(host{Host: "test", Domain: "example.com", Ip: tc.ip})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if endpoint.RecordType != tc.wantRecordType {
				t.Errorf("RecordType = %q, want %q", endpoint.RecordType, tc.wantRecordType)
			}
		})
	}
}

func TestEndpointToHostToEndpoint_AAAARoundtrip(t *testing.T) {
	s := &pfsenseService{}

	original := UnboundEndpoint{
		DNSName:    "mqtt.crystalnet.org",
		RecordType: "AAAA",
		Targets:    []string{"2a00:6020:503d:b902::100:1000"},
	}

	h, err := s.endpointToHost(original)
	if err != nil {
		t.Fatalf("endpointToHost failed: %v", err)
	}

	roundtripped, err := s.hostToEndpoint(h)
	if err != nil {
		t.Fatalf("hostToEndpoint failed: %v", err)
	}

	if roundtripped.RecordType != "AAAA" {
		t.Errorf("RecordType = %q, want AAAA", roundtripped.RecordType)
	}
	if len(roundtripped.Targets) != 1 || roundtripped.Targets[0] != original.Targets[0] {
		t.Errorf("Targets = %+v, want %+v", roundtripped.Targets, original.Targets)
	}
}
