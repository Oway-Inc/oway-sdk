// Multi-company integration example.
//
// A single SDK client serves several companies by attaching a per-request
// API key to context with WithCompanyAPIKey.
package main

import (
	"context"
	"fmt"
	"os"

	oway "github.com/Oway-Inc/oway-sdk/packages/go"
)

func main() {
	client, err := oway.New(oway.Config{
		ClientID:     os.Getenv("OWAY_M2M_CLIENT_ID"),
		ClientSecret: os.Getenv("OWAY_M2M_CLIENT_SECRET"),
		// Optional default; per-request keys below override it.
		APIKey: "oway_sk_default",
	})
	if err != nil {
		panic(err)
	}

	keys := map[string]string{
		"acme":    "oway_sk_acme_123",
		"widgets": "oway_sk_widgets_456",
	}

	for tenant, apiKey := range keys {
		ctx := oway.WithCompanyAPIKey(context.Background(), apiKey)
		quote, err := client.RequestQuote(ctx, &oway.QuoteRequest{ /* ... */ })
		if err != nil {
			fmt.Printf("%s: %v\n", tenant, err)
			continue
		}
		fmt.Printf("%s: %s\n", tenant, *quote.Id)
	}
}
