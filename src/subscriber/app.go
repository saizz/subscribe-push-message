package subscriber

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stock"

	"github.com/gocarina/gocsv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mjibson/goon"

	"cloud.google.com/go/pubsub"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var e *echo.Echo

func init() {

	e = newMux()
	g := e.Group("/_ah/push-handlers")
	g.POST("/stock-datastore", pushMassageHandler)

	gocsv.SetCSVReader(func(in io.Reader) *csv.Reader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		return r
	})
}

func newMux() *echo.Echo {
	e := echo.New()
	http.Handle("/_ah/", e)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowMethods: []string{echo.POST},
	}))
	e.Use(middleware.BodyDump(httpResponseBodyDumper))
	return e

}

func httpResponseBodyDumper(c echo.Context, req, res []byte) {

	ctx := appengine.NewContext(c.Request())
	log.Infof(ctx, "response body: %v", string(res))
}

type pushMessage struct {
	Message pubsub.Message `json:"message"`
}

type jsonResponse struct {
	Message string `json:"message"`
}

func pushMassageHandler(c echo.Context) error {

	ctx := appengine.NewContext(c.Request())

	msg := new(pushMessage)
	if err := c.Bind(msg); err != nil {
		res := &jsonResponse{
			Message: fmt.Sprintf("%+v", err),
		}

		return c.JSON(http.StatusOK, res)
	}

	ty, ok := msg.Message.Attributes["type"]
	if ok && ty == "tsv" {
		return multiPutFromTSV(ctx, msg.Message.Data)
	}

	return putFromJSON(ctx, msg.Message.Data)

}

func multiPutFromTSV(ctx context.Context, b []byte) error {

	stocks := []*stock.Stock{}

	if err := gocsv.UnmarshalBytes(b, &stocks); err != nil {
		log.Errorf(ctx, "gocsv unmarshal, %v", err.Error())
		return nil
	}

	for _, s := range stocks {
		log.Infof(ctx, "stock: %+v", s)
		if err := stock.Put(goon.FromContext(ctx), s); err != nil {
			log.Errorf(ctx, "stock.Put,%v", err.Error())
		}
	}

	return nil
}

func putFromJSON(ctx context.Context, b []byte) error {

	s := new(stock.Stock)
	if err := json.Unmarshal(b, s); err != nil {
		log.Errorf(ctx, "json Unmarshal stock, %v", err.Error())
		return nil
	}

	if err := stock.Put(goon.FromContext(ctx), s); err != nil {
		log.Errorf(ctx, "stock.Put, %v", err.Error())
		return nil
	}

	return nil

}
