package service

import "sync"

// RatingStore is an interface to store laptop ratings
type RatingStore interface {
	// Add function adds a laptop score to the store and returns the rating
	Add(laptopId string, score float64) (*Rating, error)
}

// Rating has the laptop's scores info
type Rating struct {
	Count uint32
	Sum   float64
}

// InMemoryRatingStore stores laptop ratings in the memory
type InMemoryRatingStore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

// NewInMemoryStore returns a new store for laptop ratings store
func NewInMemoryStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

// Add a new Laptop to store and return its rating
func (store *InMemoryRatingStore) Add(laptopId string, score float64) (*Rating, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	rating := store.rating[laptopId]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.rating[laptopId] = rating
	return rating, nil
}
