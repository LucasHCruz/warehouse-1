package data

//ArticleContain is the map of product and required item/amount info
type ArticleContain struct {
	ArtId    string `json:"art_id,omitempty"`
	AmountOf string `json:"amount_of,omitempty"`
}

//Product represents product
type Product struct {
	Name            string           `json:"name,omitempty"`
	ContainArticles []ArticleContain `json:"contain_articles,omitempty"`
}

// Products represents all products
type Products struct {
	Products []Product `json:"products"`
}

//ProductStock keeps product and its stock for response
type ProductStock struct {
	Name               string `json:"product_name,omitempty"`
	AvailableProductNo string `json:"stock_of_product,omitempty"`
}

//ProductStocks list of ProductStock
type ProductStocks []ProductStock
