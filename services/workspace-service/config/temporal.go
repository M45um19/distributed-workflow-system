package config

import (
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
)

func ConnectTemporal(host string) client.Client {
	c, err := client.Dial(client.Options{
		HostPort: host,
	})
	if err != nil {
		log.Fatalf("Unable to create Temporal client: %v", err)
	}
	fmt.Println("Connected to Temporal successfully via URI")

	return c
}
