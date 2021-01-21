package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aukaskavalci/IKEA_assesment/api/mocks"
	"github.com/aukaskavalci/IKEA_assesment/data"
	"github.com/aukaskavalci/IKEA_assesment/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:generate mockgen -package=mocks -destination=./mocks/mock_Inventory.go -source=../db/inventory.go
func TestServer_getInventory(t *testing.T) {
	controller := gomock.NewController(t)
	//defer controller.Finish()
	recorder := httptest.NewRecorder()
	context, engine := gin.CreateTestContext(recorder)
	inventory := mocks.NewMockInventory(controller)
	stock := data.Stock{Stock: "9", Name: "test_item", ArtId: "1"}
	stockList := []data.Stock{stock}

	type fields struct {
		Inventory db.Inventory
		router    *gin.Engine
		Config    Configuration
		Logger    *logrus.Entry
	}
	type args struct {
		context *gin.Context
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantFail          bool
		statusCode        int
		expectedInventory data.Inventory
	}{
		{
			name:              "good_case",
			fields:            fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:              args{context: context},
			wantFail:          false,
			statusCode:        http.StatusOK,
			expectedInventory: data.Inventory{Inventory: stockList},
		},
		{
			name:              "bad_case",
			fields:            fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:              args{context: context},
			wantFail:          true,
			statusCode:        http.StatusNotFound,
			expectedInventory: data.Inventory{},
		},
	}
	for _, tt := range tests {
		f := func(t *testing.T) {
			server := &Server{
				Inventory: tt.fields.Inventory,
				router:    tt.fields.router,
				Config:    tt.fields.Config,
				Logger:    tt.fields.Logger,
			}
			server.getInventory(tt.args.context)
		}

		if !tt.wantFail {
			inventory.EXPECT().GetInventory(context).Return(nil, stockList)
		} else {
			inventory.EXPECT().GetInventory(context).Return(errors.New("query test err"), nil)
		}
		t.Run(tt.name, f)
		assert.Equal(t, tt.statusCode, context.Writer.Status())
		var response ResponseProduct
		if !tt.wantFail {
			byteArr, _ := ioutil.ReadAll(recorder.Body)
			_ = json.Unmarshal(byteArr, &response)
			assert.Equal(t, response.Inventory, tt.expectedInventory.Inventory)
		}
	}
}

func TestServer_getProductStock(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := httptest.NewRecorder()
	context, engine := gin.CreateTestContext(recorder)
	inventory := mocks.NewMockInventory(controller)

	type fields struct {
		Inventory db.Inventory
		router    *gin.Engine
		Config    Configuration
		Logger    *logrus.Entry
	}
	type args struct {
		context *gin.Context
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		checkStock         bool
		expectedStock      data.ProductStocks
		expectedStatusCode int
		queryFail          bool
	}{
		{
			name:               "no_product_in_stock",
			fields:             fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:               args{context: context},
			expectedStock:      []data.ProductStock{},
			expectedStatusCode: http.StatusOK,
			checkStock:         false, // "No product in stock"
			queryFail:          false,
		},
		{
			name:               "error_query",
			fields:             fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:               args{context: context},
			expectedStock:      []data.ProductStock{},
			expectedStatusCode: http.StatusNotFound,
			checkStock:         false, // query fails
			queryFail:          true,
		},
		{
			name:               "in_stock_product",
			fields:             fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:               args{context: context},
			expectedStock:      []data.ProductStock{{Name: "test_product", AvailableProductNo: "10"}},
			expectedStatusCode: http.StatusOK,
			checkStock:         true, // query fails
			queryFail:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder.Body.Reset()
			server := &Server{
				Inventory: tt.fields.Inventory,
				router:    tt.fields.router,
				Config:    tt.fields.Config,
				Logger:    tt.fields.Logger,
			}

			if tt.queryFail {
				inventory.EXPECT().GetProductStock(context).Return(errors.New("query failed test"), nil)
			} else {
				inventory.EXPECT().GetProductStock(context).Return(nil, tt.expectedStock)
			}

			server.getProductStock(tt.args.context)

			assert.Equal(t, tt.expectedStatusCode, context.Writer.Status())
			var response ResponseProduct
			if tt.checkStock {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &response)
				assert.Equal(t, response.ProductStocks, tt.expectedStock)
			}
		})
	}
}

