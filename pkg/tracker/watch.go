package tracker

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/FrangipaneTeam/crown/pkg/db"
	"go.etcd.io/bbolt"
)

const (
	intervalLoopWatch = 30 * time.Second
)

// Watch is watcher of the issue/pr
// It is responsible for scanning the issue/pr
// and updating the database
func Watch() {

	for {
		logger.Debug().Msg("Start watching")
		// Get all the repos
		err := db.DataBase.Update(func(tx *bbolt.Tx) error {
			// Assume bucket exists and has keys
			c := tx.Bucket([]byte(db.TrackDB().Bucket())).Cursor()

			prefix := []byte("issues/")
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				// For each issue
				var x trackBase
				json.Unmarshal(v, &x)

				issue := &TrackIssue{
					base: x,
				}

				if issue.ScanIsNecessary() {
					logger.Debug().Msgf("Scan necessary for issue %s", k)
					if err := issue.Scan(); err != nil {
						logger.Error().Err(err).Msgf("Error while scanning issue %s", k)
						continue
					}

					issueJ, err := json.Marshal(issue.base)
					if err != nil {
						logger.Error().Err(err).Msgf("Error while marshaling issue %s", k)
						continue
					}
					if err := c.Bucket().Put(k, issueJ); err != nil {
						logger.Error().Err(err).Msgf("Error while saving issue %s", k)
						continue
					}
				}
			}

			return nil
		})

		if err != nil {
			logger.Error().Err(err).Msg("Error while watching")
		}

		logger.Debug().Msg("End watching waiting for next loop")
		time.Sleep(intervalLoopWatch)
	}
}
