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
)

const (
	contentTypeHeader = "Content-Type"
	dateHeader = "Date"
)

type RequestSigner func(http.Request) (http.Request, error) 

func SignBasicAuth(req http.Request, username string, pass string) http.Request {
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username + ":" + pass))
	req.Header.Add("Authorization", basicAuth)
	return req
}

func get(url string, sign RequestSigner) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	
	if err != nil {
		return nil, fmt.Errorf("Failed to construct HTTP GET for URL: %s\nCaused by: %s", url, err.Error())
	}

	req.Header.Add(contentTypeHeader, "application/xml")
	req.Header.Add(dateHeader, time.Now().Format(http.TimeFormat))

	newReq, err := sign(*req)
	
	if err != nil {
		return nil, fmt.Errorf("Failed to sign HTTP GET request to URL: %s\nCaused by: %s", url, err.Error())
	}
	
	client := &http.Client{}
	resp, err := client.Do(&newReq)
	
	if err != nil {
		return nil, fmt.Errorf("Failed HTTP GET request to URL: %s\nCaused by: %s", url, err.Error())
	}
	
	return resp.Body, nil
}

func unmarshal(data []byte, resource interface{}) error {	
	err := xml.Unmarshal(data, resource)
	
	if err != nil {
		myErr := fmt.Errorf("%T\n%s\n%#v\n", err, err, err)
		return myErr
	}
	
	return nil
}

func getAndUnmarshal(url string, sign RequestSigner, resource interface{}) (error) {
	body, err := get(url, sign)
	
	if err != nil {
		return err
	}
	
	defer body.Close()
	data, err := ioutil.ReadAll(body)
	
	return unmarshal(data, &resource)
}

func GetCard(cardNumber int, baseURL string, sign RequestSigner) (Card, error){
	var card Card
	
	url := fmt.Sprintf("%s/cards/%d.xml", baseURL, cardNumber)
	
	err := getAndUnmarshal(url, sign, interface{}(&card))

	return card, err
}

func Query(query string, baseURL string, sign RequestSigner) ([]map[string]string, error) {
	var results []map[string]string
	
	url := fmt.Sprintf("%s/cards/execute_mql.xml?mql=%s", baseURL, url.QueryEscape(query))
	
	body, err := get(url, sign)

	parser := xml.NewDecoder(body)
	
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

func createCard(card Card, baseURL string, sign RequestSigner) (int, error) {

}

*/