func TestServer_uploadInventory(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := httptest.NewRecorder()
	context, engine := gin.CreateTestContext(recorder)
	inventory := mocks.NewMockInventory(controller)

	type fields struct {
		Inventory db.Inventory
		router    *gin.Engine
		Config    Configuration
		Logger    *logrus.Entry
	}
	type args struct {
		context *gin.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFail   bool
		statusCode int
		message    string
	}{
		{
			name:       "upload_fail",
			fields:     fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:       args{context: context},
			wantFail:   true,
			statusCode: http.StatusBadRequest,
			message:    "upload failed test",
		},
		{
			name:       "upload_an_inventory",
			fields:     fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:       args{context: context},
			wantFail:   false,
			statusCode: http.StatusOK,
			message:    "1 item inserted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Inventory: tt.fields.Inventory,
				router:    tt.fields.router,
				Config:    tt.fields.Config,
				Logger:    tt.fields.Logger,
			}

			stocks := data.Inventory{Inventory: []data.Stock{{Name: "test", ArtId: "1", Stock: "1"}}}
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(stocks)
			context.Request = &http.Request{Body: ioutil.NopCloser(bytes.NewBuffer(reqBodyBytes.Bytes()))}

			if tt.wantFail {
				inventory.EXPECT().UploadInventory(context, gomock.Any()).Return(errors.New("upload failed test"), 0)
			} else {
				inventory.EXPECT().UploadInventory(context, gomock.Any()).Return(nil, 1)
			}

			server.uploadInventory(tt.args.context)

			assert.Equal(t, tt.statusCode, context.Writer.Status())
			var response ResponseProduct
			var responseErr ResponseError
			if !tt.wantFail {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &response)
				assert.Equal(t, response.Message, tt.message)
			} else {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &responseErr)
				assert.Equal(t, responseErr.Message, tt.message)
			}

		})
	}
}

func TestServer_uploadProducts(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := httptest.NewRecorder()
	context, engine := gin.CreateTestContext(recorder)
	inventory := mocks.NewMockInventory(controller)

	type fields struct {
		Inventory db.Inventory
		router    *gin.Engine
		Config    Configuration
		Logger    *logrus.Entry
	}
	type args struct {
		context *gin.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFail   bool
		statusCode int
		message    string
	}{
		{
			name:       "upload_fail",
			fields:     fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:       args{context: context},
			wantFail:   true,
			statusCode: http.StatusBadRequest,
			message:    "upload product failed test",
		},
		{
			name:       "upload_an_product",
			fields:     fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:       args{context: context},
			wantFail:   false,
			statusCode: http.StatusOK,
			message:    "1 product inserted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Inventory: tt.fields.Inventory,
				router:    tt.fields.router,
				Config:    tt.fields.Config,
				Logger:    tt.fields.Logger,
			}

			products := data.Products{Products: []data.Product{{Name: "test_product", ContainArticles: []data.ArticleContain{{ArtId: "test_item", AmountOf: "1"}}}}}
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(products)
			context.Request = &http.Request{Body: ioutil.NopCloser(bytes.NewBuffer(reqBodyBytes.Bytes()))}

			if tt.wantFail {
				inventory.EXPECT().UploadProducts(context, gomock.Any()).Return(errors.New("upload product failed test"), 0)
			} else {
				inventory.EXPECT().UploadProducts(context, gomock.Any()).Return(nil, 1)
			}

			server.uploadProducts(tt.args.context)

			assert.Equal(t, tt.statusCode, context.Writer.Status())
			var response ResponseProduct
			var responseErr ResponseError
			if !tt.wantFail {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &response)
				assert.Equal(t, response.Message, tt.message)
			} else {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &responseErr)
				assert.Equal(t, responseErr.Message, tt.message)
			}

		})
	}
}

