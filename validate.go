package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
	Cookies []*http.Cookie
}

func init() {
	client = http.Client{
		Timeout: time.Duration(time.Second * 5),
	}

}

// getToken gets the first token from the e-devlet website
func (d document) GetToken() error {
	InfoLogger.Println("Getting token...")
	body, err := d.makeRequest(http.MethodGet, "/belge-dogrulama", nil)
	if err != nil {
		return err
	}
	d.Token = extractToken(body)
	InfoLogger.Println("Token:", d.Token)
	return nil
}

// InsertBarcode inserts the barcode into the barcode form then gets the new token for next form
func (d document) InsertBarcode() error {
	InfoLogger.Println("Inserting barcode...")
	requestBody, err := json.Marshal(map[string]string{
		"sorgulananBarkod": d.Barcode,
		"token":            d.Token,
		"btn":              "Devam Et",
	})
	if err != nil {
		return err
	}
	body, err := d.makeRequest(http.MethodPost, "/belge-dogrulama?submit", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	d.Token = extractToken(body)
	InfoLogger.Println("Token:", d.Token)
	// InfoLogger.Println(body)
	return nil
}

// InsertID inserts the citizen id into the citizen id form then gets the new token for next form
func (d document) InsertID() error {
	InfoLogger.Println("Inserting ID...")
	requestBody, err := json.Marshal(map[string]string{
		"ikinciAlan": d.ID,
		"token":      d.Token,
		"btn":        "Devam Et",
	})
	if err != nil {
		return err
	}

	body, err := d.makeRequest(http.MethodPost, "/belge-dogrulama?islem=dogrulama&submit", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	d.Token = extractToken(body)
	InfoLogger.Println("Token:", d.Token)
	return nil
}

// AcceptForm accepts the agreements and returns result of validation
func (d document) AcceptForm() error {
	InfoLogger.Println("Acceptin form...")
	requestBody, err := json.Marshal(map[string]interface{}{
		"chkOnay": 1,
		"token":   d.Token,
		"btn":     "Devam Et",
	})
	if err != nil {
		return err
	}

	_, requestErr := d.makeRequest(http.MethodPost, "/belge-dogrulama?islem=onay&submit", bytes.NewBuffer(requestBody))
	if requestErr != nil {
		return requestErr
	}
	// InfoLogger.Println(body)
	return nil
}

// extractToken extracts the token from the given html body
func extractToken(body string) string {
	re := regexp.MustCompile(`data-token="\{([^}]*)\}`)
	match := re.FindStringSubmatch(body)
	return fmt.Sprintf("{%s}", match[1])
}

// makeRequest makes a request to the given url and returns the response body
func (d document) makeRequest(method string, path string, body io.Reader) (string, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", eDevletURL, path), body)

	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36")

	for _, cookie := range d.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)

	d.Cookies = resp.Cookies()

	for _, cookie := range req.Cookies() {
		InfoLogger.Println("Cookie:", cookie)
	}

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
