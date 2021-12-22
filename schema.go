package monk

import (
	"context"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/rightjoin/rutl/conv"
	"github.com/rightjoin/rutl/refl"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateCollection(mc *MongoConn, types ...interface{}) {

	// Setup Schema Validations

	// Setup Indexes (normal and unique)
	for _, t := range types {
		CreateIndexes(mc, t)
	}

	// Create initial records
	for _, t := range types {
		InsertInitalRecords(mc, t)
	}

}

// InsertInitalRecords invokes "PrePopulate" and inserts the returned
// values into the database
func InsertInitalRecords(mc *MongoConn, model interface{}) {

	ot := reflect.TypeOf(model)
	ov := reflect.ValueOf(model)

	// Dereference
	if ot.Kind() == reflect.Ptr {
		ot = ot.Elem()
		ov = ov.Elem()
	}

	// If "CollectionName" method exists, call it
	if _, ok := ot.MethodByName("PrePopulate"); ok {
		output := ov.MethodByName("PrePopulate").Call([]reflect.Value{})
		records, ok := output[0].Interface().([]interface{})
		if !ok {
			return
		}
		for _, rec := range records {
			rt := reflect.TypeOf(rec)
			if rt.Kind() == reflect.Map {
				// TODO
				// Invoke Insert
			} else {
				// TODO
				// json encode and decode into map
				// Invoke Insert
			}
		}
	}

}

// index|unique:"true|true:-1"
// index|unique:"idx_name"
// index|unique:"idx_name(field1,field2)"
func CreateIndexes(mc *MongoConn, model interface{}) {

	indexes := mc.Collection(model).Indexes()
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	indexesToCreate := GetAllIndexes(model)
	for _, idx := range indexesToCreate {

		m := bson.D{}
		for i := range idx.Fields {
			m = append(m, bson.E{Key: idx.Fields[i], Value: idx.Order[i]})
		}

		newIndex := mongo.IndexModel{
			Keys: m,
			Options: &options.IndexOptions{
				Unique: &idx.Unique,
				Name:   &idx.Name,
			},
		}

		data, err := indexes.CreateOne(ctx, newIndex)
		if err != nil {
			log.Error().
				Err(err).
				Str("Return value", data).
				Msg("Counld not create index")
		}
	}
}

func GetAllIndexes(model interface{}) []MonkIndex {
	var list = []MonkIndex{}
	fields := refl.NestedFields(model)
	for i := 0; i < len(fields); i++ {
		list = append(list, GetFieldIndexes(fields[i])...)
	}

	return list
}

func GetFieldIndexes(f reflect.StructField) []MonkIndex {
	var list = []MonkIndex{}
	indexAll := f.Tag.Get("index")
	uniqueAll := f.Tag.Get("unique")
	name := FieldKey(f)

	if indexAll != "" {
		indexes := strings.Split(indexAll, "|")
		for _, index := range indexes {
			list = append(list, GetIndex(index, name, false))
		}
	}

	if uniqueAll != "" {
		uniques := strings.Split(uniqueAll, "|")
		for _, unique := range uniques {
			list = append(list, GetIndex(unique, name, true))
		}
	}

	return list
}

type MonkIndex struct {
	Name   string
	Unique bool
	Fields []string
	Order  []int
}

// Given a struct, or address of a struct, get the
// appropriate collection name for storing that struct
func CollectionName(model interface{}) string {
	if name, ok := model.(string); ok {
		return name
	}

	// Indirect
	t := reflect.TypeOf(model)
	v := reflect.ValueOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	// If "CollectionName" method exists, call it
	if _, ok := t.MethodByName("CollectionName"); ok {
		col := v.MethodByName("CollectionName").Call([]reflect.Value{})
		return col[0].String()
	}

	return strings.TrimSpace(conv.CaseSnake(t.Name()))
}

func GetIndex(tagData, fieldName string, unique bool) MonkIndex {
	idx := MonkIndex{Unique: unique, Fields: []string{}, Order: []int{}}
	lbrace := strings.Index(tagData, "(")

	if ok, _ := regexp.MatchString("true(:(-|\\+)\\d+)?", tagData); ok {
		// index:"true" | index:"true:+1"
		split := strings.Split(tagData, ":")
		if len(split) == 1 {
			idx.Name = "idx_" + fieldName
			idx.Fields = append(idx.Fields, fieldName)
			idx.Order = append(idx.Order, 1)
		} else {
			idx.Name = "idx_" + fieldName
			idx.Fields = append(idx.Fields, fieldName)
			idx.Order = append(idx.Order, conv.IntOr(split[1], 1))
		}
	} else if lbrace == -1 {
		// index:"idx_name" | index:"idx_name:+1"
		split := strings.Split(tagData, ":")
		idx.Name = tagData
		if len(split) == 1 {
			idx.Name = tagData
			idx.Fields = append(idx.Fields, fieldName)
			idx.Order = append(idx.Order, 1)
		} else {
			idx.Name = split[0]
			idx.Fields = append(idx.Fields, fieldName)
			idx.Order = append(idx.Order, conv.IntOr(split[1], 1))
		}
	} else {
		// index:"idx_name(field1,field2)" | index:"idx_name(field1:+1,field2:-1)"
		idx.Name = tagData[0:lbrace]
		rbrace := strings.Index(tagData, ")")
		csv := tagData[lbrace+1 : rbrace]
		flds := strings.Split(csv, ",")
		for i := range flds {
			split := strings.Split(flds[i], ":")
			if len(split) == 1 {
				idx.Fields = append(idx.Fields, strings.TrimSpace(split[0]))
				idx.Order = append(idx.Order, 1)
			} else {
				idx.Fields = append(idx.Fields, strings.TrimSpace(split[0]))
				idx.Order = append(idx.Order, conv.IntOr(split[1], 1))
			}
		}
	}
	return idx
}
