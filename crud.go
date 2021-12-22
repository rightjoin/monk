package monk

type DocAction int

const (
	SELECT = iota
	INSERT
	UPDATE
	DELETE
)

/*
func Insert(m *mongo.Client, model interface{}, opt OptionalBehaviors, inputs ...interface{}) (interface{}, error) {

	// Validate inputs
	if model == nil {
		return nil, errors.New("model cannot be nil")
	}

	if len(inputs) == 0 {
		return nil, errors.New("no values provided to insert")
	}

	inputMap := NewMapFromVals(inputs)
	InvokeHooks(INSERT, model, opt, inputMap)
	RemoveUnwantedBehaviors(opt, inputMap)

	// Do actual insert operations

	return nil, nil
}

func InsertSelect(m *mongo.Client, inputs ...interface{}) (*bson.D, error) {

	//InvokeHooks(InsertDoc, nil)

	//RemoveUnwantedBehaviors(nil)
	SetAutoKeysForInsert(nil)

	return nil, nil
}

func Update(dbo interface{}, idKey string, idVal interface{}, inputs ...interface{}) error {

	//InvokeHooks(UpdateDocument, nil)

	//RemoveUnwantedBehaviors(nil)
	SetAutoKeysForUpdate(nil)
	return nil
}

func UpdateSelect(dbo interface{}, idKey string, idVal interface{}, inputs ...interface{}) interface{} {

	//InvokeHooks(UpdateDocument, nil)

	//RemoveUnwantedBehaviors(nil)
	SetAutoKeysForUpdate(nil)

	return nil
}

func Delete(dbo interface{}, idKey string, idVal interface{}, inputs ...interface{}) error {

	//InvokeHooks(DeleteDocument, nil)

	//RemoveUnwantedBehaviors(nil)
	SetAutoKeysForUpdate(nil)
	return nil
}

func InvokeHooks(action DocAction, model interface{}, opt OptionalBehaviors, input Map) {
	switch action {
	case INSERT:
		globalHook.Timed.BeforeInsert(input)
	case UPDATE:
	case DELETE:
	}
}

func RemoveUnwantedBehaviors(opt OptionalBehaviors, m Map) {

	// TODO:
	// if a collection is not implementing some behaviours then
	// should those be checked here?
}

func SetAutoKeysForInsert(input map[string]interface{}) {
	// created_at

	// updated_at
}

func SetAutoKeysForUpdate(input map[string]interface{}) {
	// updated_at
}

func PrepareData(inputs ...interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, errors.New("not implemented")
}

func isAddress(addr interface{}) bool {
	return false
}

*/
