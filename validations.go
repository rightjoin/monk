package monk

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/outerjoin/do"
	"github.com/rightjoin/rutl/conv"
)

/*
	FIELD VALIDATION

	insert: yes/no/[opt]
	update: yes/no/[opt]
	trim:   [yes]/no
	valid: ??
	enum: ??
	min: ??
	max: ??
	Recursive: ??
	auto:
		prefix:
		uuid | alphanum(12)
	verify:
		email:
		rex(...)
		enum(abc|def|ghi)

*/

type FieldErrors map[string][]string

func (fe FieldErrors) Add(issue string, keys ...string) {
	finalKey := ""
	if len(keys) > 0 {
		finalKey = strings.Join(keys, ".")
	}
	list, found := fe[finalKey]
	if found {
		fe[finalKey] = append(list, issue)
	} else {
		fe[finalKey] = []string{issue}
	}
}

func LoopAndAdd(st reflect.StructField, prefix string, out map[string]reflect.Type) {

	t := do.TypeDereference(st.Type)
	kind := t.Kind()

	isStruct := kind == reflect.Struct
	isTime := do.TypeIsTime(t)

	if isTime || kind == reflect.Bool || kind == reflect.String ||
		kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 ||
		kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int || kind == reflect.Int64 ||
		kind == reflect.Float32 || kind == reflect.Float64 {
		currKey := st.Tag.Get("json")
		if currKey == "" {
			currKey = conv.CaseSnake(st.Name)
		}
		if prefix != "" {
			currKey = prefix + "." + currKey
		}
		out[currKey] = t
	} else if isStruct {
		toPass := ""
		currKey := st.Tag.Get("json")
		if currKey == "" {
			toPass = prefix
		} else {
			if prefix != "" {
				toPass = prefix + "." + currKey
			} else {
				toPass = currKey
			}
		}
		for i := 0; i < t.NumField(); i++ {
			LoopAndAdd(t.Field(i), toPass, out)
		}
	}
}

func StructGetFieldTypeByJsonKey(modelType interface{}, jsonFieldKey string) (reflect.Type, bool) {

	allFields := map[string]reflect.Type{}
	// TODO:
	// cache allFields, so for same struct you don't create it again and again

	ot := do.TypeOf(modelType)

	for i := 0; i < ot.NumField(); i++ {
		LoopAndAdd(ot.Field(i), "", allFields)
	}

	if t, ok := allFields[jsonFieldKey]; ok {
		return t, true
	}

	return reflect.TypeOf(nil), false
}

func TraverseModel(modelType interface{}, data do.Map, errs FieldErrors, op func(reflect.StructField, do.Map, ...string), key ...string) {

	ot := do.TypeOf(modelType)
	ot = do.TypeDereference(ot)

	for i := 0; i < ot.NumField(); i++ {

		sf := ot.Field(i)
		fname := FieldKey(sf)

		ft := do.TypeDereference(sf.Type)

		if ft.Kind() == reflect.Struct && !do.TypeIsTime(ft) {
			if !data.HasKey(fname) {
				// Pass an empty map
				innerData := map[string]interface{}{}
				TraverseModel(ft, innerData, errs, op, append(key, fname)...)
				if len(innerData) > 0 { // if traverse call modified innerData, then use it in data
					data[fname] = innerData
				}
			} else { // key exists
				goMap, isMap := data.GetOr(fname, false).(map[string]interface{})
				if isMap {
					TraverseModel(ft, goMap, errs, op, append(key, fname)...)
				} else {
					issue := fmt.Sprintf("field '%s' expected dict, but found literal", fname)
					errs.Add(issue, append(key, fname)...)
				}
			}
		} else {
			op(sf, data, append(key, fname)...)
		}
	}
}

type FieldName struct {
	Tag  string
	Func func(reflect.StructField) string
}

func (fn FieldName) Get(f reflect.StructField) string {
	if fn.Tag != "" {
		return f.Tag.Get(fn.Tag)
	}

	if fn.Func != nil {
		return fn.Func(f)
	}

	return conv.CaseSnake(f.Name)
}

