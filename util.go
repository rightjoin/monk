package monk

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetConnString() (string, error) {
	return "mongodb://admin:admin@127.0.0.1:27017/", nil
}

func GetConn() (*mongo.Client, error) {
	connStr, err := GetConnString()
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connStr))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func CloseConn(m *mongo.Client) {
	if err := m.Disconnect(context.TODO()); err != nil {
		log.Error().Msg("unable to close connection")
	}
}

func GetContextWaitTime() time.Duration {
	return 10 * time.Second
}

func GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), GetContextWaitTime())
}
