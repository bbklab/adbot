package scheduler

import (
	"time"

	"github.com/bbklab/adbot/store"
)

// DBPing ping the db store and return the time cost and any errors if met
func DBPing() (time.Duration, error) {
	startAt := time.Now()
	err := store.DB().Ping()
	cost := time.Since(startAt)
	return cost, err
}
