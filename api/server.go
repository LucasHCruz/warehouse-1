package api

import (
	"encoding/json"
	"fmt"
	"github.com/aukaskavalci/IKEA_assesment/data"
	"github.com/aukaskavalci/IKEA_assesment/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

// Server serves HTTP requests
type Server struct {
	Inventory db.Inventory
	router    *gin.Engine
	Config    Configuration
	Logger    *logrus.Entry
}

//Configuration keeps required info for running server
type Configuration struct {
	BackendTimeout string `default:"25s"`
	ListenAddress  string `default:":8080"`
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(inventory db.Inventory) *Server {
	server := &Server{Inventory: inventory}
	router := gin.New()

	router.Use(
		gin.Recovery(),
		server.setRID,
		server.setDeadline, //TODO: use deadline while querying db
	)

	router.GET("v1/health", server.isHealthy)
	router.GET("v1/inventory", server.getInventory)
	router.GET("v1/product", server.getProductStock)
	router.POST("v1/product", server.uploadProducts)
	router.POST("v1/inventory", server.uploadInventory)
	router.POST("v1/product/:"+productName, server.sellProduct)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start() error {
	return server.router.Run(server.Config.ListenAddress)
}

//setDeadline sets the deadline to limit the process time of the request
func (server *Server) setDeadline(context *gin.Context) {
	backendTimeout, err := time.ParseDuration(server.Config.BackendTimeout)
	if err != nil {
		server.Logger.WithField("err", err).Fatal("Could not parse backend timeout duration")
	}

	deadline := time.Now().Add(backendTimeout)
	context.Set("deadline", deadline)
}

//setRID sets the rid by generating or using the existing one
func (server *Server) setRID(context *gin.Context) {
	_, exists := context.Get("rid")
	if !exists {
		id := uuid.New()
		context.Set("rid", id.String())
	}

}

//isHealthy checks if the service is available to respond
func (server *Server) isHealthy(context *gin.Context) {
	rid := context.GetString("rid")
	server.Logger.WithField("rid:", rid).Debug("isHealthy")
	err := server.Inventory.Ping()
	if err != nil {
		server.Logger.WithField("err", err.Error()).Error("isHealthy ping failed")
		context.JSON(http.StatusInternalServerError, ResponseError{
			Message: "unhealthy endpoint",
		})
		return
	}
	context.JSON(http.StatusOK, ResponseError{
		Message: "healthy endpoint",
	})
	return
}

//getInventory provides inventory/stock info
func (server *Server) getInventory(context *gin.Context) {
	rid := context.GetString("rid")
	server.Logger.WithField("rid:", rid).Debug("getInventory")
	err, stocks := server.Inventory.GetInventory(context)
	if err != nil {
		context.JSON(http.StatusNotFound, ResponseError{
			Message: err.Error(),
		})
		return
	} else {
		context.JSON(http.StatusOK, ResponseProduct{
			Inventory: stocks,
		})
		return
	}
}

// getProductStock provides the stock info of available products in system
func (server *Server) getProductStock(context *gin.Context) {
	rid := context.GetString("rid")
	server.Logger.WithField("rid:", rid).Debug("getProductStock")
	err, stocks := server.Inventory.GetProductStock(context)
	if err != nil {
		context.JSON(http.StatusNotFound, ResponseError{
			Message: err.Error(),
		})
		return
	} else {
		var product ResponseProduct
		if len(stocks) == 0 {
			product = ResponseProduct{
				Message: "No product in stock",
			}
		} else {
			product = ResponseProduct{
				ProductStocks: stocks,
			}
		}
		context.JSON(http.StatusOK, product)
		return
	}
}

//uploadProducts inserts given products to system
func (server *Server) uploadProducts(context *gin.Context) {
	rid := context.GetString("rid")
	server.Logger.WithField("rid:", rid).Debug("uploadProducts")
	var products data.Products
	jsonData, err := ioutil.ReadAll(context.Request.Body)
	err = json.Unmarshal(jsonData, &products)
	if err != nil {
		context.JSON(http.StatusBadRequest, ResponseError{
			Message: err.Error(),
		})
		return
	}

	insertedRecord := 0
	err, insertedRecord = server.Inventory.UploadProducts(context, products)
	if err != nil {
		context.JSON(http.StatusBadRequest, ResponseError{
			Message: err.Error(),
		})
		return
	} else {
		message := fmt.Sprintf("%d product inserted", insertedRecord)
		context.JSON(http.StatusOK, ResponseProduct{
			Message: message,
		})
		return
	}
}

//uploadInventory inserts given inventory/stock info to system
func (server *Server) uploadInventory(context *gin.Context) {
	rid := context.GetString("rid")
	server.Logger.WithField("rid:", rid).Debug("uploadInventory")
	var inventory data.Inventory
	jsonData, err := ioutil.ReadAll(context.Request.Body)
	err = json.Unmarshal(jsonData, &inventory)
	if err != nil {
		context.JSON(http.StatusBadRequest, ResponseError{
			Message: err.Error(),
		})
		return
	}

	insertedInventory := 0
	err, insertedInventory = server.Inventory.UploadInventory(context, inventory)
	if err != nil {
		context.JSON(http.StatusBadRequest, ResponseError{
			Message: err.Error(),
		})
		return
	} else {
		message := fmt.Sprintf("%d item inserted", insertedInventory)
		context.JSON(http.StatusOK, ResponseProduct{
			Message: message,
		})
		return
	}
}

//sellProduct handles the sell product request
func (server *Server) sellProduct(context *gin.Context) {
	rid := context.GetString("rid")
	server.Logger.WithField("rid:", rid).Debug("sellProduct")
	productName := context.Param(productName)
	err := server.Inventory.SellProduct(context, productName)
	if err != nil {
		context.JSON(http.StatusBadRequest, ResponseError{
			Message: err.Error(),
		})
		return
	} else {
		message := fmt.Sprintf("Product %s is sold and inventory is updated accordingly", productName)
		context.JSON(http.StatusOK, ResponseProduct{
			Message: message,
		})
		return
	}
}
