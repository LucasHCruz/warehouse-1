package db

import (
	"context"
	"github.com/auknl/warehouse/data"
)

type Inventory interface {
	Ping() error
	Open() error
	GetInventory(ctx context.Context) (error, []data.Stock)
	GetProductStock(ctx context.Context) (error, data.ProductStocks)
	UploadProducts(ctx context.Context, product data.Products) (error, int)
	UploadInventory(ctx context.Context, inventory data.Inventory) (error, int)
	SellProduct(ctx context.Context, productName string) error
}
