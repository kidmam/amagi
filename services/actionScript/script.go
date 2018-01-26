package actionScript

import (
	"fmt"
	"os"
	"regexp"

	"github.com/b-eee/amagi/services/externalSvc"

	utils "github.com/b-eee/amagi"
)

type (
	// Script script interface
	Script struct {
		Script string      `json:"script"`
		Data   interface{} `json:"data"`
	}
)

// TryScript try script
func (s *Script) TryScript() error {
	req := map[string]interface{}{
		"script": s.Script,
		"data":   s.Data,
	}

	fmt.Println(req["data"])
	var resp map[string]interface{}
	if err := externalSvc.GenericHTTPRequesterWResp("POST", "http", Host(), "/try", req, &resp); err != nil {
		utils.Error(fmt.Sprintf("error TryScript %v", err))

		return err
	}

	if resp["status"].(float64) != 200 {
		return fmt.Errorf("an error occurred")
	}

	return nil
}

// RunScriptOnUpdate run script on item update
func (s *Script) RunScriptOnUpdate() error {
	req := map[string]interface{}{
		"script": s.Script,
		"data":   s.Data,
	}

	var resp map[string]interface{}
	if err := externalSvc.GenericHTTPRequesterWResp("POST", "http", Host(), "/run", req, &resp); err != nil {
		utils.Error(fmt.Sprintf("error TryScript %v", err))

		// skip error handler as actionScriptHost may not exists
		// TODO -JP
		return nil
	}

	if resp["status"].(float64) != 200 {
		return fmt.Errorf("an error occurred")
	}

	return nil
}

// ReplaceEnvVars replace env variables values
func (s *Script) ReplaceEnvVars(envVars map[string]string) error {

	for k, v := range envVars {
		re := regexp.MustCompile(k)
		s.Script = re.ReplaceAllString(s.Script, v)
	}

	return nil
}

// getUserAPIToken get user api token from sql
func getUserAPIToken() error {

	return nil
}

// Host return action script host address
func Host() string {
	host := "localhost:3000"
	switch os.Getenv("ENV") {
	case "dev", "stg", "prod":
		host = "beee-actionscript:3000"
	}

	return host
}

// LinkerAPIHost linker api hostname or url
func LinkerAPIHost() string {
	host := "localhost:7575"
	switch os.Getenv("ENV") {
	case "dev", "stg", "prod":
		host = "beee-actionscript:7575"
	}

	return host
}
