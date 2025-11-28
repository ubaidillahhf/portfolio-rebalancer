package elasticsearch

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

var client *elasticsearch.Client

// GetClient returns the Elasticsearch client instance
func GetClient() *elasticsearch.Client {
	return client
}

// Init initializes elasticsearch connection with retry logic
func Init() error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTICSEARCH_URL"),
		},
	}

	var esClient *elasticsearch.Client
	var err error

	for i := 1; i <= 5; i++ {
		esClient, err = elasticsearch.NewClient(cfg)
		if err != nil {
			log.Printf("Failed to create Elasticsearch client: %v", err)
		} else {
			_, err = esClient.Info()
			if err == nil {
				log.Println("Connected to Elasticsearch")
				client = esClient
				return nil
			}
			log.Printf("Elasticsearch client created, but ES not ready: %v", err)
		}

		log.Printf("Retrying connection to Elasticsearch... (%d/5)", i)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("failed to connect to Elasticsearch after retries: %w", err)
}
