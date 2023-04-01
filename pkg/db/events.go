package db

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
)

type Event struct {
	// InstallationID is the id of the installation
	InstallationID int64
	// RepoOwner is the owner of the repository
	RepoOwner string
	// RepoName is the name of the repository
	RepoName string
	// Id is the id of the issue or pull request
	ID int
	// LabelsCategory is the list of labels
	LabelsCategory []string
	// LabelsType is the list of labels
	LabelsType []string
}

// GetKey returns the key of the event.
func (e *Event) GetKey() string {
	return fmt.Sprintf("%d/%s/%s/%d", e.InstallationID, e.RepoOwner, e.RepoName, e.ID)
}

// Marshal returns the event as a string.
func (e *Event) Marshal() []byte {
	vJ, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return vJ
}

type EventDB struct {
	Name
}

// EventDBNew returns a new EventDB.
func EventDBNew(db Name) *EventDB {
	x := EventDB{db}
	return &x
}

// AddEvent add an event to the database.
func (db *EventDB) AddEvent(event Event) error {
	return DataBase.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(db.Name))
		return b.Put([]byte(event.GetKey()), event.Marshal())
	})
}

// GetEvent returns an event from the database.
func (db *EventDB) GetEvent(key string) (*Event, error) {
	var event Event
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(db.Name))
		v := b.Get([]byte(key))
		if v == nil {
			return fmt.Errorf("event not found")
		}
		return json.Unmarshal(v, &event)
	})
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// DeleteEvent deletes an event from the database.
func (db *EventDB) DeleteEvent(key string) error {
	return DataBase.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(db.Name))
		return b.Delete([]byte(key))
	})
}

// GetEvents returns all events from the database.
func (db *EventDB) GetEvents() ([]Event, error) {
	var events []Event
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(db.Name))
		return b.ForEach(func(k, v []byte) error {
			var event Event
			err := json.Unmarshal(v, &event)
			if err != nil {
				return err
			}
			events = append(events, event)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GetEventsForInstallation returns all events for a given installation.
func (db *EventDB) GetEventsForInstallation(installationID int64) ([]Event, error) {
	var events []Event
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(db.Name))
		return b.ForEach(func(k, v []byte) error {
			var event Event
			err := json.Unmarshal(v, &event)
			if err != nil {
				return err
			}
			if event.InstallationID == installationID {
				events = append(events, event)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GetEventsForRepo returns all events for a given repository.
func (db *EventDB) GetEventsForRepo(installationID int64, repoOwner, repoName string) ([]Event, error) {
	var events []Event
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(db.Name))
		return b.ForEach(func(k, v []byte) error {
			var event Event
			err := json.Unmarshal(v, &event)
			if err != nil {
				return err
			}
			if event.InstallationID == installationID && event.RepoOwner == repoOwner && event.RepoName == repoName {
				events = append(events, event)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GetEventsForIssue returns all events for a given issue.
func (db *EventDB) GetEventsForIssue(installationID int64, repoOwner, repoName string, id int) ([]Event, error) {
	var events []Event
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(db.Name))
		return b.ForEach(func(k, v []byte) error {
			var event Event
			err := json.Unmarshal(v, &event)
			if err != nil {
				return err
			}
			if event.InstallationID == installationID && event.RepoOwner == repoOwner && event.RepoName == repoName && event.ID == id {
				events = append(events, event)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

// KeyExist returns true if the key exists.
func (db *EventDB) KeyExist(key string) (bool, error) {
	var exist bool
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(db.Name))
		v := b.Get([]byte(key))
		if v != nil {
			exist = true
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return exist, nil
}
