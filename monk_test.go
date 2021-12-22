package monk

import (
	"fmt"
	"os"
	"testing"
)

//var testCollection = ""

var testConnection = MongoConn{
	ConnStr:    "mongodb://admin:admin@127.0.0.1:27017/",
	DB:         "db-testing",
	Coll:       "coll-testing",
	CollSuffix: NewUUID(16),
}

func TestMain(m *testing.M) {

	func() { // SETUP
		fmt.Println("test collection:", testConnection.CollSuffix)
	}()

	retCode := m.Run()

	func() { // TEARDOWN

		// Delete the "testing" database (if exists)
		{
			ctx, cancel := GetContext()
			defer cancel()
			testConnection.Database().Drop(ctx)
		}
	}()

	os.Exit(retCode)
}
