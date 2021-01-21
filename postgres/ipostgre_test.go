// +build integration

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aukaskavalci/IKEA_assesment/data"
	"github.com/aukaskavalci/IKEA_assesment/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
	"io/ioutil"
	"log"
	"net"
	"net/http/httptest"
	"net/url"
	"runtime"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type dockerDBConn struct {
	Conn *sql.DB
}

var (
	// DockerDBConn holds the connection to our DB in the container
	DockerDBConn *dockerDBConn
	logger       = logrus.New()
)

// refs: https://github.com/ory/dockertest
func initDB(log *logrus.Logger) (*dockertest.Pool, *dockertest.Resource) {
	pgURL := initPostgres()
	pgPass, _ := pgURL.User.Password()

	runOpts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=" + pgURL.User.Username(),
			"POSTGRES_PASSWORD=" + pgPass,
			"POSTGRES_DB=" + pgURL.Path,
		},
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.WithError(err).Fatal("Could not connect to docker")
	}

	resource, err := pool.RunWithOptions(&runOpts)
	if err != nil {
		log.WithError(err).Fatal("Could start postgres container")
	}

	pgURL.Host = resource.Container.NetworkSettings.IPAddress

	// Docker layer network is different on Mac
	if runtime.GOOS == "darwin" {
		pgURL.Host = net.JoinHostPort(resource.GetBoundIP("5432/tcp"), resource.GetPort("5432/tcp"))
	}

	DockerDBConn = &dockerDBConn{}
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		DockerDBConn.Conn, err = sql.Open("postgres", pgURL.String())
		if err != nil {
			return err
		}
		return DockerDBConn.Conn.Ping()
	}); err != nil {
		phrase := fmt.Sprintf("Could not connect to docker: %s", err)
		log.Error(phrase)
	}

	DockerDBConn.initMigrations()

	return pool, resource
}

func closeDB(pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		phrase := fmt.Sprintf("Could not purge resource: %s", err)
		log.Fatal(phrase)
	}
}

func (db dockerDBConn) initMigrations() {
	driver, err := postgres.WithInstance(db.Conn, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	migrateSql, err := migrate.NewWithDatabaseInstance(
		"file://../db/migrations",
		"inventory", driver)
	if err != nil {
		log.Fatal(err)
	}

	err = migrateSql.Steps(1)
	if err != nil {
		log.Fatal(err)
	}
}

func initPostgres() *url.URL {
	pgURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("postgres", "1234"),
		Path:   "inventory",
	}
	q := pgURL.Query()
	q.Add("sslmode", "disable")
	pgURL.RawQuery = q.Encode()

	return pgURL
}

var products data.Products
var inventoryData data.Inventory

func uploadInventory(inventorydb db.Inventory, ctx context.Context) {
	logger := logrus.NewEntry(logrus.New())
	file, _ := ioutil.ReadFile("./testdata/example_inventory.json")
	json.Unmarshal([]byte(file), &inventoryData)

	err, _ := inventorydb.UploadInventory(ctx, inventoryData)
	if err != nil {
		logger.WithError(err).Fatal("Could not upload inventory")
	}
}

func uploadProduct(inventorydb db.Inventory, ctx context.Context) {
	file, _ := ioutil.ReadFile("./testdata/example_products.json")
	json.Unmarshal([]byte(file), &products)

	err, _ := inventorydb.UploadProducts(ctx, products)
	if err != nil {
		logger.WithError(err).Fatal("Could not upload products")
	}
}

func TestPInventoryDB_UploadInventory(t *testing.T) {
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New())},
	}

	file, _ := ioutil.ReadFile("./testdata/example_inventory.json") //testdata sirayi check et!!
	json.Unmarshal([]byte(file), &inventoryData)

	err, stock := inventory.UploadInventory(ctx, inventoryData)
	assert.Equal(t, stock, len(inventoryData.Inventory))
	assert.Equal(t, err, nil)
	closeDB(pool, resource)
}

