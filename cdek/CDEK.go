package cdek

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type API struct {
	account        string
	securePassword string
	apiURL         string
}

type Client struct {
	Token    string
	Method   string
	Endpoint string
}

type Phone struct {
	Number string `json:"number"`
}
type Recipient struct {
	Name   string  `json:"name"`
	Phones []Phone `json:"phones"`
}

type Payment struct {
	Value  float64 `json:"value"`
	VatSum float64 `json:"vat_sum,omitempty"`
}

type Request struct {
	RequestUUID string `json:"request_uuid"`
	Type        string `json:"type"`
	DateTime    string `json:"date_time"`
	State       string `json:"state"`
	Errors      []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

type Item struct {
	Name           string  `json:"name,omitempty"`
	WareKey        string  `json:"ware_key"`
	Payment        Payment `json:"payment"`
	Weight         int     `json:"weight"`
	WeightGross    int     `json:"weight_gross"`
	Amount         int     `json:"amount"`
	DeliveryAmount int     `json:"delivery_amount"`
	NameI18N       string  `json:"name_i18n"`
	Url            string  `json:"url"`
	Cost           float64 `json:"cost"`
}

type Package struct {
	Height int    `json:"height,omitempty"`
	Length int    `json:"length,omitempty"`
	Weight int    `json:"weight,omitempty"`
	Width  int    `json:"width,omitempty"`
	Number string `json:"number,omitempty"`
	Items  []Item `json:"items"`
}

// NewCDEK creates a new client.
func NewCDEK(account, securePassword, apiURL string) *API {
	return &API{
		account:        account,
		securePassword: securePassword,
		apiURL:         apiURL,
	}
}

func (c *API) Calculate(addrFrom string, addrTo string, size Size) ([]PriceSending, error) {
	if c.apiURL == "" {
		return nil, errors.New("API URL is empty")
	}

	token, _, err := c.getToken()

	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	endpoint := "calculator/tarifflist"

	_, err = json.Marshal(size)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal size: %w", err)
	}

	data := Delivery{
		Type:         1,
		Date:         "2020-11-03T11:49:32+0700",
		Currency:     1,
		Lang:         "rus",
		FromLocation: Location{Address: addrFrom},
		ToLocation:   Location{Address: addrTo},
		Packages:     size,
	}

	dataJSON, err := json.Marshal(data)
	client := Client{Token: token, Endpoint: endpoint, Method: "POST"}
	body, err := SendRequest(c.apiURL, &client, dataJSON)

	if err != nil {
		return nil, err
	}

	jsonData, err := io.ReadAll(body)

	if err != nil {
		return nil, err
	}

	var array TariffCodes

	err = json.Unmarshal(jsonData, &array)
	if err != nil {
		return nil, err
	}

	return array.TariffCodes, nil
}

func (c *API) ValidateAddress(add string) (bool, string, error) {

	if add == "" {
		return false, "", fmt.Errorf("empty address")
	}

	endpoint := "deliverypoints"
	token, _, err := c.getToken()

	if err != nil {
		return false, "", fmt.Errorf("failed to get token")
	}

	if token == "" {
		return false, "", fmt.Errorf("Token")
	}

	client := Client{Token: token, Endpoint: endpoint, Method: http.MethodGet}

	body, err := SendRequest(c.apiURL, &client, nil)

	if err != nil {
		return false, "", fmt.Errorf("failed to send request ", err.Error())
	}

	jsonData, err := io.ReadAll(body)

	type Region struct {
		AddressComment string `json:"address_comment"`
		Name           string `json:"name"`
		Email          string `json:"email"`
		Location       struct {
			CountryCode string  `json:"country_code"`
			RegionCode  int     `json:"region_code"`
			Region      string  `json:"region"`
			CityCode    int     `json:"city_code"`
			City        string  `json:"city"`
			PostalCode  string  `json:"postal_code"`
			Longitude   float64 `json:"longitude"`
			Latitude    float64 `json:"latitude"`
			Address     string  `json:"address"`
			AddressFull string  `json:"address_full"`
		} `json:"location"`
	}
	var regions []Region
	err = json.Unmarshal(jsonData, &regions)

	if err != nil {
		return false, "", err
	}
	for _, region := range regions {
		if strings.Contains(region.Location.Address, add) ||
			strings.Contains(region.Location.AddressFull, add) ||
			strings.Contains(region.Name, add) {
			return true, region.Location.AddressFull, nil
		}
	}

	return false, add, fmt.Errorf("address is not verified")
}

