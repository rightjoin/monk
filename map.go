package monk

// type Map map[string]interface{}

// func NewMap() Map {
// 	return Map{}
// }

// func NewMapFromBsonD(d bson.D) Map {
// 	return NewMapFromGoMap(d.Map())
// }

// func NewMapFromGoMap(goMap map[string]interface{}) Map {
// 	m := Map(goMap)
// 	return m
// }

// func NewMapFromVals(kv ...interface{}) Map {
// 	if len(kv) == 1 {
// 		if kvMap, ok := kv[0].(map[string]interface{}); ok {
// 			return NewMapFromGoMap(kvMap)
// 		} else if kvDoc, ok := kv[0].(bson.D); ok {
// 			return NewMapFromBsonD(kvDoc)
// 		} else {
// 			//
// 			panic("only expected map or bson")
// 		}
// 	} else {
// 		// Loop it through and create a map
// 		tmp := map[string]interface{}{}
// 		for i := 0; i+1 < len(kv); i = i + 2 {
// 			tmp[fmt.Sprint(kv[i])] = kv[i+1]
// 		}
// 		return NewMapFromGoMap(tmp)
// 	}
// }

// func (m Map) HasKey(key string) bool {
// 	goMap := map[string]interface{}(m)
// 	_, ok := goMap[key]
// 	return ok
// }

// func (m Map) Get(key string) (interface{}, bool) {
// 	goMap := map[string]interface{}(m)
// 	val, ok := goMap[key]
// 	if ok {
// 		return val, true
// 	} else {
// 		return nil, false
// 	}
// }

// func (m Map) GetOr(key string, defValue interface{}) interface{} {
// 	goMap := map[string]interface{}(m)
// 	val, ok := goMap[key]
// 	if ok {
// 		return val
// 	} else {
// 		return defValue
// 	}
// }
