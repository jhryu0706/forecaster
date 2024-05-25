package db

import (
	"fmt"
	"log"
	"proj3/data/requestdata"

	"github.com/coocood/qbs"
	_ "github.com/mattn/go-sqlite3"
	"gonum.org/v1/gonum/stat/distuv"
)

type MovementData struct {
	Symbol  string `qbs:"unique,pk"`
	Mean    float64
	Stdv    float64
	Updated string
}

func (m MovementData) String() string {
	return fmt.Sprintf("[Symbol] %s [Mean] %.4f [Stdv] %.4f [Updated] %s", m.Symbol, m.Mean, m.Stdv, m.Updated)
}

func CreateTable() error {
	migration, err := qbs.GetMigration()
	if err != nil {
		log.Println(err)
	}
	defer migration.Close()
	return migration.CreateTableIfNotExists(new(MovementData))
}

// Inserting new data into the db
func UpdateDB(symbol string) *MovementData {
	requestdata.LoadData(symbol)
	mean, stdv, updated := requestdata.GetDailyMovement(symbol)
	newdata := MovementData{
		Symbol:  symbol,
		Mean:    mean,
		Stdv:    stdv,
		Updated: updated,
	}
	q, err := qbs.GetQbs()
	if err != nil {
		panic(err)
	}
	defer q.Close()
	_, err = q.Save(&newdata)
	log.Println("data saved in table")
	if err != nil {
		panic(err)
	}
	return &newdata
}

// Checks if symbol exists, and whether the information is up to date, two cases to handle -> stale data exists, or data does not exist. Always return boolean value, if stale data exists, simply remove entry
func CheckDB(symbol string) *distuv.Normal {
	q, err := qbs.GetQbs()
	if err != nil {
		panic(err)
	}
	defer q.Close()
	data := new(MovementData)
	err = q.WhereEqual("Symbol", symbol).Find(data)
	//this has to go in production code
	if err != nil {
		log.Println(err)
	}
	// if data.Symbol == "" || data.Updated < time.Now().Format("2006-01-02") {
	// 	if data.Symbol != "" {
	// 		// means the data exists, but was stale
	// 		log.Println(symbol, "stale")
	// 		del := new(MovementData)
	// 		del.Symbol = symbol
	// 		q.Delete(del)
	// 	}
	// 	// log.Println("returning with update")
	// 	updated := UpdateDB(symbol)
	// 	return &distuv.Normal{Mu: updated.Mean, Sigma: updated.Stdv}
	// }
	// log.Println("returning straight from table")
	return &distuv.Normal{Mu: data.Mean, Sigma: data.Stdv}
}

func GetAllDB() {
	q, _ := qbs.GetQbs()
	var all []*MovementData
	q.FindAll(&all)
	for i, val := range all {
		log.Println(i, val)
	}
}
