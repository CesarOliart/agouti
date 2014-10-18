package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Session struct {
	URL string
}

func (s *Session) Execute(endpoint, method string, body interface{}, result ...interface{}) error {
	client := &http.Client{}

	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("invalid request body: %s", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}

	request, err := http.NewRequest(method, strings.TrimSuffix(s.URL+"/"+endpoint, "/"), bodyReader)
	if err != nil {
		return fmt.Errorf("invalid request: %s", err)
	}

	if method == "POST" {
		request.Header.Add("Content-Type", "application/json")
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request failed: %s", err)
	}

	responseBody, _ := ioutil.ReadAll(response.Body)

	if response.StatusCode < 200 || response.StatusCode > 299 {
		var errBody struct{ Value struct{ Message string } }
		if err := json.Unmarshal(responseBody, &errBody); err != nil {
			return fmt.Errorf("request unsuccessful: error unreadable")
		}

		var errMessage struct{ ErrorMessage string }
		if err := json.Unmarshal([]byte(errBody.Value.Message), &errMessage); err != nil {
			return fmt.Errorf("request unsuccessful: error message unreadable")
		}

		return fmt.Errorf("request unsuccessful: %s", errMessage.ErrorMessage)
	}

	if len(result) > 0 {
		bodyValue := struct{ Value interface{} }{result[0]}

		if err := json.Unmarshal(responseBody, &bodyValue); err != nil {
			return fmt.Errorf("failed to parse response value: %s", err)
		}
	}

	return nil
}
