package api

import "github.com/auknl/warehouse/data"

// ResponseError is the only type of error response any user should ever get
type ResponseError struct {
	StatusCode int    `json:"code,omitempty"` //in case new error codes need to be designed
	Message    string `json:"message,omitempty"`
	Error      string `json:"errors,omitempty"`
}

// ResponseData is the holder for the actual data in an API response
type ResponseProduct struct {
	StatusCode    int                `json:"code,omitempty"` //in case new error codes need to be designed
	Products      []data.Product     `json:"products,omitempty"`
	Inventory     []data.Stock       `json:"inventory,omitempty"`
	ProductStocks data.ProductStocks `json:"product_stocks,omitempty"`
	Message       string             `json:"message,omitempty"`
}
