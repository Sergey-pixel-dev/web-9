package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

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

func handler(c echo.Context) error {
	cc := c.(*DBContext)
	name := c.QueryParam("name")
	time, err := cc.dbProvider.GetTimeLastVisit(name)
	if err == sql.ErrNoRows {
		err2 := cc.dbProvider.SetTimeLastVisit(name)
		if err2 != nil {
			cerr := echo.NewHTTPError(http.StatusInternalServerError, "Error: SetTimeLastVisit: "+err2.Error())
			return cerr.Unwrap()
		}
		c.String(http.StatusOK, "Hello, "+name)
	} else if err != nil {
		cerr := echo.NewHTTPError(http.StatusInternalServerError, "Error: GetTimeLastVisit: "+err.Error())
		return cerr.Unwrap()
	} else {
		c.String(http.StatusOK, "Hello, "+name+" . your last visit was in "+time)
		err = cc.dbProvider.UpdateTimeLastVisit(name)
		if err != nil {
			cerr := echo.NewHTTPError(http.StatusInternalServerError, "Error: UpdateTimeLastVisit: "+err.Error())
			return cerr.Unwrap()
		}
	}
	return nil
}

func (dp *DatabaseProvider) GetTimeLastVisit(name string) (string, error) {
	var msg string
	row := dp.db.QueryRow("SELECT time FROM labs where name=($1)", name) //чтоб наверняка последнее взяли
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}
func (dp *DatabaseProvider) UpdateTimeLastVisit(name string) error {
	_, err := dp.db.Exec("update labs set time = ($1) where name = ($2)", time.Now().Format("2006-01-02 15:04:05"), name)
	return err
}
func (dp *DatabaseProvider) SetTimeLastVisit(name string) error {
	_, err := dp.db.Exec("insert into labs (count, name, time) values (0, ($1), ($2))",
		name, time.Now().Format("2006-01-02 15:04:05"))
	return err
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
	e.GET("/api/user", handler)
	e.Start(*address)
}
