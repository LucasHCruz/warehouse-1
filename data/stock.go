package data

//Stock the inventory info per item
type Stock struct {
	ArtId string `json:"art_id,omitempty"`
	Name  string `json:"name,omitempty"`
	Stock string `json:"stock,omitempty"`
}

//type StockList []Stock

//Inventory stock info of all items
type Inventory struct {
	Inventory []Stock `json:"inventory"`
}