func verifyInputs(modelType interface{}, action int, data do.Map) []do.ErrorReference {
	errs := []do.ErrorReference{}

	// Input validations as defined in 'verify' tag
	verify := func(fld reflect.StructField, data do.Map, keys ...string) []do.ErrorReference {
		fname := keys[len(keys)-1]
		if (action == INSERT || action == UPDATE) && data.HasKey(fname) {
			checks := getFieldTests(fld)
			for _, check := range checks {
				if success, message := check.Verify(fld.Type, data[fname]); !success {
					errs = append(errs, do.ErrorReference{
						Message:   message,
						Reference: strings.Join(keys, "."),
					})
				}
			}
		}
		return nil
	}
	do.StructWalk(modelType, do.WalkConfig{"json"}, data, verify)
	return errs
}

func convertFieldType(modelType interface{}, action int, data do.Map) []do.ErrorReference {
	errs := []do.ErrorReference{}

	// Convert types of fields from STRING to appropriate type
	// as it is specified in the Struct
	convert := func(fld reflect.StructField, data do.Map, keys ...string) []do.ErrorReference {
		fname := keys[len(keys)-1]
		inp, found := data[fname]
		inpStr, isStr := inp.(string)
		expType := fld.Type.String()

		if found && (action == INSERT || action == UPDATE) && isStr && expType != "string" {
			val, err := do.ParseType(inpStr, do.TypeDereference(fld.Type))
			if err == nil {
				data[fname] = val
			} else {
				errs = append(errs, do.ErrorReference{
					Message:   err.Error(),
					Reference: strings.Join(keys, "."),
				})
			}
		}

		return nil
	}

	do.StructWalk(modelType, do.WalkConfig{"json"}, data, convert)
	return errs
}

func populateTimedFields(modelType interface{}, action int, data do.Map) []do.ErrorReference {
	errs := []do.ErrorReference{}
	isMongoStore := do.TypeComposedOf(modelType, MongoStore{})

	// Manage timestamp fields (inserted_at / updated_at)
	// during insert / update of records - do this for only
	// MongoStores for now

	// setTimed := func(fld reflect.StructField, data do.Map, keys ...string) []do.ErrorReference {
	if isMongoStore && do.TypeComposedOf(modelType, Timed{}) {
		now := time.Now()
		switch action {
		case INSERT:
			data["created_at"] = now
			data["updated_at"] = now
		case UPDATE:
			data["updated_at"] = now
		}
	}
	// 	return nil
	// }
	// do.StructWalk(modelType, do.WalkConfig{"json"}, data, setTimed)
	return errs
}

func populateAutoFields(modelType interface{}, action int, data do.Map) []do.ErrorReference {
	errs := []do.ErrorReference{}
	isMongoStore := do.TypeComposedOf(modelType, MongoStore{})

	// Set fields marked auto - to give them appropriate value upon insertion
	setAuto := func(fld reflect.StructField, data do.Map, keys ...string) []do.ErrorReference {
		fname := keys[len(keys)-1]
		if action == INSERT && !data.HasKey(fname) && fld.Tag.Get("auto") != "" {
			auto := parseAutoTag(fld)
			if auto != nil {
				if isMongoStore && fname == "id" && fld.Tag.Get("bson") != "" {
					// give preference to bson
					data[fld.Tag.Get("bson")] = auto.Generate()
				} else {
					data[fname] = auto.Generate()
				}
			}
		}
		return nil
	}
	do.StructWalk(modelType, do.WalkConfig{"json"}, data, setAuto)
	return errs
}

func trimFields(modelType interface{}, action int, data do.Map) []do.ErrorReference {
	errs := []do.ErrorReference{}

	// Trim any input strings fields, unless markeed no (trim=no)
	trim := func(fld reflect.StructField, data do.Map, keys ...string) []do.ErrorReference {
		fname := keys[len(keys)-1]
		if (action == INSERT || action == UPDATE) && data.HasKey(fname) && fld.Tag.Get("trim") != "no" {
			str, isString := data[fname].(string)
			if isString {
				data[fname] = strings.TrimSpace(str)
			}
		}
		return nil
	}
	do.StructWalk(modelType, do.WalkConfig{"json"}, data, trim)
	return errs
}