func TestServer_sellProduct(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := httptest.NewRecorder()
	context, engine := gin.CreateTestContext(recorder)
	inventory := mocks.NewMockInventory(controller)
	context.Params = []gin.Param{
		{
			Key:   productName,
			Value: "product_test",
		},
	}

	type fields struct {
		Inventory db.Inventory
		router    *gin.Engine
		Config    Configuration
		Logger    *logrus.Entry
	}
	type args struct {
		context *gin.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFail   bool
		statusCode int
		message    string
	}{
		{
			name:       "product_sold",
			fields:     fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:       args{context: context},
			wantFail:   false,
			statusCode: http.StatusOK,
			message:    "Product product_test is sold and inventory is updated accordingly",
		},
		{
			name:       "product_not_sold",
			fields:     fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:       args{context: context},
			wantFail:   true,
			statusCode: http.StatusBadRequest,
			message:    "sell product failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Inventory: tt.fields.Inventory,
				router:    tt.fields.router,
				Config:    tt.fields.Config,
				Logger:    tt.fields.Logger,
			}

			if tt.wantFail {
				inventory.EXPECT().SellProduct(context, gomock.Any()).Return(errors.New("sell product failed"))
			} else {
				inventory.EXPECT().SellProduct(context, gomock.Any()).Return(nil)
			}

			server.sellProduct(tt.args.context)

			assert.Equal(t, tt.statusCode, context.Writer.Status())
			var response ResponseProduct
			var responseErr ResponseError
			if !tt.wantFail {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &response)
				assert.Equal(t, response.Message, tt.message)
			} else {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &responseErr)
				assert.Equal(t, responseErr.Message, tt.message)
			}

		})
	}
}

func TestServer_isHealthy(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := httptest.NewRecorder()
	context, engine := gin.CreateTestContext(recorder)
	inventory := mocks.NewMockInventory(controller)

	type fields struct {
		Inventory db.Inventory
		router    *gin.Engine
		Config    Configuration
		Logger    *logrus.Entry
	}
	type args struct {
		context *gin.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFail   bool
		statusCode int
		message    string
	}{
		{
			name:       "ping_unhealthy",
			fields:     fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:       args{context: context},
			wantFail:   true,
			statusCode: http.StatusInternalServerError,
			message:    "unhealthy endpoint",
		},
		{
			name:       "ping_healthy",
			fields:     fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:       args{context: context},
			wantFail:   false,
			statusCode: http.StatusOK,
			message:    "healthy endpoint",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Inventory: tt.fields.Inventory,
				router:    tt.fields.router,
				Config:    tt.fields.Config,
				Logger:    tt.fields.Logger,
			}
			if tt.wantFail {
				inventory.EXPECT().Ping().Return(errors.New("unhealthy"))
			} else {
				inventory.EXPECT().Ping().Return(nil)
			}

			server.isHealthy(tt.args.context)

			assert.Equal(t, tt.statusCode, context.Writer.Status())
			var response ResponseProduct
			var responseErr ResponseError
			if !tt.wantFail {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &response)
				assert.Equal(t, response.Message, tt.message)
			} else {
				byteArr, _ := ioutil.ReadAll(recorder.Body)
				_ = json.Unmarshal(byteArr, &responseErr)
				assert.Equal(t, responseErr.Message, tt.message)
			}

		})
	}
}

func TestServer_setRID(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := httptest.NewRecorder()
	contextWithID, engine := gin.CreateTestContext(recorder)
	contextWithID.Set("rid", "test")
	contextWithoutID, engine := gin.CreateTestContext(recorder)

	inventory := mocks.NewMockInventory(controller)

	type fields struct {
		Inventory db.Inventory
		router    *gin.Engine
		Config    Configuration
		Logger    *logrus.Entry
	}
	type args struct {
		context *gin.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		isExist bool
	}{
		{
			name:    "new uuid",
			fields:  fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:    args{context: contextWithoutID},
			isExist: false,
		},
		{
			name:    "current uuid",
			fields:  fields{Logger: logrus.NewEntry(logrus.New()), router: engine, Inventory: inventory, Config: Configuration{ListenAddress: "localhost:8080", BackendTimeout: "25s"}},
			args:    args{context: contextWithID},
			isExist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Inventory: tt.fields.Inventory,
				router:    tt.fields.router,
				Config:    tt.fields.Config,
				Logger:    tt.fields.Logger,
			}

			server.setRID(tt.args.context)
			rid, _ := tt.args.context.Get("rid")
			if tt.isExist {
				assert.Equal(t, rid, "test")
			} else {
				assert.Equal(t, rid != "", true)
			}
		})
	}
}
