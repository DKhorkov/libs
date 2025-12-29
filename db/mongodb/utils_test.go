package mongodb

import "testing"

func TestBuildDsn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "normal host and port",
			config: Config{
				Host: "localhost",
				Port: 27017,
			},
			expected: "mongodb://localhost:27017",
		},
		{
			name: "localhost with default port",
			config: Config{
				Host: "127.0.0.1",
				Port: 27017,
			},
			expected: "mongodb://127.0.0.1:27017",
		},
		{
			name: "remote host with custom port",
			config: Config{
				Host: "mongo.example.com",
				Port: 27018,
			},
			expected: "mongodb://mongo.example.com:27018",
		},
		{
			name: "IPv6 address with port",
			config: Config{
				Host: "::1",
				Port: 27017,
			},
			expected: "mongodb://::1:27017",
		},
		{
			name: "hostname with zero port",
			config: Config{
				Host: "localhost",
				Port: 0,
			},
			expected: "mongodb://localhost:0",
		},
		{
			name: "empty host with port",
			config: Config{
				Host: "",
				Port: 27017,
			},
			expected: "mongodb://:27017",
		},
		{
			name: "domain host with high port number",
			config: Config{
				Host: "cluster.mongodb.net",
				Port: 65535,
			},
			expected: "mongodb://cluster.mongodb.net:65535",
		},
		{
			name: "host with negative port (edge case)",
			config: Config{
				Host: "localhost",
				Port: -1,
			},
			expected: "mongodb://localhost:-1",
		},
		{
			name: "host with maximum int port",
			config: Config{
				Host: "host",
				Port: 2147483647,
			},
			expected: "mongodb://host:2147483647",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := BuildDsn(tt.config)
			if result != tt.expected {
				t.Errorf("BuildDsn() = %v, want %v", result, tt.expected)
			}
		})
	}
}
