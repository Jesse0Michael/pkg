package config

import (
	"context"

	firestore "cloud.google.com/go/firestore/apiv1"
	"google.golang.org/api/option"
)

type FirestoreConfig struct {
	APIKey string `envconfig:"FIRESTORE_API_KEY"`
}

func NewFirestoreClient(cfg FirestoreConfig) (*firestore.Client, error) {
	firestoreClient, err := firestore.NewClient(context.Background(), option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, err
	}
	return firestoreClient, nil
}
