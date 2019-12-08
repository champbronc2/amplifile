package bottlepay

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// Set Bottlepay secret
const (
	BottlePayTokenURL       = "https://bottle.dev/oauth/token"
	BottlePayUserURL        = "https://bottle.dev/api/user"
	BottlePayUserInvoiceURL = "https://bottle.dev/api/wallet/invoice"
	BottlePayUserBalanceURL = "https://wallet.bottle.dev/rpc/v1/balance/blockchain"
	ClientID                = "18"
	ClientSecret            = ""
	WebhookURL              = "https://amplifile.net/webhooks"
)

type AuthResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type BalanceResponse struct {
	TotalBalance       int `json:"total_balance"`
	ConfirmedBalance   int `json:"confirmed_balance"`
	UnconfirmedBalance int `json:"unconfirmed_balance"`
}

type InvoicePayload struct {
	Value    int       `json:"value"`
	Memo     string    `json:"memo"`
	AddIndex int       `json:"add_index"`
	Webhooks []Webhook `json:"webhooks"`
}

type Webhook struct {
	URL       string `json:"url"`
	Token     string `json:"token"`
	EnableSSL bool   `json:"enable_ssl_verification"`
}

// TO-DO: fully populate
type WebhookRequest struct {
	RHash string `json:"r_hash"`
}

type InvoiceResponse struct {
	RHash          string   `json:"r_hash"`
	PaymentRequest string   `json:"payment_request"`
	AddIndex       int      `json:"add_index"`
	Webhooks       []string `json:"webhooks"`
}

type UserResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Currency   string `json:"currency"`
	Locale     string `json:"locale"`
	IsBot      string `json:"is_bot"`
	IsWelcomed string `json:"is_welcomed"`
}

func FetchAccessToken(code string) (*AuthResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	payload := `{
		"grant_type":"authorization_code",
		"client_id":"` + ClientID + `",
		"client_secret":"` + ClientSecret + `",
		"code":"` + code + `",
		"redirect_uri":"https://amplifile.net/oauth/redirect"
	}`
	var jsonStr = []byte(payload)

	req, err := http.NewRequest("POST", BottlePayTokenURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	authResponse := &AuthResponse{}
	json.Unmarshal(bodyBytes, &authResponse)

	return authResponse, nil
}

func FetchUser(token string) (*UserResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", BottlePayUserURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	userResponse := &UserResponse{}
	json.Unmarshal(bodyBytes, &userResponse)

	return userResponse, nil
}

func FetchUserInvoice(token string, amount int, memo string) (*InvoiceResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	payload := `{
		"value":` + strconv.Itoa(amount) + `,
		"memo":"` + memo + `",
		"webhooks": [
		{
			"url": "` + WebhookURL + `",
			"token": "boltathon",
			"enable_ssl_verification": false
		}
	]
	}`
	var jsonStr = []byte(payload)

	log.Println(fmt.Sprintf(BottlePayUserInvoiceURL, memo, amount))
	log.Println(token)

	req, err := http.NewRequest("POST", BottlePayUserInvoiceURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	log.Println(string(bodyBytes))

	if err != nil {
		log.Println(err)
		return nil, err
	}
	invoiceResponse := &InvoiceResponse{}
	json.Unmarshal(bodyBytes, &invoiceResponse)

	log.Println("howdy")
	log.Println(invoiceResponse)

	return invoiceResponse, nil
}

// NOT WORKING/NOT PUBLIC
func FetchUserBalance(token string) (*BalanceResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", BottlePayUserBalanceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	balanceResponse := &BalanceResponse{}
	json.Unmarshal(bodyBytes, &balanceResponse)

	log.Println("oknaybe")
	log.Println(balanceResponse)

	return balanceResponse, nil
}
