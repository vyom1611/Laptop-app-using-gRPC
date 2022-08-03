package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"laptop-app-using-grpc/pb/pb"
	"log"

	"sync"
)

var ErrorAlreadyExists = errors.New("Error already exists")

type LaptopStore interface {
	//Saves the laptop to store
	Save(laptop *pb.Laptop) error
	//Find laptop by Id
	Find(id string) (*pb.Laptop, error)

	//Searching laptops and returning one by one with found function
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop)) error
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
	other, err := DeepCopy(laptop)
	if err != nil {
		return err
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

	return DeepCopy(laptop)
}

func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop)) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	//Looping through all the laptops in the store service
	for _, laptop := range store.data {
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Print("context is cancelled")
			return nil
		}

		// time.Sleep(time.Second)
		// log.Print("checking laptop id: ", laptop.GetId())
		if isQualified(filter, laptop) {
			other, err := DeepCopy(laptop)
			if err != nil {
				return err
			}

			found(other)
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetCpuCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 2^3 = 8
	case pb.Memory_KILOBYTE:
		return value << 13
	case pb.Memory_MEGABYTE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

//Deep Copy utility function
func DeepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return other, nil
}
