package mingle

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	contentTypeHeader = "Content-Type"
	dateHeader        = "Date"
	acceptHeader      = "Accept"
)

type RequestSigner func(http.Request) (http.Request, error)

func SignBasicAuth(req http.Request, username string, pass string) http.Request {
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+pass))
	req.Header.Add("Authorization", basicAuth)
	return req
}

func doRequest(method string, url string, sign RequestSigner, body io.Reader) (*http.Response /*io.ReadCloser*/, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, fmt.Errorf("Failed to construct HTTP %s for URL: %s\nCaused by: %s", method, url, err.Error())
	}

	req.Header.Add(contentTypeHeader, "application/xml")
	req.Header.Add(acceptHeader, "application/xml")
	req.Header.Add(dateHeader, time.Now().Format(http.TimeFormat))

	newReq, err := sign(*req)

	if err != nil {
		return nil, fmt.Errorf("Failed to sign HTTP %s request to URL: %s\nCaused by: %s", method, url, err.Error())
	}

	client := &http.Client{}
	resp, err := client.Do(&newReq)

	if err != nil {
		return nil, fmt.Errorf("Failed HTTP %s request to URL: %s\nCaused by: %s", method, url, err.Error())
	}

	return resp, nil
}

func unmarshal(data []byte, resource interface{}) error {
	err := xml.Unmarshal(data, resource)

	if err != nil {
		myErr := fmt.Errorf("%T\n%s\n%#v\n", err, err, err)
		return myErr
	}

	return nil
}

func getCard(url string, sign RequestSigner) (*Card, error) {
	var card Card

	response, err := doRequest("GET", url, sign, nil)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, nil
		} else {
			return nil, fmt.Errorf("Failed executing request '%s' with HTTP response %s", url, response.Status)
		}
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)

	err = unmarshal(data, interface{}(&card))

	return &card, err
}

// GetCard with cardNumber from Mingle or nil if the specified card does not exist.
func GetCard(cardNumber int, baseURL string, sign RequestSigner) (*Card, error) {
	url := fmt.Sprintf("%s/cards/%d.xml", baseURL, cardNumber)

	card, err := getCard(url, sign)

	return card, err
}

func UpdateCard(card Card, baseURL string, sign RequestSigner) error {
	url := fmt.Sprintf("%s/cards/%d.xml", baseURL, card.Number)

	data, err := xml.Marshal(card)

	if err != nil {
		myErr := fmt.Errorf("%T\n%s\n%#v\n", err, err, err)
		return myErr
	}

	body := bytes.NewBuffer(data)

	response, err := doRequest("PUT", url, sign, body)

	log.Printf("Response code: %s", response.StatusCode)

	return err
}

func CreateCard(card Card, baseURL string, sign RequestSigner) (*Card, error) {
	url := fmt.Sprintf("%s/cards.xml", baseURL)

	data, err := xml.Marshal(card)

	if err != nil {
		myErr := fmt.Errorf("%T\n%s\n%#v\n", err, err, err)
		return nil, myErr
	}

	body := bytes.NewBuffer(data)

	response, err := doRequest("POST", url, sign, body)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 201 {
		return nil, fmt.Errorf("Unable to create card - Mingle returned a HTTP response of %s", response.Status)
	}

	result, err := getCard(response.Header.Get("Location"), sign)

	return result, err
}

func Query(query string, baseURL string, sign RequestSigner) ([]map[string]string, error) {
	var results []map[string]string

	url := fmt.Sprintf("%s/cards/execute_mql.xml?mql=%s", baseURL, url.QueryEscape(query))

	response, err := doRequest("GET", url, sign, nil)

	defer response.Body.Close()
	parser := xml.NewDecoder(response.Body)

	var currentRecord map[string]string
	for {
		token, _ := parser.Token()

		if token == nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "results": // do nothing
			case "result":
				currentRecord = make(map[string]string)
				results = append(results, currentRecord)
			default:
				var field string
				parser.DecodeElement(&field, &t)
				currentRecord[t.Name.Local] = field
			}
		}
	}

	return results, err
}

/*

func AddComment(comment string, cardNumber int, baseURL string, sign RequestSigner) error {
	return nil
}

*/