func provideDefualts(modelType interface{}, action int, data do.Map) []do.ErrorReference {
	errs := []do.ErrorReference{}

	// During inserts, if input fields are not provided and a default value is provided
	// in the field tags then do use it
	setDefaults := func(fld reflect.StructField, data do.Map, keys ...string) []do.ErrorReference {
		fname := keys[len(keys)-1]
		defStr := fld.Tag.Get("default")
		if action == INSERT && !data.HasKey(fname) && defStr != "" && fld.Tag.Get("insert") != "no" {
			data[fname] = defStr
		}
		return nil
	}
	do.StructWalk(modelType, do.WalkConfig{"json"}, data, setDefaults)
	return errs
}

func checkInsertableUpdatable(modelType interface{}, action int, data do.Map) []do.ErrorReference {
	errs := []do.ErrorReference{}

	// Do validations for those fields wherein input fields are extra or
	// input fields are expected but missing
	checkInsertUpdate := func(fld reflect.StructField, data do.Map, keys ...string) []do.ErrorReference {
		fname := keys[len(keys)-1]

		switch action {
		case INSERT:
			if data.HasKey(fname) && fld.Tag.Get("insert") == "no" {
				issue := fmt.Sprintf("field '%s' cannot be given a value (%v) upon insertion", fname, data.GetOr(fname, nil))
				errs = append(errs, do.ErrorReference{issue, strings.Join(keys, ".")})
			}
			if !data.HasKey(fname) && fld.Tag.Get("insert") == "yes" {
				issue := fmt.Sprintf("field '%s' needs a value upon insertion", fname)
				errs = append(errs, do.ErrorReference{issue, strings.Join(keys, ".")})
			}
		case UPDATE:
			if data.HasKey(fname) && fld.Tag.Get("update") == "no" {
				issue := fmt.Sprintf("field '%s' cannot be given a value (%v) upon updation", fname, data.GetOr(fname, nil))
				errs = append(errs, do.ErrorReference{issue, strings.Join(keys, ".")})
			}
		}
		return nil
	}
	do.StructWalk(modelType, do.WalkConfig{"json"}, data, checkInsertUpdate)
	return errs
}

