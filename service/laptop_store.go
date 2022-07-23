package service

import (
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"laptop-app-using-grpc/pb/pb"
	"sync"
)

var ErrorAlreadyExists = errors.New("Error already exists")

type LaptopStore interface {
	//Saves the laptop to store
	Save(laptop *pb.Laptop) error
	//Find laptop by Id
	Find(id string) (*pb.Laptop, error)
}

//Store laptop in-memory
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

//Returning new in memory laptop store
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

//Saving the laptop to store
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrorAlreadyExists
	}

	//deep copy
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("Cannot copy laptop data: %w", err)
	}

	store.data[other.Id] = other
	return nil
}

//Finding laptop by its Id on the store
func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	//find laptop from data store
	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	//Deep copying to another object
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return other, nil
}
