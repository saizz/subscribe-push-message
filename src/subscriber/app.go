package subscriber

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock"

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

	s := new(stock.Stock)
	if err := json.Unmarshal(msg.Message.Data, s); err != nil {
		log.Errorf(ctx, "json Unmarshal stock, %v", err.Error())
		return nil
	}

	if err := stock.Put(goon.FromContext(ctx), s); err != nil {
		log.Errorf(ctx, "stock.Put, %v", err.Error())
		return nil
	}

	return nil
}
