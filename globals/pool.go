package globals

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
)

var ByteArrayPool Pool
var ByteArrayIdGenerator IdGenertor

func init() {
	ByteArrayIdGenerator = NewIdGenertor()

	var err error
	total := MyConfig.DeviceMaxConcurrency
	size := MyConfig.BufferSize
	var byteArrayEntity = myByteArrayEntity{}
	ByteArrayPool, err = NewPool(total, size, reflect.TypeOf(&byteArrayEntity), GenByteArray)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}
}

type Pool interface {
	Take() (IEntity, error)
	Return(entity IEntity) error
	Total() int
	Used() int
	FreeCount() int
}

type IEntity interface {
	Id() int
}

type myPool struct {
	total       int
	etype       reflect.Type
	genEntity   func(int) IEntity
	container   chan IEntity
	idContainer map[int]bool
	mutex       sync.Mutex
}

type IByteArrayEntity interface {
	IEntity
	Bytes() []byte
}

type myByteArrayEntity struct {
	id    int
	bytes []byte
}

func GenByteArray(size int) IEntity {
	return &myByteArrayEntity{id: int(ByteArrayIdGenerator.GetUint32()), bytes: make([]byte, size)}
}

func (this *myByteArrayEntity) Id() int {
	return this.id
}

func (this *myByteArrayEntity) Bytes() []byte {
	return this.bytes
}

func NewPool(total int, size int, entityType reflect.Type, genEntity func(int) IEntity) (Pool, error) {
	if total == 0 {
		errMsg := fmt.Sprintf("The pool can not be initialized! (total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	container := make(chan IEntity, total)
	idContainer := make(map[int]bool)
	for i := 0; i < int(total); i++ {
		newEntity := genEntity(size)
		if entityType != reflect.TypeOf(newEntity) {
			errMsg := fmt.Sprintf("The type of result of function genEntity() is NOT %v, is:%v!\n", entityType, reflect.TypeOf(newEntity))
			return nil, errors.New(errMsg)
		}
		container <- newEntity
		idContainer[newEntity.Id()] = true
	}
	pool := &myPool{
		total:       total,
		etype:       entityType,
		genEntity:   genEntity,
		container:   container,
		idContainer: idContainer,
	}

	return pool, nil
}

func (pool *myPool) Take() (IEntity, error) {
	entity, ok := <-pool.container
	if !ok {
		return nil, errors.New("The inner container is invalid!")
	}

	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.idContainer[entity.Id()] = false
	return entity, nil
}

func (pool *myPool) Return(entity IEntity) error {
	if entity == nil {
		return errors.New("The returning entity is invalid!")
	}

	if pool.etype != reflect.TypeOf(entity) {
		errMsg := fmt.Sprintf("The type of returning entity is NOT %s!\n", pool.etype)
		return errors.New(errMsg)
	}

	entityId := entity.Id()
	casResult := pool.compareAndSetForIdContainer(entityId, false, true)
	if casResult == 1 {
		pool.container <- entity
		return nil
	} else if casResult == 0 {
		errMsg := fmt.Sprintf("The entity (id=%d) is already in the pool!\n", entity.Id())
		return errors.New(errMsg)
	} else {
		errMsg := fmt.Sprintf("The entity (id=%d) is illegal!\n", entity.Id())
		return errors.New(errMsg)
	}
}

func (pool *myPool) compareAndSetForIdContainer(entityId int, oldValue, newValue bool) int8 {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	v, ok := pool.idContainer[entityId]
	if !ok {
		return -1
	}
	if v != oldValue {
		return 0
	}
	pool.idContainer[entityId] = newValue
	return 1
}

func (pool *myPool) Total() int {
	return pool.total
}

func (pool *myPool) Used() int {
	return pool.total - int(len(pool.container))
}

func (pool *myPool) FreeCount() int {
	return int(len(pool.container))
}