func (c *API) CreateOrder(from, to string, size Size, typeSending int) (string, error) {
	endpoint := "orders"

	token, _, err := c.getToken()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	client := Client{Token: token, Endpoint: endpoint, Method: http.MethodPost}
	data := struct {
		FromLocation Location  `json:"from_location"`
		ToLocation   Location  `json:"to_location"`
		Recipient    Recipient `json:"recipient"`
		Packages     []Package `json:"packages"`
		TariffCode   int       `json:"tariff_code"`
	}{
		FromLocation: Location{Address: from},
		ToLocation:   Location{Address: to},
		Recipient: Recipient{
			Name: "Семенов Семен",
			Phones: []Phone{
				{Number: "78888888888"},
			},
		},
		Packages: []Package{
			{
				Height: size.Height,
				Width:  size.Width,
				Length: size.Length,
				Weight: size.Weight,
				Number: "bar-001",
				Items: []Item{
					{
						Name:           "Товар",
						WareKey:        "00055",
						Weight:         700,
						Amount:         2,
						Cost:           300,
						Payment:        Payment{Value: 6000.0, VatSum: 0.0},
						Url:            "www.item.ru",
						NameI18N:       "",
						DeliveryAmount: 2,
						WeightGross:    0,
					},
				},
			},
		},
		TariffCode: typeSending,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to Marshal data: %w", err)
	}

	body, err := SendRequest(c.apiURL, &client, dataJSON)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer body.Close()

	jsonData, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	var order struct {
		Entity struct {
			UUID string `json:"uuid,omitempty"`
		} `json:"entity,omitempty"`
		Requests []Request `json:"requests"`
	}
	err = json.Unmarshal(jsonData, &order)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return order.Entity.UUID, nil
}

func (c *API) GetStatus(orderID string) (string, error) {
	// Construct the endpoint for the API call.
	endpoint := "orders/" + orderID

	// Get the token needed for the API call.
	token, _, err := c.getToken()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	// Create the client object with the necessary information.
	client := Client{Token: token, Endpoint: endpoint, Method: http.MethodGet}

	// Define the struct types used in this function.
	type Status struct {
		Code     string `json:"code"`
		Name     string `json:"name"`
		DateTime string `json:"date_time"`
		City     string `json:"city"`
	}

	type entity struct {
		UUID       string    `json:"uuid"`
		TariffCode int       `json:"tariff_code"`
		Recipient  Recipient `json:"recipient"`
		From       Location  `json:"from_location"`
		To         Location  `json:"to_location"`
		Packages   []Package `json:"packages"`
		Statuses   []Status  `json:"statuses"`
	}

	type jsonResponse struct {
		Entity   entity    `json:"entity"`
		Requests []Request `json:"requests"`
	}

	// Make the API call and read the response.
	body, err := SendRequest(c.apiURL, &client, nil)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	jsonData, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response.
	var response jsonResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Extract the status code from the JSON response.

	statuses := response.Entity.Statuses
	if len(statuses) == 0 {
		return "", fmt.Errorf("no status information found")
	}
	return statuses[0].Code, nil
}

func (c *API) getToken() (string, int, error) {
	data := url.Values{
		"client_id":     {c.account},
		"client_secret": {c.securePassword},
		"grant_type":    {"client_credentials"},
	}

	req, err := http.NewRequest(http.MethodPost, c.apiURL+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var token struct {
		Token     string `json:"access_token"`
		TokenType string `json:"token_type"`
		ExpiresIn int    `json:"expires_in"`
		Scope     string `json:"scope"`
		JTI       string `json:"jti"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return token.Token, token.ExpiresIn, nil
}

func SendRequest(url string, client *Client, body []byte) (io.ReadCloser, error) {
	// Check if the input parameters are valid
	if client == nil {
		return nil, fmt.Errorf("client is empty")
	}
	if client.Endpoint == "" {
		return nil, errors.New("API URL is empty")
	}

	// Build the full URL
	fullURL := url + client.Endpoint

	// Create a new HTTP request with the specified method, URL and request body
	req, err := http.NewRequestWithContext(context.Background(), client.Method, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.Token)

	// Send the request using the default HTTP client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	// Check the HTTP response status code
	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		// The response status code is valid, return the response body
		return resp.Body, nil
	default:
		// The response status code is invalid, return an error
		defer resp.Body.Close()
		return nil, fmt.Errorf("failed to get response. HTTP status code: %d", resp.StatusCode)
	}
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

type Location struct {
	Code        string `json:"code"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
	City        string `json:"city"`
	Address     string `json:"address"`
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

type Delivery struct {
	Type         int      `json:"type"`
	Date         string   `json:"date"`
	Currency     int      `json:"currency"`
	Lang         string   `json:"lang"`
	FromLocation Location `json:"from_location"`
	ToLocation   Location `json:"to_location"`
	Packages     Size     `json:"packages"`
}
