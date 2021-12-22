package monk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type AbraKaDabra struct {
}

type AbraKaDabraOverride struct {
}

func (AbraKaDabraOverride) CollectionName() string {
	return "i_AM_different"
}

func TestCollectionName(t *testing.T) {

	assert.Equal(t, CollectionName(AbraKaDabra{}), "abra_ka_dabra")
	assert.Equal(t, CollectionName(&AbraKaDabra{}), "abra_ka_dabra")

	assert.Equal(t, CollectionName(AbraKaDabraOverride{}), "i_AM_different")
	assert.Equal(t, CollectionName(&AbraKaDabraOverride{}), "i_AM_different")
}

type AbcIndex struct {
	Field1 string `index:"true"`
	Field2 string `index:"true:-1"`
	Field3 string `index:"idx_something"`
	Field4 string `index:"idx_something2:-1"`
	Field5 int    `index:"idx_name(field1, field2:-1, field3)"`
	Field6 int    `index:"true|idx_that(field3:-1)"`
}

func TestGetAllIndexes(t *testing.T) {

	var list = GetAllIndexes(AbcIndex{})
	// assert.Len(t, list, 6)

	// Field1
	assert.Equal(t, "idx_field1", list[0].Name)
	assert.Equal(t, false, list[0].Unique)
	assert.Len(t, list[0].Fields, 1)
	assert.Equal(t, "field1", list[0].Fields[0])
	assert.Equal(t, 1, list[0].Order[0])

	// Field2
	assert.Equal(t, "idx_field2", list[1].Name)
	assert.Equal(t, false, list[1].Unique)
	assert.Len(t, list[1].Fields, 1)
	assert.Equal(t, "field2", list[1].Fields[0])
	assert.Equal(t, -1, list[1].Order[0])

	// Field3
	assert.Equal(t, "idx_something", list[2].Name)
	assert.Equal(t, false, list[2].Unique)
	assert.Len(t, list[2].Fields, 1)
	assert.Equal(t, "field3", list[2].Fields[0])
	assert.Equal(t, 1, list[2].Order[0])

	// Field4
	assert.Equal(t, "idx_something2", list[3].Name)
	assert.Equal(t, false, list[3].Unique)
	assert.Len(t, list[3].Fields, 1)
	assert.Equal(t, "field4", list[3].Fields[0])
	assert.Equal(t, -1, list[3].Order[0])

	// Field4
	assert.Equal(t, "idx_something2", list[3].Name)
	assert.Equal(t, false, list[3].Unique)
	assert.Len(t, list[3].Fields, 1)
	assert.Equal(t, "field4", list[3].Fields[0])
	assert.Equal(t, -1, list[3].Order[0])

	// Field5
	assert.Equal(t, "idx_name", list[4].Name)
	assert.Equal(t, false, list[4].Unique)
	assert.Len(t, list[4].Fields, 3)
	assert.Equal(t, "field1", list[4].Fields[0])
	assert.Equal(t, "field2", list[4].Fields[1])
	assert.Equal(t, "field3", list[4].Fields[2])
	assert.Equal(t, 1, list[4].Order[0])
	assert.Equal(t, -1, list[4].Order[1])
	assert.Equal(t, 1, list[4].Order[2])

	// Field6
	assert.Equal(t, "idx_field6", list[5].Name)
	assert.Equal(t, false, list[5].Unique)
	assert.Len(t, list[5].Fields, 1)
	assert.Equal(t, "field6", list[5].Fields[0])
	assert.Equal(t, 1, list[5].Order[0])
	//
	assert.Equal(t, "idx_that", list[6].Name)
	assert.Equal(t, false, list[6].Unique)
	assert.Len(t, list[6].Fields, 1)
	assert.Equal(t, "field3", list[6].Fields[0])
	assert.Equal(t, -1, list[6].Order[0])
}

type DefUniqueIndex struct {
	Field1 string `unique:"true"`
	Field2 string `unique:"true:-1"`
	Field3 string `unique:"idx_something"`
	Field4 string `unique:"idx_something2:-1"`
	Field5 int    `unique:"idx_name(field1, field2:-1, field3)"`
	Field6 int    `unique:"true|idx_that(field3:-1)"`
}

func TestGetAllUniqueIndexes(t *testing.T) {

	var list = GetAllIndexes(DefUniqueIndex{})
	// assert.Len(t, list, 6)

	// Field1
	assert.Equal(t, "idx_field1", list[0].Name)
	assert.Equal(t, true, list[0].Unique)
	assert.Len(t, list[0].Fields, 1)
	assert.Equal(t, "field1", list[0].Fields[0])
	assert.Equal(t, 1, list[0].Order[0])

	// Field2
	assert.Equal(t, "idx_field2", list[1].Name)
	assert.Equal(t, true, list[1].Unique)
	assert.Len(t, list[1].Fields, 1)
	assert.Equal(t, "field2", list[1].Fields[0])
	assert.Equal(t, -1, list[1].Order[0])

	// Field3
	assert.Equal(t, "idx_something", list[2].Name)
	assert.Equal(t, true, list[2].Unique)
	assert.Len(t, list[2].Fields, 1)
	assert.Equal(t, "field3", list[2].Fields[0])
	assert.Equal(t, 1, list[2].Order[0])

	// Field4
	assert.Equal(t, "idx_something2", list[3].Name)
	assert.Equal(t, true, list[3].Unique)
	assert.Len(t, list[3].Fields, 1)
	assert.Equal(t, "field4", list[3].Fields[0])
	assert.Equal(t, -1, list[3].Order[0])

	// Field4
	assert.Equal(t, "idx_something2", list[3].Name)
	assert.Equal(t, true, list[3].Unique)
	assert.Len(t, list[3].Fields, 1)
	assert.Equal(t, "field4", list[3].Fields[0])
	assert.Equal(t, -1, list[3].Order[0])

	// Field5
	assert.Equal(t, "idx_name", list[4].Name)
	assert.Equal(t, true, list[4].Unique)
	assert.Len(t, list[4].Fields, 3)
	assert.Equal(t, "field1", list[4].Fields[0])
	assert.Equal(t, "field2", list[4].Fields[1])
	assert.Equal(t, "field3", list[4].Fields[2])
	assert.Equal(t, 1, list[4].Order[0])
	assert.Equal(t, -1, list[4].Order[1])
	assert.Equal(t, 1, list[4].Order[2])

	// Field6
	assert.Equal(t, "idx_field6", list[5].Name)
	assert.Equal(t, true, list[5].Unique)
	assert.Len(t, list[5].Fields, 1)
	assert.Equal(t, "field6", list[5].Fields[0])
	assert.Equal(t, 1, list[5].Order[0])
	//
	assert.Equal(t, "idx_that", list[6].Name)
	assert.Equal(t, true, list[6].Unique)
	assert.Len(t, list[6].Fields, 1)
	assert.Equal(t, "field3", list[6].Fields[0])
	assert.Equal(t, -1, list[6].Order[0])
}

func TestCreateIndexes(t *testing.T) {

	// Check no indexes
	indexes := testConnection.Collection(DefUniqueIndex{}).Indexes()
	ctx, cancel := GetContext()
	defer cancel()
	cur, _ := indexes.List(ctx)

	var result []bson.M
	err := cur.All(context.TODO(), &result)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))

	// Create the indexes
	CreateIndexes(&testConnection, DefUniqueIndex{})

	// Check the count of indexes
	cur, _ = indexes.List(ctx)
	err = cur.All(context.TODO(), &result)
	assert.Nil(t, err)
	assert.Equal(t, 7+1 /*for _id*/, len(result))
}
