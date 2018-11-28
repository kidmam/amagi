package jsonCache

import (
	"encoding/json"
	"fmt"
	"linkermodels"
	"os"
	"testing"

	"github.com/b-eee/amagi/services/database"
)

func init() {
	database.StartRedis()

	os.Setenv("ENABLE_REDIS_CACHE", "1")
	fmt.Println("starting test CacheGetEx")
}

type (
	// ApplicationDatastore application datastore struct
	ApplicationDatastore struct {
		Application struct {
			ApplicationID string                   `json:"application_id"`
			DisplayID     string                   `json:"display_id"`
			DisplayOrder  int                      `json:"display_order"`
			Datastores    []linkermodels.Datastore `json:"datastores"`
		} `json:"application"`
	}
)

func TestCacheGetEx(t *testing.T) {
	keys := []string{"applications", "list", "5bdabd6d8726e333886baaca"}
	var res interface{}
	rep, err := CacheGetEx(keys, &res, 12000)
	if err != nil {
		t.Error(err)
	}

	// r := []struct {
	// 	Application linkermodels.Application `json:"application"`
	// }{}

	var r []linkermodels.Application
	// var r []ApplicationDatastore
	// var r interface{}
	fmt.Println(json.Unmarshal([]byte(rep), &r))

	fmt.Println(r)
}
