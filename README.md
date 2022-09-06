![tests](https://github.com/tech-branch/iron-hook/actions/workflows/tests.yml/badge.svg)

# iron hook

You're looking at a simple plug-in webhook service for your Golang application. 

It's agnostic of your wider architectural choices and doesnt make too many assumptions about your application.

It persists state like registered endpoints and all sent notifications to a database controlled by the [Gorm](https://gorm.io/) ORM, but the database can be as simpe as an in-memory sqlite if you don't require persistence between processes.

## How it works

The `service` runs around two structs: `WebhookEndpoint` and a `WebhookNotification`. 

You can create the service like this:

```Golang
service, err := ironhook.NewWebhookService(nil)
```

Then we can create a new `WebhookEndpoint` like this:

```Golang
e := ironhook.WebhookEndpoint{
    URL: "http://localhost:8080"
}

endpoint := service.Create(e)
```

At first, the `endpoint` is `Unverified`, but we can fix that quickly by Verifying it:

```Golang
verified_endpoint, err := service.Verify(endpoint)
```

You may want to persist the UUID of this endpoint to be able to reference it in the future

```Golang
endpoint_id := verified_endpoint.UUID // save to your database
```

And with that out of the way, we can send notifications

```Golang
uuid_trace := uuid.Must(uuid.NewV4())

notification := ironhook.NewNotification(
    event: uuid_trace,
    topic: "Batch processing #1 complete",
    body: "All batches have been processed"
)

err := service.Notify(verified_endpoint, notification)
```

## Returning customer

If you want to send another notification to the same endpoint, the starting point will be the UUID of the endpoint.

```Golang
endpoint_id := verified_endpoint.UUID // or read from your DB
```

With that, we can create a new notification and send it in one go:

```Golang
endpoint := ironhook.WebhookEndpoint{
    UUID: endpoint_id,
}

notification := ironhook.NewNotification(
    event: uuid_trace,
    topic: "Batch processing #2 complete",
    body: "All batches have been processed"
)

err := service.Notify(endpoint, notification)
```


## Verification

Every Endpoint has to be verified before it can be used for notifications.

The verification process is relatively simple and best shown with an example. 

Given an endpoint like this: 

`webhooks.example.com`

A verification request will be sent to 

`webhooks.example.com/verification` 

with an `id` query parameter cointaining a `uuid` identifier of the endpoint, like this: `id=6551e000-947a-40a4-948d-18d01e3660d4`

Putting it together, a request like this will be made: 

`webhooks.example.com/verification?id=6551e000-947a-40a4-948d-18d01e3660d4`

And **in response to this request, we expect to see the uuid as body of the response.**

In Golang, one of the simpler approaches to the verification would be something like this:

```Golang
func VerificationHandler(w http.ResponseWriter, r *http.Request) {

    qID, ok := r.URL.Query()["id"]
    
    if !ok || len(qID[0]) < 1 {
		fmt.Fprintf(w, "URL Parameter <id> is missing")
		return
	}
    
    fmt.Fprint(w, qID[0])
}
```

Which you can also see in the examples section: [_examples/receiver/http_handlers.go](_examples/receiver/http_handlers.go)

## Configuration

### Database

By default, if no environment variables are specified, the service will persist state to an in-memory sqlite database.

If you want to have it stored in a file, you might want to specify the below:

```
HOOK_DB_DSN=webhooks.db
```

Available database engine options are: 

- postgres
    - cockroach
- mysql
- sqlserver
- sqlite

Naturally, you can use databases like `cockroachdb` since they may be compatible with one of the supported engines. In this example, `cockroachdb` is compatible with `postgres`.

In order to specify an engine you can use:

```
HOOK_DB_ENGINE=postgres
```

And specify the connection string for it:

```
HOOK_DB_DSN=postgres://postgres:123456@localhost:5432/mywebapp
```

### Logging

The service uses [zap](https://github.com/uber-go/zap) for logging, at the moment, you can configure the logging level:

```
HOOK_LOG_LEVEL=warn
```

By default, `info` level is used.

If you specify an unsupported log level, the configuration will default to `info`.

### HTTP Client

You can also tell ironhook to use your pre-configured http client if the default isn't what you're after.

```Golang
http_client = &http.Client{
    Timeout: time.Second * 3,
}
service, err := ironhook.NewWebhookService(http_client)
```

The above is also the default configuration. 

