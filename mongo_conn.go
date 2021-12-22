package monk

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConn struct {
	ConnStr    string // must be present
	DB         string // must be present
	DBSuffix   string // optional
	Coll       string // optional
	CollSuffix string // optional

	// for reuse
	client *mongo.Client
}

func (mc *MongoConn) GetClient() *mongo.Client {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mc.ConnStr))
	if err != nil {
		log.Error().
			Err(err).
			Str("connection string", mc.ConnStr).
			Msg("unable to open connection")
		return nil
	}
	return client
}

func (mc *MongoConn) Database() *mongo.Database {
	if mc.DBSuffix == "" {
		return mc.GetClient().Database(mc.DB)
	} else {
		return mc.GetClient().Database(fmt.Sprintf("%s-%s", mc.DB, mc.DBSuffix))
	}
}

func (mc *MongoConn) Collection(model ...interface{}) *mongo.Collection {

	coll := ""
	if len(model) == 0 {
		coll = mc.Coll
	} else {
		coll = CollectionName(model[0])
	}

	if mc.CollSuffix == "" {
		return mc.Database().Collection(coll)
	} else {
		return mc.Database().Collection(fmt.Sprintf("%s-%s", coll, mc.CollSuffix))
	}
}

func NewMongoConnToDB(envUUID string) MongoConn {
	return MongoConn{}
}

func NewMongoConnToColl(envUUID string, collUUID string) MongoConn {
	return MongoConn{}
}
