package stock

import (
	"time"

	"github.com/mjibson/goon"
)

var (
	jst = time.FixedZone("Asia/Tokyo", 9*60*60)
)

type CustomLoadSaver interface {
	Save() error
	Load() error
}

type Stock struct {
	ID         string    `json:"-" datastore:"-" goon:"id"`
	Group      string    `json:"group"`
	StoreCode  string    `json:"store_code"`
	JanIsbn    string    `json:"jan_isbn"`
	Quantity   int       `json:"quantity"`
	ResisterAt time.Time `json:"register_at"`
}

func (e *Stock) Save() error {
	e.ID = e.JanIsbn + "+" + e.StoreCode
	e.ResisterAt = time.Now()
	return nil
}

func (e *Stock) Load() error {
	e.ResisterAt = e.ResisterAt.In(jst)
	return nil
}

func Put(g *goon.Goon, e CustomLoadSaver) error {

	if err := e.Save(); err != nil {
		return err
	}

	_, err := g.Put(e)
	if err != nil {
		return err
	}

	return nil
}
