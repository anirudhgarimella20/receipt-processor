package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"crypto/sha1"
	"encoding/hex"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

type Endpoints struct {
	Process endpoint.Endpoint
	Points  endpoint.Endpoint
}

type Item struct {
	Description string `json:"shortDescription"`
	Price       string `json:"price"`
}

type (
	ProcessRequest struct {
		Retailer     string `json:"retailer"`
		PurchaseDate string `json:"purchaseDate"`
		PurchaseTime string `json:"purchaseTime"`
		Items        []Item `json:"items"`
		Total        string `json:"total"`
	}
	ProcessResponse struct {
		Id string `json:id`
	}
	PointsRequest struct {
		Id string `json:"id"`
	}
	PointsResponse struct {
		Points int `json:"points"`
	}
)
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	Purchasetime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

var global_score map[string]int

type Service interface {
	//receipts/process
	ProcessReceipt(ctx context.Context, receipt ProcessRequest) (string, error)
	Points(ctx context.Context, id string) (int, error)
}

func ProcessReceipt(ctx context.Context, receipt ProcessRequest) (string, error) {
	//fmt.Printf("process receipt %v", receipt)
	var score int
	score = 0

	// Count alphanumeric characters in the retailer name, ignoring spaces
	alphanumericCount := 0
	for _, char := range receipt.Retailer {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			alphanumericCount++
		}
	}

	// Add the alphanumeric count to the score
	score += alphanumericCount

	total, err := strconv.ParseFloat(strings.TrimSpace(receipt.Total), 64)
	if err == nil && math.Trunc(total) == total {
		score += 50

	}
	if math.Mod(total,0.25) == 0 {
		score += 25
	}
	if len(receipt.Items) > 0 {
		score += (len(receipt.Items) / 2) * 5
	}

	var pts float64
	for i, _ := range receipt.Items {
		if len(receipt.Items[i].Description)%3 == 0 {
			price, err := strconv.ParseFloat(receipt.Items[i].Price, 64)
			if err == nil {
				pts += math.Ceil(price * 0.2)
			}
		}
	}
	score += int(pts)


	dateList := strings.Split(receipt.PurchaseDate, "-")
	if len(dateList) != 3 {
		return "", errors.New("invalid date format")
	} else {
		day, err := strconv.Atoi(dateList[2:][0])
		if err != nil {
			return "", errors.New("invalid format, day not a numeral")
		}
		if day%2 != 0 {
			score += 6
		}
	}

	timeList := strings.Split(receipt.PurchaseTime, ":")
	if len(timeList) != 2 {
		return "", errors.New("invalid time format")
	} else {
		hour, err := strconv.Atoi(timeList[:1][0])
		if err != nil {
			return "", errors.New("invalid format, time not a numeral")
		}

		if hour >= 14 && hour <= 16 {
			score += 10
		}
	}

	//fmt.Printf("score %d", score)
	hasher := sha1.New()
	_, _ = hasher.Write([]byte(fmt.Sprintf("%v", receipt)))
	sha1_hash := hex.EncodeToString(hasher.Sum(nil))
	global_score[sha1_hash] = score
	return sha1_hash, err
}

func MakeEndpoints(s Service) Endpoints {
	return Endpoints{
		Process: makeProccessEndpoint(s),
		Points:  makePointsEndpoint(s),
	}
}

func makeProccessEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ProcessRequest)
		fmt.Printf("%v", req)

		id, err := ProcessReceipt(ctx, req)

		return ProcessResponse{Id: id}, err

	}
}
func makePointsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PointsRequest)
		//fmt.Printf(" req for id    %v", req)

		return PointsResponse{Points: global_score[req.Id]}, nil
	}

}

func main() {
	var httpAddr = flag.String("http", ":8080", "http listen address")
	global_score = map[string]int{}

	ctx := context.Background()
	var srv Service

	endpoints := MakeEndpoints(srv)
	errs := make(chan error)

	go func() {
		handler := NewHttpServer(ctx, endpoints)
		errs <- http.ListenAndServe(*httpAddr, handler)
	}()

	println("errors ", <-errs)
}

func CommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")
		next.ServeHTTP(w, r)

	})

}

func NewHttpServer(ctx context.Context, endpoints Endpoints) http.Handler {
	r := mux.NewRouter()
	r.Use(CommonMiddleware)

	r.Methods("POST").Path("/receipts/process").Handler(httptransport.NewServer(endpoints.Process, decodeProcessReceiptRequest, encodeResponse))
	r.Methods("GET").Path("/receipts/{ID}/points").Handler(httptransport.NewServer(endpoints.Points, decodePointsReq, encodeResponse))

	return r
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func decodeProcessReceiptRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req ProcessRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	//fmt.Printf("decoding processReceipt req %v\n", req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func decodePointsReq(ctx context.Context, r *http.Request) (interface{}, error) {
	var req PointsRequest
	vars := mux.Vars(r)
	//fmt.Println("req points for ID", vars["ID"])
	req = PointsRequest{Id: vars["ID"]}
	return req, nil
}
