package health

import (
	"encoding/xml"
	"fmt"
	"github.com/delucks/go-subsonic"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
)

// этот метод скопироан из внутренностей go-subsonic, так как там нет возможности закрыть соединение
// TODO: remove if https://github.com/delucks/go-subsonic/pull/8 will be merged
func getStream(s *subsonic.Client, id string, parameters map[string]string) (io.ReadCloser, error) {
	params := url.Values{}
	params.Add("id", id)
	for k, v := range parameters {
		params.Add(k, v)
	}
	response, err := s.Request("GET", "stream", params)
	if err != nil {
		return nil, err
	}
	contentType := response.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/xml") || strings.HasPrefix(contentType, "application/xml") {
		// An error was returned
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		resp := subsonic.Response{}
		err = xml.Unmarshal(responseBody, &resp)
		if err != nil {
			return nil, err
		}
		if resp.Error != nil {
			err = fmt.Errorf("Error #%d: %s\n", resp.Error.Code, resp.Error.Message)
		} else {
			err = fmt.Errorf("An error occurred: %#v\n", resp)
		}
		return nil, err
	}
	return response.Body, nil
}
