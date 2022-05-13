package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/centrifugal/centrifuge-go"
    "github.com/google/uuid"
    "github.com/hasura/go-graphql-client"
    "github.com/nats-io/nats.go"
    "io"
    "log"
    "net/http"
    "time"
)

type Status struct {
    Timestamp time.Time `json:"timestamp"`
    Message   string    `json:"message"`
}

type ScrapeJob struct {
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
    URL       string    `json:"url"`
}

type CannabisJSON struct {
    UID                     uuid.UUID `json:"uid"`
    Strain                  string    `json:"strain"`
    CannabinoidAbbreviation string    `json:"cannabinoid_abbreviation"`
    Cannabinoid             string    `json:"cannabinoid"`
    Terpene                 string    `json:"terpene"`
    MedicalUse              string    `json:"medical_use"`
    HealthBenefit           string    `json:"health_benefit"`
    Category                string    `json:"category"`
    Type                    string    `json:"type"`
    Buzzword                string    `json:"buzzword"`
    Brand                   string    `json:"brand"`
}

const token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2MzQwMDk2MTksImV4cCI6MTY2NTU0NTYxOSwiYXVkIjoic2FuZHBpdCIsInN1YiI6Im5vYm9keUBsb2NhbGRvbWFpbi5sb2NhbGhvc3QifQ.c7NKmDcUAJuRu-TB_128CAK0hqH9-8Mt65dyyifeJfo"

var (
    httpClient = http.Client{
        Timeout: time.Second * 5,
    }

    graphqlClient = graphql.NewClient(
        "http://hasura:8080/v1/graphql",
        &httpClient,
        // TODO: work out auth
        // oauth2.NewClient(
        //     context.Background(),
        //     oauth2.StaticTokenSource(
        //         &oauth2.Token{AccessToken: "s@ndp1t123!@#",
        //         },
        //     ),
        // ),
    )

    centrifugeClient = centrifuge.NewJsonClient("ws://centrifugo:8000/connection/websocket", centrifuge.DefaultConfig())

    centrifugeSubscription *centrifuge.Subscription

    workerID uuid.UUID
)

func publish(message string) error {
    status := Status{
        Timestamp: time.Now().UTC(),
        Message:   fmt.Sprintf("%v: %v", workerID.String(), message),
    }

    buf, err := json.Marshal(status)
    if err != nil {
        return err
    }

    log.Print(string(buf))

    _, err = centrifugeClient.Publish("status", buf)
    if err != nil {
        return err
    }

    return nil
}

func handler(m *nats.Msg) {
    scrapeJob := ScrapeJob{}

    err := json.Unmarshal(m.Data, &scrapeJob)
    if err != nil {
        log.Printf("warning: failed to unmarshal %#+v because %v", string(m.Data), err)
        return
    }

    if time.Now().UTC().After(scrapeJob.ExpiresAt) {
        log.Printf("warning: failed because job expired")
        return
    }

    response, err := httpClient.Get(scrapeJob.URL)
    if err != nil {
        log.Printf("warning: failed to get %#+v because %v", scrapeJob.URL, err)
        return
    }

    data, err := io.ReadAll(response.Body)
    if err != nil {
        log.Printf("warning: failed to read response body because %v", err)
        return
    }

    cannabisJSONs := make([]CannabisJSON, 0)

    err = json.Unmarshal(data, &cannabisJSONs)
    if err != nil {
        log.Printf("warning: failed to unmarshal %#+v because %v", string(data), err)
        return
    }

    log.Printf("%#+v", cannabisJSONs)

    err = publish("handling job")
    if err != nil {
        log.Fatal(err)
    }

    for _, cannabisJSON := range cannabisJSONs {
        mutation := struct {
            CreateCannabis struct {
                ID graphql.Int `graphql:"id"`
            } `graphql:"insert_cannabis_one(object: {brand: $brand, buzzword: $buzzword, cannabinoid: $cannabinoid, cannabinoid_abbreviation: $cannabinoid_abbreviation, category: $category, health_benefit: $health_benefit, medical_use: $medical_use, strain: $strain, terpene: $terpene, type: $type, uid: $uid})"`
        }{}

        variables := map[string]interface{}{
            "brand":                    graphql.String(cannabisJSON.Brand),
            "buzzword":                 graphql.String(cannabisJSON.Buzzword),
            "type":                     graphql.String(cannabisJSON.Type),
            "category":                 graphql.String(cannabisJSON.Category),
            "health_benefit":           graphql.String(cannabisJSON.HealthBenefit),
            "medical_use":              graphql.String(cannabisJSON.MedicalUse),
            "terpene":                  graphql.String(cannabisJSON.Terpene),
            "cannabinoid":              graphql.String(cannabisJSON.Cannabinoid),
            "cannabinoid_abbreviation": graphql.String(cannabisJSON.CannabinoidAbbreviation),
            "strain":                   graphql.String(cannabisJSON.Strain),
            "uid":                      graphql.String(cannabisJSON.UID.String()),
        }

        err = publish("running mutation")
        if err != nil {
            log.Fatal(err)
        }

        err = graphqlClient.Mutate(context.Background(), &mutation, variables)
        if err != nil {
            log.Printf("warning: failed to push %#+v to hasura because %v", cannabisJSON, err)
        }
    }

    err = publish("job handled")
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    var err error

    workerID, err = uuid.NewRandom()
    if err != nil {
        log.Fatalf("failed to generate workerID because %v", err)
    }

    centrifugeClient.SetToken(token)

    err = centrifugeClient.Connect()
    if err != nil {
        log.Fatalf("failed to connected to centrifugo because %v", err)
    }

    centrifugeSubscription, err = centrifugeClient.NewSubscription("status")
    if err != nil {
        log.Fatalf("failed to subscribe to centrifugo because %v", err)
    }

    err = publish("connected to centrifugo")
    if err != nil {
        log.Fatalf("failed to publish to centrifugo because %v", err)
    }

    defer func() {
        _ = centrifugeClient.Disconnect()
    }()

    nc, err := nats.Connect("nats://nats:4222")
    if err != nil {
        log.Fatalf("failed to connect to nats because %v", err)
    }

    err = publish("connected to nats")
    if err != nil {
        log.Fatal(err)
    }

    defer nc.Close()

    subscription, err := nc.QueueSubscribe("scrape_jobs", "scrape_jobs_queue", handler)
    if err != nil {
        log.Fatal(err)
    }

    err = publish("subscribed to scrape_jobs queue")
    if err != nil {
        log.Fatal(err)
    }

    defer func() {
        _ = subscription.Unsubscribe()
    }()

    err = publish("blocking on main sleep loop")
    if err != nil {
        log.Fatal(err)
    }

    for {
        time.Sleep(time.Second)
    }
}
