package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	client http.Client
)

const eDevletURL = "https://turkiye.gov.tr"

type document struct {
	Barcode string
	ID      string
	Token   string
}

func init() {
	client = http.Client{
		Timeout: time.Duration(time.Second * 5),
	}

}

func (d document) InsertBarcode() {
	requestBody, err := json.Marshal(map[string]string{
		"sorgulananBarkod": d.Barcode,
		"token":            d.Token,
	})
	if err != nil {
		ErrorLogger.Println(err.Error())
		os.Exit(1)
	}

	body, err := makeRequest(http.MethodPost, "/belge-dogrulama?submit", bytes.NewBuffer(requestBody))
	if err != nil {
		ErrorLogger.Println(err.Error())
		os.Exit(1)
	}
	d.Token = extractToken(body)
}

func (d document) InsertID() {
	requestBody, err := json.Marshal(map[string]string{
		"ikinciAlan": d.ID,
		"token":      d.Token,
	})
	if err != nil {
		ErrorLogger.Println(err.Error())
		os.Exit(1)
	}

	body, err := makeRequest(http.MethodPost, "/belge-dogrulama?islem=dogrulama&submit", bytes.NewBuffer(requestBody))
	if err != nil {
		ErrorLogger.Println(err.Error())
		os.Exit(1)
	}
	d.Token = extractToken(body)
}

// getToken gets the first token from the e-devlet website
func (d document) GetToken() {
	body, err := makeRequest(http.MethodGet, "/belge-dogrulama", nil)
	if err != nil {
		ErrorLogger.Println(err.Error())
		os.Exit(1)
	}
	token := extractToken(body)
	d.Token = token
}

// extractToken extracts the token from the given html body
func extractToken(body string) string {
	re := regexp.MustCompile(`data-token="\{([^}]*)\}`)
	match := re.FindStringSubmatch(body)
	return fmt.Sprintf("{%s}", match[1])
}

// makeRequest makes a request to the given url and returns the response body
func makeRequest(method string, path string, body io.Reader) (string, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", eDevletURL, path), body)

	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36")

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
