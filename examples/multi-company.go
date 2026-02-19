// Multi-company integration example
package main

import (
	"context"
	"fmt"
	"os"

	oway "github.com/Oway-Inc/oway-sdk/packages/go"
)

func main() {
	// M2M credentials from Sales Engineering
	client, _ := oway.New(oway.Config{
		ClientID:     os.Getenv("OWAY_M2M_CLIENT_ID"),
		ClientSecret: os.Getenv("OWAY_M2M_CLIENT_SECRET"),
		APIKey:       "oway_sk_default", // Optional default
	})

	ctx := context.Background()

	// Per-company API keys
	keys := map[string]string{
		"acme": "oway_sk_acme_123",
		"widgets": "oway_sk_widgets_456",
	}

	// Quote for ACME (uses their API key)
	quoteA, _ := client.RequestQuoteForCompany(ctx, &oway.QuoteRequest{}, keys["acme"])
	fmt.Printf("ACME: %s\n", quoteA.Id)

	// Quote for Widgets (uses their API key)
	quoteB, _ := client.RequestQuoteForCompany(ctx, &oway.QuoteRequest{}, keys["widgets"])
	fmt.Printf("Widgets: %s\n", quoteB.Id)
}
