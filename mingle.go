package mingle

import (
	"net/http"
	"net/url"
	"io/ioutil"
	"io"
	"encoding/base64"
	"time"
	"fmt"
	"encoding/xml"
	"bytes"
	"log"
)

const (
	contentTypeHeader = "Content-Type"
	dateHeader = "Date"
	acceptHeader = "Accept"
)

type RequestSigner func(http.Request) (http.Request, error) 

func SignBasicAuth(req http.Request, username string, pass string) http.Request {
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username + ":" + pass))
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

func doAndUnmarshal(method string, url string, sign RequestSigner, resource interface{}, body io.Reader) (error) {
	response, err := doRequest(method, url, sign, body)
	
	if err != nil {
		return err
	}
	
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	
	return unmarshal(data, &resource)
}

func GetCard(cardNumber int, baseURL string, sign RequestSigner) (Card, error){
	var card Card
	
	url := fmt.Sprintf("%s/cards/%d.xml", baseURL, cardNumber)
	
	err := doAndUnmarshal("GET", url, sign, interface{}(&card), nil)

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
	
	_, err = doRequest("PUT", url, sign, body)
		
	return err
}

func CreateCard(card Card, baseURL string, sign RequestSigner) (int, error) {
	var cardNumber int
	
	url := fmt.Sprintf("%s/cards", baseURL)
	
	data, err := xml.Marshal(card)
	
	if err != nil {
		myErr := fmt.Errorf("%T\n%s\n%#v\n", err, err, err)
		return cardNumber, myErr
	}
		
	body := bytes.NewBuffer(data)
	
	response, err := doRequest("POST", url, sign, body)
	
	// TODO parse cardNumber (from Location header in response?)
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	
	log.Printf("Response: %d %s\nLocation: %s\n%s", response.StatusCode, response.Status, response.Header.Get("Location"), string(responseBody))
	
	return cardNumber, err
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