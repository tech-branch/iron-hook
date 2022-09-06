package ironhook

import (
	"bytes"
	"io"
	"testing"

	"github.com/gofrs/uuid"
)

func Test_prepareEndpointForVerification(t *testing.T) {

	type args struct {
		endpoint string
		uuid     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Correct1",
			args: args{
				endpoint: "https://webhooks.example.com/",
				uuid:     "3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			},
			want:    "https://webhooks.example.com/verification?id=3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			wantErr: false,
		},
		{
			name: "Correct2",
			args: args{
				endpoint: "http://example.com/",
				uuid:     "3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			},
			want:    "http://example.com/verification?id=3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			wantErr: false,
		},
		{
			name: "Incorrect1",
			args: args{
				endpoint: "example.com",
				uuid:     "3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Incorrect2",
			args: args{
				endpoint: "udp://example.com",
				uuid:     "3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Incorrect3",
			args: args{
				endpoint: "lol://example.com",
				uuid:     "3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Incorrect4",
			args: args{
				endpoint: "httpexample.com",
				uuid:     "3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Correct3",
			args: args{
				endpoint: "http://example.com/webhooks",
				uuid:     "3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			},
			want:    "http://example.com/webhooks/verification?id=3f7dc3d8-ac53-4819-a0db-8e643da31fa0",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareEndpointForVerification(tt.args.endpoint, tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepareEndpointForVerification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("prepareEndpointForVerification() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_verifyResponseBodyForVerification(t *testing.T) {

	// test datasets
	//
	// correct1
	data1 := []byte("abcd")
	reader1 := bytes.NewReader(data1)
	// correct2
	data2 := []byte("6551e000-947a-40a4-948d-18d01e3660d4")
	reader2 := bytes.NewReader(data2)
	// incorrect1
	data3 := []byte("")
	reader3 := bytes.NewReader(data3)
	// incorrect2
	data4 := []byte("6551e000-947a-40a4-948d-18d01e3660d5")
	reader4 := bytes.NewReader(data4)

	type args struct {
		iobody io.Reader
		uuid   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Correct1",
			args: args{
				iobody: reader1,
				uuid:   "abcd",
			},
			wantErr: false,
		},
		{
			name: "Correct2",
			args: args{
				iobody: reader2,
				uuid:   "6551e000-947a-40a4-948d-18d01e3660d4",
			},
			wantErr: false,
		},
		{
			name: "Incorrect1",
			args: args{
				iobody: reader3,
				uuid:   "abcd",
			},
			wantErr: true,
		},
		{
			name: "Incorrect2",
			args: args{
				iobody: reader4,
				uuid:   "6551e000-947a-40a4-948d-18d01e3660d4",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := verifyResponseBodyForVerification(tt.args.iobody, tt.args.uuid); (err != nil) != tt.wantErr {
				t.Errorf("verifyResponseBodyForVerification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Flow 1
// Create -> Verify
func Test_CreateVerifyFlow(t *testing.T) {

	server := mockHttpWebhooksServerForTests()
	mock_http_client := server.Client()
	defer server.Close()

	svc, err := NewWebhookService(mock_http_client)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := svc.Create(
		WebhookEndpoint{
			URL: server.URL,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Verify(endpoint)
	if err != nil {
		t.Fatal(err)
	}
}

// Flow 2
// Create -> Update -> Verify
func Test_CreateUpdateVerifyFlow(t *testing.T) {

	server := mockHttpWebhooksServerForTests()
	mock_http_client := server.Client()
	defer server.Close()

	svc, err := NewWebhookService(mock_http_client)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := svc.Create(
		WebhookEndpoint{
			URL: "http://localhost",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	updated_endpoint, err := svc.UpdateURL(
		WebhookEndpoint{
			UUID: endpoint.UUID,
			URL:  server.URL,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Verify(updated_endpoint)
	if err != nil {
		t.Fatal(err)
	}
}

// Flow 3
// Create -> List
func Test_CreateListFlow(t *testing.T) {

	svc, err := NewWebhookService(nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Create(
		WebhookEndpoint{
			URL: "http://example.com:8080",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Create(
		WebhookEndpoint{
			URL: "http://example.com:8081",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err := svc.ListEndpoints()
	if err != nil {
		t.Fatal(err)
	}
	if len(*endpoints) != 2 {
		t.Fatal("Expected two endpoints, found ", len(*endpoints))
	}

}

// Flow 4
// Create -> Verify -> Notify -> ListNotifs
func Test_GetNotifyFlow(t *testing.T) {

	server := mockHttpWebhooksServerForTests()
	mock_http_client := server.Client()
	defer server.Close()

	svc, err := NewWebhookService(mock_http_client)
	if err != nil {
		t.Fatal(err)
	}

	// mock events
	customer_registered_event := uuid.Must(uuid.NewV4())
	another_customer_registered := uuid.Must(uuid.NewV4())

	endpoint, err := svc.Create(
		WebhookEndpoint{
			URL: server.URL,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	verified_endpoint, err := svc.Verify(endpoint)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.Notify(
		verified_endpoint,
		WebhookNotification{
			EventUUID: customer_registered_event,
			Topic:     "",
			Body:      "",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.Notify(
		verified_endpoint,
		WebhookNotification{
			EventUUID: another_customer_registered,
			Topic:     "",
			Body:      "",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	noti, err := svc.LastNotificationSent(verified_endpoint)
	if err != nil {
		t.Fatal(err)
	}

	if noti.EventUUID != another_customer_registered {
		t.Fatal("Expected UUID", another_customer_registered.String(), " found ", noti.EventUUID.String())
	}
}

// Flow 5
// Create -> LastNotif (error)
func Test_NoNotificationsFlow(t *testing.T) {

	svc, err := NewWebhookService(nil)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := svc.Create(
		WebhookEndpoint{
			URL: "http://localhost:8080",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.LastNotificationSent(endpoint)
	if err != ErrRecordNotFound {
		t.Fatal(err)
	}

}

// Flow 6
// Create -> List -> Delete -> List
func Test_CreateListDeleteListFlow(t *testing.T) {

	svc, err := NewWebhookService(nil)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := svc.Create(
		WebhookEndpoint{
			URL: "http://localhost:8080",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	lsOne, err := svc.ListEndpoints()
	if err != nil {
		t.Fatal(err)
	}

	if len(*lsOne) != 1 {
		t.Fatal("Expected there to be one endpoint after creating it")
	}

	err = svc.Delete(endpoint)
	if err != nil {
		t.Fatal(err)
	}

	lsZero, err := svc.ListEndpoints()
	if err != nil {
		t.Fatal(err)
	}

	if len(*lsZero) != 0 {
		t.Fatal("Expected there to be no endpoints after deleting the only one")
	}

}

// Flow 7
// Create -> Update -> Fail
func Test_CreateUpdateFailFlow(t *testing.T) {

	svc, err := NewWebhookService(nil)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := svc.Create(
		WebhookEndpoint{
			URL: "http://localhost",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// successful given there's no change, URL is the same
	_, err = svc.UpdateURL(
		WebhookEndpoint{
			UUID: endpoint.UUID,
			URL:  "http://localhost",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// fail because no UUID is provided to identify the endpoint
	_, err = svc.UpdateURL(
		WebhookEndpoint{
			URL: "http://localhost:8080",
		},
	)
	if err != ErrEmptyEndpointUUID {
		t.Fatal(err)
	}
}