func TestPInventoryDB_UploadProducts(t *testing.T) {
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New())},
	}

	file, _ := ioutil.ReadFile("./testdata/example_products.json")
	json.Unmarshal([]byte(file), &products)

	//Product table has foreign key from Inventory table
	uploadInventory(inventory, context)

	err, stock := inventory.UploadProducts(context, products)
	assert.Equal(t, stock, len(products.Products))
	assert.Equal(t, err, nil)
	closeDB(pool, resource)
}

func TestPInventoryDB_GetInventory(t *testing.T) {
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New())},
	}
	uploadInventory(inventory, ctx)

	err, stock := inventory.GetInventory(ctx)
	assert.Equal(t, len(stock), len(inventoryData.Inventory))
	for i := range inventoryData.Inventory {
		assert.Equal(t, stock[i], inventoryData.Inventory[i])
	}
	assert.Equal(t, err, nil)
	closeDB(pool, resource)
}

func TestPInventoryDB_GetProductStock(t *testing.T) {
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New())},
	}
	//fill the tables before apply query
	uploadInventory(inventory, ctx)
	uploadProduct(inventory, ctx)

	err, stockOfProduct := inventory.GetProductStock(ctx)
	assert.Equal(t, len(stockOfProduct), 2)
	assert.Equal(t, err, nil)
	closeDB(pool, resource)
}

func TestPInventoryDB_SellProduct(t *testing.T) {
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New())},
	}

	//fill the tables before apply query
	uploadInventory(inventory, ctx)
	uploadProduct(inventory, ctx)

	err := inventory.SellProduct(ctx, "Dinning Table")
	assert.Equal(t, err, nil)
	closeDB(pool, resource)
}

func TestPInventoryDB_GetProductStockOOS(t *testing.T) { //After One "Dinning Table" Product Out Of Stock
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New())},
	}

	//fill the tables before apply query
	uploadInventory(inventory, ctx)
	uploadProduct(inventory, ctx)

	//Only one product was in the stock,selling it
	inventory.SellProduct(ctx, "Dinning Table")

	err, stockOfProduct := inventory.GetProductStock(ctx)
	assert.Equal(t, len(stockOfProduct), 1)
	assert.Equal(t, err, nil)
	closeDB(pool, resource)
}

func TestPInventoryDB_Ping(t *testing.T) {
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New()), Dbname: "inventory", Port: "5432", Host: "localhost", Driver: "postgres", Password: "1234", User: "postgres"},
	}
	err := inventory.Ping()
	assert.Equal(t, err, nil)
	closeDB(pool, resource)
}

func TestPInventoryDB_Open(t *testing.T) {
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New()), Dbname: "inventory", Port: "5432", Host: "localhost", Driver: "postgres", Password: "1234", User: "postgres"},
	}
	err := inventory.Open()
	assert.Equal(t, err, nil)
	closeDB(pool, resource)
}

func TestPInventoryDB_SellOOSProduct(t *testing.T) { //Try to sell "Dinning Table" which is  Out Of Stock
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New())},
	}
	//fill the tables before apply query
	uploadInventory(inventory, ctx)
	uploadProduct(inventory, ctx)

	//Only one product was in the stock,selling it
	inventory.SellProduct(ctx, "Dinning Table")

	err := inventory.SellProduct(ctx, "Dinning Table")
	if err != nil {
		assert.Equal(t, err.Error(), "this product is not in stock, cannot be sold")
	}
	closeDB(pool, resource)
}

func TestPInventoryDB_SellProductNotExist(t *testing.T) { //Try to sell a product that is not in system
	pool, resource := initDB(logger)
	conn := DockerDBConn.Conn
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	inventory := &PInventoryDB{
		db:     conn,
		config: Config{Logger: logrus.NewEntry(logrus.New())},
	}

	//fill the tables before apply query
	uploadInventory(inventory, ctx)
	uploadProduct(inventory, ctx)

	err := inventory.SellProduct(ctx, "NotExist")
	if err != nil {
		assert.Equal(t, err.Error(), "this product is not in system, cannot be sold")
	}
	closeDB(pool, resource)
}
