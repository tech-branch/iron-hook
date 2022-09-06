package ironhook

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gofrs/uuid"
)

// Benchmark_EndpointDbToWeb100-8  412392  2961 ns/op  4096 B/op  1 allocs/op
func Benchmark_EndpointDbToWeb100(b *testing.B) {

	// prepare data
	dbes := make([]WebhookEndpointDB, 100)
	for i := range dbes {
		dbes[i] = WebhookEndpointDB{
			UUID:   uuid.Must(uuid.NewV4()),
			URL:    "http://localhost:" + fmt.Sprint(i),
			Status: Unverified,
		}
	}

	// dont count the data preparation step in the benchmark
	b.ResetTimer()

	// run the benchmarking loop
	for n := 0; n < b.N; n++ {
		_ = endpointsDbToWeb(&dbes)
	}
}

func Test_dbEndpointToWebEndpoint(t *testing.T) {

	dbe := WebhookEndpointDB{
		UUID:   uuid.Must(uuid.NewV4()),
		URL:    "http://localhost:8081",
		Status: Unverified,
	}

	type args struct {
		dbe *WebhookEndpointDB
	}
	tests := []struct {
		name string
		args args
		want *WebhookEndpoint
	}{
		{
			name: "Simple case",
			args: args{
				dbe: &dbe,
			},
			want: &WebhookEndpoint{
				UUID:   dbe.UUID,
				URL:    "http://localhost:8081",
				Status: Unverified,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpointDbToWeb(tt.args.dbe); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("endpointDbToWeb() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_endpointsDbToWeb(t *testing.T) {

	uuid1 := uuid.Must(uuid.NewV4())
	uuid2 := uuid.Must(uuid.NewV4())

	dbes := []WebhookEndpointDB{
		{
			UUID:   uuid1,
			URL:    "http://localhost:8080",
			Status: Unverified,
		},
		{
			UUID:   uuid2,
			URL:    "http://localhost:8081",
			Status: Unverified,
		},
	}

	webes := []WebhookEndpoint{
		{
			UUID:   uuid1,
			URL:    "http://localhost:8080",
			Status: Unverified,
		},
		{
			UUID:   uuid2,
			URL:    "http://localhost:8081",
			Status: Unverified,
		},
	}

	type args struct {
		dbes *[]WebhookEndpointDB
	}
	tests := []struct {
		name string
		args args
		want *[]WebhookEndpoint
	}{
		{
			name: "Simple case",
			args: args{
				dbes: &dbes,
			},
			want: &webes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpointsDbToWeb(tt.args.dbes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("endpointsDbToWeb() = %v, want %v", got, tt.want)
			}
		})
	}
}
