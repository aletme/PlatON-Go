package statsdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/rlp"

	leveldbError "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/PlatONnetwork/PlatON-Go/log"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DBPath = "statsdb"
)

var (
	dbpath string

	dbInstance *StatsDB

	instance sync.Mutex

	levelDBcache   = int(16)
	levelDBhandles = int(500)

	logger = log.Root().New("package", "statsdb")

	//ErrNotFound when db not found
	ErrNotFound = errors.New("statsDB: not found")
)

type StatsDB struct {
	path    string
	levelDB *leveldb.DB
	closed  bool
	dbError error
}

func (db *StatsDB) WriteExeBlockData(blockNumber *big.Int, data *common.ExeBlockData) {
	if data == nil {
		return
	}

	jsonBytes, _ := json.Marshal(data)
	log.Info("WriteExeBlockData", "blockNumber", blockNumber, "data", string(jsonBytes))

	encoded := common.MustRlpEncode(data)
	if err := db.PutLevelDB(blockNumber.Bytes(), encoded); err != nil {
		log.Crit("Failed to write ExeBlockData", "blockNumber", blockNumber, "data", common.Bytes2Hex(encoded), "err", err)
	}
}

func (db *StatsDB) ReadExeBlockData(blockNumber *big.Int) *common.ExeBlockData {
	bytes, _ := db.GetLevelDB(blockNumber.Bytes())
	if len(bytes) == 0 {
		return nil
	}
	var data common.ExeBlockData
	if err := rlp.DecodeBytes(bytes, &data); err != nil {
		log.Crit("Failed to read ExeBlockData", "blockNumber", blockNumber, "data", common.Bytes2Hex(bytes), "err", err)
		return nil
	}
	return &data
}

func (db *StatsDB) DeleteExeBlockData(blockNumber *big.Int) {
	if err := db.DeleteLevelDB(blockNumber.Bytes()); err != nil {
		log.Crit("Failed to delete ExeBlockData", "blockNumber", blockNumber, "err", err)
	}
}

func (db *StatsDB) PutLevelDB(key, value []byte) error {
	err := db.levelDB.Put(key, value, nil)
	if err != nil {
		return err
	}
	return nil
}

func (db *StatsDB) DeleteLevelDB(key []byte) error {
	err := db.levelDB.Delete(key, nil)
	if err != nil {
		return err
	}
	return nil
}

func (db *StatsDB) GetLevelDB(key []byte) ([]byte, error) {
	if v, err := db.levelDB.Get(key, nil); err == nil {
		return v, nil
	} else if err != leveldb.ErrNotFound {
		return nil, err
	} else {
		return nil, ErrNotFound
	}
}

func (db *StatsDB) HasLevelDB(key []byte) (bool, error) {
	_, err := db.GetLevelDB(key)
	if err == nil {
		return true, nil
	} else if err == ErrNotFound {
		return true, ErrNotFound
	} else {
		return false, err
	}
}

func (db *StatsDB) Close() error {
	if db.levelDB != nil {
		if err := db.levelDB.Close(); err != nil {
			return fmt.Errorf("[statsDB]close level db fail:%v", err)
		}
	}
	db.closed = true
	return nil
}

func SetDBPath(path string) {
	dbpath = path
	logger.Info("set path", "path", dbpath)
}

func SetDBOptions(cache int, handles int) {
	levelDBcache = cache
	levelDBhandles = handles
}

func Instance() *StatsDB {
	instance.Lock()
	defer instance.Unlock()
	if dbInstance == nil || dbInstance.closed {
		logger.Debug("dbInstance is nil", "path", dbpath)
		if dbInstance == nil {
			dbInstance = new(StatsDB)
		}
		if levelDB, err := openLevelDB(levelDBcache, levelDBhandles); err != nil {
			logger.Error("init db fail", "err", err)
			panic(err)
		} else {
			dbInstance.levelDB = levelDB
		}
	}
	return dbInstance
}

func openLevelDB(cache int, handles int) (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(dbpath, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if err != nil {
		if _, corrupted := err.(*leveldbError.ErrCorrupted); corrupted {
			db, err = leveldb.RecoverFile(dbpath, nil)
			if err != nil {
				return nil, fmt.Errorf("[StatsDB.recover]RecoverFile levelDB fail:%v", err)
			}
		} else {
			return nil, err
		}
	}
	return db, nil
}
