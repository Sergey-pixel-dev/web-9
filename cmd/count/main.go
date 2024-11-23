package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "bmstu"
)

type DatabaseProvider struct {
	db *sql.DB
}
type DBContext struct {
	echo.Context
	dbProvider DatabaseProvider
}
type envelope map[string]string

func handlerPOST(c echo.Context) error {
	cc := c.(*DBContext)
	i2 := envelope{"count": "0"}
	if err := c.Bind(&i2); err != nil {
		cerr := echo.NewHTTPError(http.StatusBadRequest, "Error: json: "+err.Error())
		return cerr.Unwrap()
	}
	i, err := strconv.Atoi(i2["count"])
	if err != nil {
		cerr := echo.NewHTTPError(http.StatusBadRequest, "Error: not number: "+err.Error())
		return cerr.Unwrap()
	}
	count, err := cc.dbProvider.GetCount()
	if err != nil {
		cerr := echo.NewHTTPError(http.StatusInternalServerError, "Error: GetCount(): "+err.Error())
		return cerr.Unwrap()
	}
	err = cc.dbProvider.UpdateCount(count + i)
	if err != nil {
		cerr := echo.NewHTTPError(http.StatusInternalServerError, "Error: UpdateCount(): "+err.Error())
		return cerr.Unwrap()
	}
	return nil
}

func handlerGET(c echo.Context) error {
	cc := c.(*DBContext)
	count, err := cc.dbProvider.GetCount()
	if err != nil {
		cerr := echo.NewHTTPError(http.StatusInternalServerError, "Error: GetCount(): "+err.Error())
		return cerr.Unwrap()
	}
	return c.String(http.StatusOK, strconv.Itoa(count))
}

func (dp *DatabaseProvider) GetCount() (int, error) {
	var msg string
	row := dp.db.QueryRow("SELECT count FROM labs order by id LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(msg)
}
func (dp *DatabaseProvider) UpdateCount(count int) error {
	_, err := dp.db.Exec("update labs set count = ($1)", count)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	e := echo.New()
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &DBContext{Context: c, dbProvider: dp}
			return next(cc)
		}
	})
	e.GET("/", handlerGET)
	e.POST("/", handlerPOST)
	err = e.Start(*address)
	if err != nil {
		log.Fatal(err)
	}
}