func Validate(modelType interface{}, action int, data do.Map) (success bool, issues map[string][]string) {
	errs := FieldErrors(map[string][]string{})

	isMongoStore := do.TypeComposedOf(modelType, MongoStore{})

	// Do validations for those fields wherein input fields are extra or
	// input fields are expected but missing
	checkInsertUpdate := func(fld reflect.StructField, data do.Map, keys ...string) {
		fname := keys[len(keys)-1]

		switch action {
		case INSERT:
			if data.HasKey(fname) && fld.Tag.Get("insert") == "no" {
				issue := fmt.Sprintf("field '%s' cannot be given a value (%v) upon insertion", fname, data.GetOr(fname, nil))
				errs.Add(issue, keys...)
			}
			if !data.HasKey(fname) && fld.Tag.Get("insert") == "yes" {
				issue := fmt.Sprintf("field '%s' needs a value upon insertion", fname)
				errs.Add(issue, keys...)
			}
		case UPDATE:
			if data.HasKey(fname) && fld.Tag.Get("update") == "no" {
				issue := fmt.Sprintf("field '%s' cannot be given a value (%v) upon updation", fname, data.GetOr(fname, nil))
				errs.Add(issue, keys...)
			}
		}
	}
	TraverseModel(modelType, data, errs, checkInsertUpdate)

	// During inserts, if input fields are not provided and a default value is provided
	// in the field tags then do use it
	setDefaults := func(fld reflect.StructField, data do.Map, keys ...string) {
		fname := keys[len(keys)-1]
		defStr := fld.Tag.Get("default")
		if action == INSERT && !data.HasKey(fname) && defStr != "" && fld.Tag.Get("insert") != "no" {
			data[fname] = defStr
		}
	}
	TraverseModel(modelType, data, errs, setDefaults)

	// Trim any input strings fields, unless they are off limits (trim=no)
	trimStrings := func(fld reflect.StructField, data do.Map, keys ...string) {
		fname := keys[len(keys)-1]
		if (action == INSERT || action == UPDATE) && data.HasKey(fname) && fld.Tag.Get("trim") != "no" {
			str, isString := data[fname].(string)
			if isString {
				data[fname] = strings.TrimSpace(str)
			}
		}
	}
	TraverseModel(modelType, data, errs, trimStrings)

	// Convert types of fields from STRING to appropriate type
	// as it is specified in the Struct
	if isMongoStore {
		adjustType := func(fld reflect.StructField, data do.Map, keys ...string) {
			fname := keys[len(keys)-1]
			inp, found := data[fname]
			inpStr, isStr := inp.(string)
			expType := fld.Type.String()
			if found && (action == INSERT || action == UPDATE) && isStr && expType != "string" {
				switch expType {
				case "bool":
					data[fname] = inpStr == "1" || inpStr == "yes" || inpStr == "true" || inpStr == "Y" || inpStr == "y"
				case "int":
					i, err := strconv.ParseInt(inpStr, 10, 32)
					if err == nil {
						data[fname] = int(i)
					} else {
						issue := fmt.Sprintf("field '%s' expects 'int' but received %s", fname, inpStr)
						errs.Add(issue, keys...)
					}
				case "int64":
					i, err := strconv.ParseInt(inpStr, 10, 64)
					if err == nil {
						data[fname] = i
					} else {
						issue := fmt.Sprintf("field '%s' expects 'int' but received %s", fname, inpStr)
						errs.Add(issue, keys...)
					}
				default:
					// TODO: panic or error
					issue := fmt.Sprintf("unhandled field '%s' (%s : %s)", fname, expType, inpStr)
					errs.Add(issue, keys...)
				}
			}
		}
		TraverseModel(modelType, data, errs, adjustType)
	}

	// Manage timestamp fields (inserted_at / updated_at)
	// during insert / update of records - do this for only
	// MongoStores
	if isMongoStore && do.TypeComposedOf(modelType, Timed{}) {
		now := time.Now()
		switch action {
		case INSERT:
			data["created_at"] = now
			data["updated_at"] = now
		case UPDATE:
			data["updated_at"] = now
		}
	}

	// Set fields marked auto - to give them a value upon insertion
	setAuto := func(fld reflect.StructField, data do.Map, keys ...string) {
		fname := keys[len(keys)-1]
		if action == INSERT && !data.HasKey(fname) && fld.Tag.Get("auto") != "" {
			auto := parseAutoTag(fld)
			if auto != nil {
				if isMongoStore && fname == "id" && fld.Tag.Get("bson") != "" {
					// give preference to bson
					data[fld.Tag.Get("bson")] = auto.Generate()
				} else {
					data[fname] = auto.Generate()
				}
			}
		}
	}
	TraverseModel(modelType, data, errs, setAuto)

	// Input validations as defined in 'verify' tag
	validateInput := func(fld reflect.StructField, data do.Map, keys ...string) {
		fname := keys[len(keys)-1]
		if (action == INSERT || action == UPDATE) && data.HasKey(fname) {
			checks := getFieldTests(fld)
			for _, check := range checks {
				if success, message := check.Verify(fld.Type, data[fname]); !success {
					errs.Add(message, keys...)
				}
			}
		}
	}
	TraverseModel(modelType, data, errs, validateInput)

	return len(errs) == 0, errs
}

type FieldTest struct {
	Test   string
	Option string
}

