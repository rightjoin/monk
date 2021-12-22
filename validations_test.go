package monk

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func keys(errors map[string][]string) []string {
	out := []string{}
	for k, _ := range errors {
		out = append(out, k)
	}
	return out
}

func TestInsertableChecks(t *testing.T) {

	// If you insert a field tagged insert:no
	// then an error is generated
	{
		a := struct {
			Field1 string `insert:"no"`
		}{}
		errs2 := checkInsertableUpdatable(a, INSERT, map[string]interface{}{"field1": "abc"})
		assert.Equal(t, 1, len(errs2))
		assert.Equal(t, "field1", errs2[0].Reference)

		// And this works even if you pass
		// the address of struct
		{
			errs2 = checkInsertableUpdatable(&a, INSERT, map[string]interface{}{"field1": "abc"})
			assert.Equal(t, 1, len(errs2))
			assert.Equal(t, "field1", errs2[0].Reference)
		}
	}

	// If you insert a field tagged insert:no
	// even at a nested struct level,
	// then an error is generated
	{
		b := struct {
			FieldA struct {
				Field1 string `insert:"no"`
			}
		}{}
		errs2 := checkInsertableUpdatable(b, INSERT, map[string]interface{}{"field_a": map[string]interface{}{"field1": "abc"}})
		assert.Equal(t, 1, len(errs2))
		assert.Equal(t, "field_a.field1", errs2[0].Reference)

		// And this works even if you pass
		// the address of struct
		{
			errs2 := checkInsertableUpdatable(&b, INSERT, map[string]interface{}{"field_a": map[string]interface{}{"field1": "abc"}})
			assert.Equal(t, 1, len(errs2))
			assert.Equal(t, "field_a.field1", errs2[0].Reference)
		}
	}

	// If you insert a field tagged insert:no
	// even at a nested pointer-struct level,
	// then an error is generated
	{
		c := struct {
			FieldB *struct {
				Field1 string `insert:"no"`
			}
		}{}
		errs2 := checkInsertableUpdatable(c, INSERT, map[string]interface{}{"field_b": map[string]interface{}{"field1": "abc"}})
		assert.Equal(t, 1, len(errs2))
		assert.Equal(t, "field_b.field1", errs2[0].Reference)

		// And this works even if you pass
		// the address of struct
		{
			errs2 := checkInsertableUpdatable(&c, INSERT, map[string]interface{}{"field_b": map[string]interface{}{"field1": "abc"}})
			assert.Equal(t, 1, len(errs2))
			assert.Equal(t, "field_b.field1", errs2[0].Reference)
		}
	}

	// insert:yes field is missing
	// then you get errors
	{
		a := struct {
			Field1 string `insert:"yes"`
		}{}
		errs := checkInsertableUpdatable(a, INSERT, map[string]interface{}{})
		assert.Equal(t, 1, len(errs))
		assert.Equal(t, "field1", errs[0].Reference)

		// Errors go away wnen data is provided
		{
			errs := checkInsertableUpdatable(a, INSERT, map[string]interface{}{"field1": "abc"})
			assert.Equal(t, 0, len(errs))
		}
	}

	// insert:yes field is missing,
	// at a nested struct level
	// then you get errors
	{
		a := struct {
			FieldA struct {
				Field1 string `insert:"yes"`
			}
		}{}
		errs := checkInsertableUpdatable(a, INSERT, map[string]interface{}{})
		assert.Equal(t, 1, len(errs))
		assert.Equal(t, "field_a.field1", errs[0].Reference)

		// Errors go away wnen data is provided
		{
			errs := checkInsertableUpdatable(a, INSERT, map[string]interface{}{"field_a": map[string]interface{}{"field1": "abc"}})
			assert.Equal(t, 0, len(errs))
		}
	}

	// Field is optional to insert
	// Missing it is OK
	{
		a := struct {
			Field1 string `insert:"opt"`
		}{}
		errs := checkInsertableUpdatable(a, INSERT, map[string]interface{}{})
		assert.Equal(t, 0, len(errs))
	}

	// Defualt value for insert is [opt]
	{
		a := struct {
			Field1 string /*`insert:"opt"`*/
		}{}
		errs := checkInsertableUpdatable(a, INSERT, map[string]interface{}{})
		assert.Equal(t, 0, len(errs))
	}
}

func TestInsertFieldUpdateAction(t *testing.T) {

	var ok bool

	// Field is tagged to insert:NO
	// There should be no impact on Update action
	{
		a := struct {
			Field1 string `insert:"no"`
		}{}
		ok, _ = Validate(a, UPDATE, map[string]interface{}{"field1": "abc"})
		assert.Equal(t, true, ok)
	}
}

func TestDefaultsOnInsert(t *testing.T) {

	// For an optional field, if a default value is provided
	// then it should get used
	{
		a := struct {
			Field1 string `default:"abra"`
		}{}
		m := map[string]interface{}{}
		errs := provideDefualts(a, INSERT, m)
		assert.Equal(t, 0, len(errs))
		assert.Equal(t, "abra", m["field1"])
	}

	// When the field is insert:"no",
	// any specified default values should not kick in
	{
		a := struct {
			Field1 string `default:"abra" insert:"no"`
		}{}
		m := map[string]interface{}{}
		errs := provideDefualts(a, INSERT, m)
		assert.Equal(t, 0, len(errs))
		_, found := m["field1"]
		assert.False(t, found)
	}
}

