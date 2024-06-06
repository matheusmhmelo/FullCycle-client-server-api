package main

import (
	"github.com/go-chi/chi"
	"github.com/matheusmhmelo/FullCycle-client-server-api/internal/entity"
	"github.com/matheusmhmelo/FullCycle-client-server-api/internal/infra/database"
	"github.com/matheusmhmelo/FullCycle-client-server-api/internal/infra/webserver/handlers"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	db, err := gorm.Open(sqlite.Open("dollar.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&entity.Dollar{})
	dollarDB := database.NewDollar(db)
	dollarHandler := handlers.NewDollarHandler(dollarDB)

	r := chi.NewRouter()
	r.Get("/cotacao", dollarHandler.GetDollar)

	http.ListenAndServe(":8080", r)
}