func (ft FieldTest) Verify(t reflect.Type, v interface{}) (bool, string) {

	switch t.String() {
	case "string":
		vstr := fmt.Sprintf("%s", v)
		switch ft.Test {
		case "email":
			if govalidator.IsEmail(vstr) {
				return true, ""
			} else {
				return false, fmt.Sprintf("%s is not a valid email", vstr)
			}
		case "rex":
			reg, err := regexp.Compile(ft.Option)
			if err != nil {
				return false, fmt.Sprintf("%s is not a valid regular expression", ft.Option)
			}
			if reg.MatchString(vstr) {
				return true, ""
			} else {
				return false, fmt.Sprintf("%s does not match the regular expression", vstr)
			}
		case "enum":
			if strings.Contains(ft.Option, "|"+vstr+"|") {
				return true, ""
			} else {
				return false, fmt.Sprintf("%s must be one of predefined set", vstr)
			}
		}
	}

	return false, "validation not supported: " + ft.Test
}

func getFieldTests(f reflect.StructField) (fv []FieldTest) {
	fv = []FieldTest{}

	verify := f.Tag.Get("verify")
	if verify == "" {
		return
	}

	tasks := strings.Split(verify, ";")
	for _, task := range tasks {
		f := parseFieldTest(task)
		if f != nil {
			fv = append(fv, *f)
		}
	}
	return
}

func parseFieldTest(input string) *FieldTest {
	fv := FieldTest{}

	i := strings.Index(input, "(")
	if i == -1 {
		fv.Test = input
		return &fv
	}

	j := strings.LastIndex(input, ")")
	fv.Test = input[0:i]
	fv.Option = input[i+1 : j]

	// If enum then set | at beginging and
	// end of options
	if fv.Test == "enum" {
		fv.Option = strings.TrimSpace(fv.Option)
		if !strings.HasPrefix(fv.Option, "|") {
			fv.Option = "|" + fv.Option
		}
		if !strings.HasSuffix(fv.Option, "|") {
			fv.Option = fv.Option + "|"
		}
	}

	return &fv
}

//
// auto:"prefix:p-;uuid"
type AutoField struct {
	Method string // uuid | alphanum (length)?
	Length int    // 0

	Prefix string // optional
}

func (af *AutoField) Generate() string {
	val := af.Prefix
	switch af.Method {
	case "uuid":
		val += uuid.NewString()
	case "alphanum":
		val += RandStringBytesMaskImprSrcSB(af.Length)
	}

	return val
}

func parseAutoTag(f reflect.StructField) *AutoField {

	input := f.Tag.Get("auto")
	if input == "" {
		return nil
	}

	// Split by ;
	parts := strings.Split(input, ";")
	af := AutoField{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "prefix:") {
			af.Prefix = part[7:]
		} else if part == "uuid" {
			af.Method = part
		} else if strings.HasPrefix(part, "alphanum(") {
			af.Method = "alphanum"
			count := part[9 : len(part)-1]
			af.Length = conv.IntOr(count, 16)
		} else {
			// TODO: log unsupported methods "panic"
			return nil
		}
	}

	return &af
}

func TraverseStruct(modelType interface{}, naming FieldName, data do.Map, errs FieldErrors, operation func(reflect.StructField, do.Map, ...string), key ...string) {

	ot := do.TypeOf(modelType)
	ot = do.TypeDereference(ot)

	for i := 0; i < ot.NumField(); i++ {

		sf := ot.Field(i)
		fname := naming.Get(sf)

		keys := key
		if fname != "" {
			keys = append(keys, fname)
		}

		ft := do.TypeDereference(sf.Type)

		if ft.Kind() == reflect.Struct && !do.TypeIsTime(ft) {
			if !data.HasKey(fname) {
				// Pass an empty map
				empty := map[string]interface{}{}
				TraverseStruct(ft, naming, empty, errs, operation, keys...)
				if len(empty) > 0 { // if traverse call modified map passed, then use it in parent
					data[fname] = empty
				}
			} else { // key exists
				goMap, isMap := data.GetOr(fname, false).(map[string]interface{})
				if isMap {
					TraverseStruct(ft, naming, goMap, errs, operation, keys...)
				} else {
					issue := fmt.Sprintf("field '%s' expected dict, but found literal", fname)
					errs.Add(issue, keys...)
				}
			}
		} else {
			operation(sf, data, append(key, fname)...)
		}
	}
}