func TestFieldConversion(t *testing.T) {

	var m map[string]interface{}

	{
		a := struct {
			MongoStore
			Field1 int ``
		}{}
		m = map[string]interface{}{"field1": "12345"}
		errs := convertFieldType(a, INSERT, m)
		_, isInt := m["field1"].(int)
		assert.Equal(t, 0, len(errs))
		assert.True(t, isInt)
	}

	{
		b := struct {
			MongoStore
			Field1 int64 ``
		}{}
		m = map[string]interface{}{"field1": "12345"}
		errs := convertFieldType(b, INSERT, m)
		_, isInt64 := m["field1"].(int64)
		assert.Equal(t, 0, len(errs))
		assert.True(t, isInt64)
	}
}

func TestTimedFields(t *testing.T) {

	var m map[string]interface{}

	a := struct {
		MongoStore
		Timed
	}{}
	m = map[string]interface{}{}
	errs := populateTimedFields(a, INSERT, m)
	assert.Equal(t, 0, len(errs))

	_, isTime := m["created_at"].(time.Time)
	assert.True(t, isTime)

	_, isTime = m["updated_at"].(time.Time)
	assert.True(t, isTime)
}

func TestPopulateAutoFields(t *testing.T) {

	{
		a := struct {
			Field1 string `auto:"prefix:555-;uuid"`
		}{}
		m := map[string]interface{}{}
		errs := populateAutoFields(a, INSERT, m)
		assert.Equal(t, 0, len(errs))

		str, found := m["field1"].(string)
		assert.True(t, found)
		assert.True(t, strings.HasPrefix(str, "555-"))
	}

	{
		a := struct {
			Field1 string `auto:"alphanum(5)"`
		}{}
		m := map[string]interface{}{}
		errs := populateAutoFields(a, INSERT, m)
		assert.Equal(t, 0, len(errs))

		str, found := m["field1"].(string)
		assert.True(t, found)
		assert.Regexp(t, `^[a-z0-9A-Z]{5}$`, str)
	}
}

func TestVerifications(t *testing.T) {

	// verify:email
	{
		a := struct {
			Field1 string `verify:"email"`
		}{}
		errs := verifyInputs(a, INSERT, map[string]interface{}{
			"field1": "abc @ def . com",
		})
		assert.Equal(t, 1, len(errs))

		// pass
		{
			errs := verifyInputs(a, INSERT, map[string]interface{}{
				"field1": "abc@def.com",
			})
			assert.Equal(t, 0, len(errs))
		}
	}

	// Regular expression
	// verify:rex(...)
	{
		a := struct {
			Field1 string `verify:"rex(^a+$)"`
		}{}
		errs := verifyInputs(a, INSERT, map[string]interface{}{
			"field1": "abc",
		})
		assert.Equal(t, 1, len(errs))

		// pass
		{
			errs := verifyInputs(a, INSERT, map[string]interface{}{
				"field1": "aaa",
			})
			assert.Equal(t, 0, len(errs))
		}
	}

	// Complex Regular expression
	{
		a := struct {
			Mobile string `verify:"rex(^[6-9]\\d{9}$)"`
		}{}
		errs := verifyInputs(a, INSERT, map[string]interface{}{
			"mobile": "abc",
		})
		assert.Equal(t, 1, len(errs))

		// pass
		{
			errs := verifyInputs(a, INSERT, map[string]interface{}{
				"mobile": "9977887799",
			})
			assert.Equal(t, 0, len(errs))
		}
	}

	// enum
	{
		a := struct {
			Color string `verify:"enum(green|yellow|red)"`
		}{}
		errs := verifyInputs(a, INSERT, map[string]interface{}{
			"color": "black",
		})
		assert.Equal(t, 1, len(errs))

		// pass
		{
			errs := verifyInputs(a, INSERT, map[string]interface{}{
				"color": "red",
			})
			assert.Equal(t, 0, len(errs))
		}
	}

}

func TestTrim(t *testing.T) {

	a := struct {
		Field1 string `trim:"no"`
		Field2 string ``
	}{}
	m := map[string]interface{}{
		"field1": " 1 ",
		"field2": " 2 ",
	}
	errs := trimFields(a, INSERT, m)
	assert.Equal(t, 0, len(errs))

	assert.Equal(t, " 1 ", m["field1"])
	assert.Equal(t, "2", m["field2"])
}

func TestStructGetFieldTypeByJsonKey(t *testing.T) {
	a := struct {
		FieldC1 string
		FieldC2 int
		Parent  struct {
			FieldP1 string
		} `json:"abc"`
	}{}
	t1, _ := StructGetFieldTypeByJsonKey(a, "abc.field_p1")
	t2, _ := StructGetFieldTypeByJsonKey(a, "field_c1")
	t3, _ := StructGetFieldTypeByJsonKey(a, "field_c2")

	assert.Equal(t, "string", t1.String())
	assert.Equal(t, "string", t2.String())
	assert.Equal(t, "int", t3.String())
}
