package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

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

type DBContext struct {
	echo.Context
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

type message struct {
	Msg string `json:"msg"`
}

func GetHello(c echo.Context) error {
	cc := c.(*DBContext)
	msg, err := cc.dbProvider.SelectHello()
	if err != nil {
		cerr := echo.NewHTTPError(http.StatusInternalServerError, "Error: SetTimeLastVisit: "+err.Error())
		return cerr.Unwrap()
	}

	return c.String(http.StatusOK, msg)
}
func PostHello(c echo.Context) error {
	cc := c.(*DBContext)
	input := new(message)

	err := c.Bind(input)
	if err != nil {
		cerr := echo.NewHTTPError(http.StatusBadRequest, "Error: Decode: "+err.Error())
		return cerr.Unwrap()
	}

	err = cc.dbProvider.InsertHello(input.Msg)
	if err != nil {
		cerr := echo.NewHTTPError(http.StatusBadRequest, "Error: InsertHello(): "+err.Error())
		return cerr.Unwrap()
	}

	return c.String(http.StatusCreated, "")
}

func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string

	row := dp.db.QueryRow("SELECT message FROM labs WHERE message IS NOT NULL AND message != '' ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}
func (dp *DatabaseProvider) InsertHello(msg string) error {
	_, err := dp.db.Exec("INSERT INTO labs (message) VALUES ($1)", msg)
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
	e.GET("/", GetHello)
	e.POST("/", PostHello)
	err = e.Start(*address)
	if err != nil {
		log.Fatal(err)
	}
}
