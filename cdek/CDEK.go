package cdek

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CDEK struct {
	account        string
	securePassword string
	apiURL         string
}

// NewCDEK creates a new client.
func NewCDEK(account, securePassword, apiURL string) *CDEK {
	return &CDEK{
		account:        account,
		securePassword: securePassword,
		apiURL:         apiURL,
	}
}

func (c *CDEK) Calculate(addrFrom string, addrTo string, size Size) ([]PriceSending, error) {
	if c.apiURL == "" {
		return nil, errors.New("API URL is empty")
	}

	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	reqURL := fmt.Sprintf("%s/calculator/tarifflist", c.apiURL)

	_, err = json.Marshal(size)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal size: %w", err)
	}

	data := map[string]interface{}{
		"type":     1,
		"date":     "2020-11-03T11:49:32+0700",
		"currency": 1,
		"lang":     "rus",
		"from_location": map[string]string{
			"code": addrFrom,
		},
		"to_location": map[string]string{
			"code": addrTo,
		},
		"packages": map[string]int{
			"height": size.Height,
			"length": size.Length,
			"weight": size.Weight,
			"width":  size.Width,
		},
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(dataJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get response. HTTP status code: %d", resp.StatusCode)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	jsonData, err := io.ReadAll(resp.Body)

	var array TariffCodes

	err = json.Unmarshal(jsonData, &array)
	if err != nil {
		return nil, err
	}

	return array.TariffCodes, nil
}

func (c *CDEK) getToken() (string, error) {
	dt := url.Values{}
	dt.Set("client_id", c.account)
	dt.Set("client_secret", c.securePassword)
	u, _ := url.ParseRequestURI(c.apiURL + "/oauth/token")

	dt.Set("grant_type", "client_credentials")
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(dt.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(r)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var result struct {
		Token     string `json:"access_token"`
		TokenType string `json:"token_type"`
		ExpiresIn int    `json:"expires_in"`
		Scope     string `json:"scope"`
		JTI       string `json:"jti"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return result.Token, nil
}

func (ps *PriceSending) Print() {
	fmt.Println()
	fmt.Println()
	fmt.Println("--------------------------")
	fmt.Printf("Tariff Code: %d\n", ps.TariffCode)
	fmt.Printf("Tariff Name: %s\n", ps.TariffName)
	fmt.Printf("Tariff Description: %s\n", ps.TariffDescription)
	fmt.Printf("Delivery Mode: %d\n", ps.DeliveryMode)
	fmt.Printf("Delivery Sum: %f\n", ps.DeliverySum)
	fmt.Printf("Period Min: %d\n", ps.PeriodMin)
	fmt.Printf("Period Max: %d\n", ps.PeriodMax)
	fmt.Printf("Calendar Min: %d\n", ps.CalendarMin)
	fmt.Printf("Calendar Max: %d\n", ps.CalendarMax)
	fmt.Println("--------------------------")
	fmt.Println()
	fmt.Println()
}

type Size struct {
	Height int `json:"height"`
	Length int `json:"length"`
	Weight int `json:"weight"`
	Width  int `json:"width"`
}

type TariffCodes struct {
	TariffCodes []PriceSending `json:"tariff_codes"`
}

type PriceSending struct {
	TariffCode        int     `json:"tariff_code"`
	TariffName        string  `json:"tariff_name"`
	TariffDescription string  `json:"tariff_description"`
	DeliveryMode      int     `json:"delivery_mode"`
	DeliverySum       float64 `json:"delivery_sum"`
	PeriodMin         int     `json:"period_min"`
	PeriodMax         int     `json:"period_max"`
	CalendarMin       int     `json:"calendar_min"`
	CalendarMax       int     `json:"calendar_max"`
}
