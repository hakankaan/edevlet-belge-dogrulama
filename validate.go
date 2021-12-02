package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	client http.Client
)

const eDevletURL = "https://turkiye.gov.tr"

type Document struct {
	Barcode string
	ID      string
	Token   string
	Cookies []*http.Cookie
	IsValid bool
}

func init() {
	client = http.Client{
		Timeout: time.Duration(time.Second * 5),
	}

}

// getToken gets the first token from the e-devlet website
func (d *Document) GetToken() error {
	InfoLogger.Println("Getting token...")

	data := url.Values{}

	body, err := d.makeRequest(http.MethodGet, "/belge-dogrulama", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	d.Token = extractToken(body)
	return nil
}

// InsertBarcode inserts the barcode into the barcode form then gets the new token for next form
func (d *Document) InsertBarcode() error {
	InfoLogger.Println("Inserting barcode...")

	data := url.Values{}
	data.Set("sorgulananBarkod", d.Barcode)
	data.Set("token", d.Token)

	body, err := d.makeRequest(http.MethodPost, "/belge-dogrulama?submit", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	d.Token = extractToken(body)
	return nil
}

// InsertID inserts the citizen id into the citizen id form then gets the new token for next form
func (d *Document) InsertID() error {
	InfoLogger.Println("Inserting ID...")
	data := url.Values{}
	data.Set("ikinciAlan", d.ID)
	data.Set("token", d.Token)
	data.Set("btn", "Devam Et")

	body, err := d.makeRequest(http.MethodPost, "/belge-dogrulama?islem=dogrulama&submit", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	d.Token = extractToken(body)
	return nil
}

// AcceptForm accepts the agreements and returns result of validation
func (d *Document) AcceptForm() error {
	InfoLogger.Println("Acceptin form...")
	data := url.Values{}
	data.Set("chkOnay", "1")
	data.Set("token", d.Token)

	body, err := d.makeRequest(http.MethodPost, "/belge-dogrulama?islem=onay&submit", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	d.IsValid = checkForPdfLink(body)
	return nil
}

// makeRequest makes a request to the given url and returns the response body
func (d *Document) makeRequest(method string, path string, body io.Reader) (string, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", eDevletURL, path), body)

	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for _, cookie := range d.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)

	respCookies := resp.Cookies()

	if len(d.Cookies) == 0 && len(respCookies) != 1 {
		d.Cookies = respCookies
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

// extractToken extracts the token from the given html body
func extractToken(body string) string {
	re := regexp.MustCompile(`data-token="\{([^}]*)\}`)
	match := re.FindStringSubmatch(body)
	return fmt.Sprintf("{%s}", match[1])
}

func checkForPdfLink(body string) bool {
	return strings.Contains(body, "/belge-dogrulama?belge=goster&goster=1")
}
