package gdax

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	ws "github.com/gorilla/websocket"
)

// EndPoint is the GDAX sandbox endpoint.
const EndPoint = "https://api-public.sandbox.gdax.com"

// An AccessInfo stores credentials.
type AccessInfo struct {
	PublicKey  string `json:"public_api"`
	PrivateKey string `json:"private_api"`
	Passphrase string `json:"passphrase"`
	Client     *http.Client
}

// RetrieveAccessInfoFromEnvironmentVariables retrieves credentials from environment variables.
func RetrieveAccessInfoFromEnvironmentVariables() (*AccessInfo, error) {
	var accessInfo AccessInfo

	accessInfo.PublicKey = os.Getenv("PUBLIC_KEY")
	accessInfo.PrivateKey = os.Getenv("PRIVATE_KEY")
	accessInfo.Passphrase = os.Getenv("PASSPHRASE")
	accessInfo.Client = &http.Client{}
	return &accessInfo, nil
}

// RetrieveAccessInfoFromFile retrieves credentials from a specified file.
func RetrieveAccessInfoFromFile(fileName string) (*AccessInfo, error) {
	var accessInfo AccessInfo

	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fileData, &accessInfo)
	if err != nil {
		return nil, err
	}
	accessInfo.Client = &http.Client{}
	return &accessInfo, nil
}

// collectionRequest is a creates and handles a request and its cursors.
func (accessInfo *AccessInfo) collectionRequest(method, path, jsonBody string) (string, *pagination, error) {
	var errorMessage map[string]string

	req, err := accessInfo.createRequest(method, path, jsonBody)
	if err != nil {
		return "", nil, err
	}

	resp, err := accessInfo.Client.Do(req)
	if err != nil {
		return "", nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	if !(http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices) {
		err = json.Unmarshal(body, &errorMessage)
		if err != nil {
			return "", nil, err
		}
		return "", nil, errors.New(errorMessage["message"])
	}

	cursor := pagination{
		after: resp.Header.Get("CB-AFTER"),
		limit: -1,
	}
	return string(body), &cursor, nil
}

// request creates and handles a request and parses the marshals the json body response into the specified struct.
func (accessInfo *AccessInfo) request(method, path, jsonBody string, v interface{}) (*pagination, error) {
	var errorMessage map[string]string

	req, err := accessInfo.createRequest(method, path, jsonBody)
	if err != nil {
		return nil, err
	}

	log.Println("created req", jsonBody)
	resp, err := accessInfo.Client.Do(req)
	log.Printf("resp: %+v\n", resp)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if !(http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices) {
		err = json.Unmarshal(body, &errorMessage)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errorMessage["message"])
	}

	cursor := pagination{
		after: resp.Header.Get("CB-AFTER"),
		limit: -1,
	}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}
	return &cursor, nil
}

// createRequest builds, creates, and sends an HTTP request.
func (accessInfo *AccessInfo) createRequest(method, requestPath, body string) (*http.Request, error) {
	// https://docs.gdax.com/#signing-a-message

	// get ISO 8601 formatted timestamp.
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	fullRequestPath := EndPoint + requestPath
	// create prehash string
	prehash := timestamp + method + requestPath + body
	privateKeyDecoded, err := base64.StdEncoding.DecodeString(accessInfo.PrivateKey)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, privateKeyDecoded)
	mac.Write([]byte(prehash))
	hashSum := mac.Sum(nil)
	accessSign := base64.StdEncoding.EncodeToString(hashSum)

	req, err := http.NewRequest(method, fullRequestPath, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("CB-ACCESS-KEY", accessInfo.PublicKey)
	req.Header.Set("CB-ACCESS-SIGN", accessSign)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("CB-ACCESS-PASSPHRASE", accessInfo.Passphrase)
	req.Header.Set("Content-Type", "application/json")
	req.Method = method
	url, err := url.Parse(fullRequestPath)
	if err != nil {
		return nil, err
	}
	req.URL = url
	log.Printf("req: %+v\n", req)
	return req, nil
}

// createWebsocketConnection creates a websocket connection.
// This function does not block; this function creates a go routine.
func createWebsocketConnection(addr string, initialMessage []byte, messageType chan string, jsonString chan []byte, errorChan chan error) error {
	var wsDialer ws.Dialer
	conn, _, err := wsDialer.Dial(addr, nil)
	if err != nil {
		return err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(initialMessage, &m); err != nil {
		return err
	}
	if err := conn.WriteJSON(m); err != nil {
		return err
	}
	go func() {
		for {
			var v []byte
			var q map[string]interface{}
			_, v, err := conn.ReadMessage()
			if err != nil {
				errorChan <- err
				break
			}
			if err := json.Unmarshal(v, &q); err != nil {
				errorChan <- err
				break
			}
			errorChan <- nil
			if t, ok := q["type"]; ok {
				z := reflect.ValueOf(t).Convert(reflect.TypeOf(string(v))).Interface().(string)
				messageType <- z
				jsonString <- v
				if z == "error" {
					break
				}
			}
		}
	}()
	return nil
}
