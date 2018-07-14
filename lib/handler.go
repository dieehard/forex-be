//handler.go
package lib

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"strconv"
	"time"
)

type AppHandler struct {
	Router *mux.Router
	DB     *sql.DB
}

const InvalidRequest = "Invalid request payload"

// Initialize our handler
func (a *AppHandler) Initialize(user, password, host, port, db string) {
	//DSN format: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	//Go time.Time, can be parser as mySQL DATE or DATETIME
	//if parseTime is set false, we get DATE format
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=false", user, password, host, port, db)
	//fmt.Println(dsn)
	var err error
	a.DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

// Run at given port
func (a *AppHandler) Run(addr string) {
	log.Printf("Server started at port %v", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

//Init Routes
func (a *AppHandler) initializeRoutes() {
	// all rate, create, update and delete routes
	a.Router.HandleFunc("/rates", a.getAllRates).Methods(http.MethodGet)
	a.Router.HandleFunc("/rate/{id:[0-9]+}", a.getRate).Methods(http.MethodGet)
	a.Router.HandleFunc("/rate", a.createRate).Methods(http.MethodPost)
	a.Router.HandleFunc("/rate/{id:[0-9]+}", a.updateRate).Methods(http.MethodPut)
	a.Router.HandleFunc("/rate/{id:[0-9]+}", a.deleteRateById).Methods(http.MethodDelete)
	a.Router.HandleFunc("/rate", a.deleteRateByFromTo).Methods(http.MethodDelete)

	// daily exchange rate data
	a.Router.HandleFunc("/rate/daily", a.dailyRate).Methods(http.MethodPost)
	a.Router.HandleFunc("/rate/track", a.trackRate).Methods(http.MethodPost)

}

//get rate by id
func (a *AppHandler) getRate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Id specified")
		return
	}

	rate := Rate{Id: id}
	if err := rate.GetRate(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Rate not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, rate)
}

//Get all rates, short of
//Set count and to start for pagination
func (a *AppHandler) getAllRates(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 100 || count < 1 {
		count = 100
	}
	if start < 0 {
		start = 0
	}

	rates, err := GetAllRates(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, rates)
}

// Create new Rate
// Expected post params
// {
// 	"trom" => "USD",
// 	"to" => "IDR"
// }
func (a *AppHandler) createRate(w http.ResponseWriter, r *http.Request) {
	var rate Rate
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rate); err != nil {
		respondWithError(w, http.StatusBadRequest, InvalidRequest)
		return
	}
	defer r.Body.Close()

	if err := rate.CreateRate(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, rate)
}

func (a *AppHandler) updateRate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid rate Id")
		return
	}

	var rate Rate
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rate); err != nil {
		respondWithError(w, http.StatusBadRequest, InvalidRequest)
		return
	}
	defer r.Body.Close()
	rate.Id = id

	if err := rate.UpdateRate(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, rate)
}

func (a *AppHandler) deleteRateByFromTo(w http.ResponseWriter, r *http.Request) {
	var rate Rate
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rate); err != nil {
		respondWithError(w, http.StatusBadRequest, InvalidRequest)
		return
	}
	defer r.Body.Close()

	if err := rate.DeleteRate(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *AppHandler) deleteRateById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid rate Id")
		return
	}

	rate := Rate{Id: id}
	if err := rate.DeleteRateById(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *AppHandler) dailyRate(w http.ResponseWriter, r *http.Request) {
	var rate RateData

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rate); err != nil {
		respondWithError(w, http.StatusBadRequest, InvalidRequest)
		return
	}

	defer r.Body.Close()

	fmt.Println(rate)

	// check  date format against RFC3339, "2006-01-02T15:04:05Z07:00"
	_, err := time.Parse("2006-01-02", rate.Date)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, err = rate.Rate.Float64()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = rate.DailyRate(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, rate)
}

func (a *AppHandler) trackRate(w http.ResponseWriter, r *http.Request) {
	var d DateStruct

	decoder := json.NewDecoder(r.Body)

	// check the payload send by client
	if err := decoder.Decode(&d); err != nil {
		respondWithError(w, http.StatusBadRequest, InvalidRequest)
		return
	}
	defer r.Body.Close()

	fmt.Println(d)

	// check  date format against RFC3339, "2006-01-02T15:04:05Z07:00"
	_, err := time.Parse("2006-01-02", d.Date)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	track, err := TrackRates(a.DB, d.Date)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid date specified")
		return
	}

	respondWithJSON(w, http.StatusOK, track)
}



// marshal an interface to JSON, set headers, and set the status code given.
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
