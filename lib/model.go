//model.go
package lib

import (
	"database/sql"
	"fmt"
	"encoding/json"
	"errors"
)

//Struct to hold our Rate
type Rate struct {
	Id    int     `json:"id"`
	From  string  `json:"from"`
	To    string  `json:"to"`
}

//Struct to hold our Rate Data
type RateData struct {
	Date  string      `json:"date"`
	From  string      `json:"from"`
	To    string      `json:"to"`
	Rate  json.Number `json:"rate"`
}

// Custom struct to hold joined query of Rate Track
type RateTrack struct {
	RateId  int     `json:"rate_id"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	// by requirement, we need to pass "insufficient data" if there is not enough 7 days data
	Rate    string  `json:"rate"`
	WeekAvg string  `json:"7_day_avg"`
}

type DateStruct struct {
	Date string `json:"date"`
}


// query to get all Rate
func GetAllRates(db *sql.DB, start, count int) ([]Rate, error) {
	statement := fmt.Sprintf("SELECT id, `from`, `to` FROM rate LIMIT %d OFFSET %d", count, start)
	rows, err := db.Query(statement)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rates := []Rate{}

	for rows.Next() {
		var r Rate
		if err := rows.Scan(&r.Id, &r.From, &r.To); err != nil {
			return nil, err
		}
		rates = append(rates, r)
	}
	return rates, nil
}

// query rate by given id
func (r *Rate) GetRate(db *sql.DB) error {
	statement := fmt.Sprintf("SELECT id, `from`, `to` FROM rate WHERE id=%d", r.Id)
	return db.QueryRow(statement).Scan(&r.Id, &r.From, &r.To)
}
// query create new rate
func (r *Rate) CreateRate(db *sql.DB) error {
	statement := fmt.Sprintf("INSERT INTO rate(`from`, `to`) VALUES('%s', '%s')", r.From, r.To)
	_, err := db.Exec(statement)

	if err != nil {
		return err
	}

	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&r.Id)

	if err != nil {
		return err
	}

	return nil
}

// query for update rate by given id
func (r *Rate) UpdateRate(db *sql.DB) error {
	statement := fmt.Sprintf("UPDATE rate SET `from`='%s', `to`='%s' WHERE id=%d", r.From, r.To, r.Id)
	_, err := db.Exec(statement)
	return err
}

// query for deleting rate and it's data with "from" and "to" param
func (r *Rate) DeleteRate(db *sql.DB) error {
	statement := fmt.Sprintf("DELETE FROM rate WHERE `from`='%s' AND `to`='%s'", r.From, r.To)
	_, err := db.Exec(statement)

	return err
}

func (r *Rate) DeleteRateById(db *sql.DB) error {
	//var err error
	//var statement string

	// if FOREIGN KEYS is ON DELETE RESTRICT, delete the child first
	// if FOREIGN KEYS is ON DELETE CASCADE, the child will automatically deleted, if parent row is deleted
	// so we don't need this, because foreign key is cascade
	// statement = fmt.Sprintf("DELETE FROM rate_data WHERE rate_id=%d", r.Id)
	// _, err = db.Exec(statement)

	statement := fmt.Sprintf("DELETE FROM rate WHERE id=%d", r.Id)
	_, err := db.Exec(statement)

	return err
}


func TrackRates(db *sql.DB, date string) ([]RateTrack, error) {
	statement := fmt.Sprintf(`
	SELECT id, r.from, r.to, IFNULL(rate, 'insufficient data') AS rate, IFNULL(week_avg, 'insufficient data') AS 7_day_avg
	FROM rate r
	LEFT JOIN
	(
		SELECT DISTINCT t1.rate_id,
		IF(COUNT(t1.rate_id) >= 7, IFNULL(t1.rate, t2.rate), NULL) AS rate,
		IF(COUNT(t2.rate_id) >= 7, AVG(t2.rate), NULL) AS week_avg
		FROM rate_data t1
		INNER JOIN rate_data t2 ON t1.rate_id = t2.rate_id AND DATEDIFF(t1.date, t2.date) BETWEEN 0 AND 6
		WHERE (t1.date = '%s')
		GROUP BY t1.rate_id, t1.date
	) z ON z.rate_id = r.id`, date)

	fmt.Println(statement)

	rows, err := db.Query(statement)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := []RateTrack{}

	for rows.Next() {
		var d RateTrack
		if err := rows.Scan(&d.RateId, &d.From, &d.To, &d.Rate, &d.WeekAvg); err != nil {
			return nil, err
		}
		data = append(data, d)
	}
	return data, nil
}

func (r *RateData) DailyRate(db *sql.DB) error {

	var statement string
	var err error

	value, err := r.Rate.Float64()

	// this query will insert new rate data, or update the rate if it's already exist
	statement = fmt.Sprintf(`
	INSERT INTO rate_data(rate_id, date, rate)
	SELECT r.id, '%s', '%f'
	FROM rate r WHERE r.from = '%s' AND r.to = '%s'
	ON DUPLICATE KEY UPDATE rate = VALUES(rate)`, r.Date, value, r.From, r.To)
	_, err = db.Exec(statement)

	if err != nil {
		return err
	}

	var id int
	// id will be 1 if new row inserted, and 2 if row updated
	err = db.QueryRow("SELECT ROW_COUNT()").Scan(&id)

	//fmt.Println("row affected", id)

	if err != nil {
		return err
	}

	if (id <= 0) {
		return errors.New( fmt.Sprintf("Invalid rate specified / %d row updated/inserted", id) )
	}

	return nil
}