package cache

import (
	"bytes"
	"encoding/gob"
	"log"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

func (c *Cache) fetchEntry(itemName string, entity interface{},
	dataRetrieveFunction func() (interface{}, types.Error)) types.Error {

	if c == nil || c.freecache == nil {
		return types.NewDatabaseError(nil)
	}

	// Get cachekey based upon object type & name of item we retrieve
	cachekey, entityType := getCacheKeyAndType(itemName, entity)
	c.logger.Debug("fetchEntry", zap.String("cachekey", string(cachekey)))

	if cached, err := c.freecache.Get(cachekey); err == nil && cached != nil {
		if err = gob.NewDecoder(bytes.NewBuffer(cached)).Decode(entity); err != nil {
			c.logger.Error("cache decode failed", zap.Error(err))
			return types.NewDatabaseError(err)
		}
		c.metrics.EntityCacheHit(entityType)
		return nil
	}

	// Cache miss
	c.metrics.EntityCacheMiss(entityType)

	// Retrieve request data from database layer
	data, err := dataRetrieveFunction()
	if err != nil {
		log.Print("DI: data base read failed!")
		return err
	}
	encodedData, e := encode(data)
	if e != nil {
		c.logger.Error("cache encoding failed", zap.Error(err))
		return types.NewDatabaseError(e)
	}
	// Store in cache
	if err := c.freecache.Set(cachekey, encodedData, c.config.TTL); err != nil {
		c.logger.Error("cache store failed", zap.Error(err))
	}
	// We decode the encoded data back into native type(!)
	// We do this do provide the retrieve database back to the calling function
	_ = gob.NewDecoder(bytes.NewBuffer(encodedData)).Decode(entity)
	return nil
}

func (c *Cache) deleteEntry(itemName string, entity interface{}) {

	// Get cachekey based upon object type and item's name to delete
	cachekey, _ := getCacheKeyAndType(itemName, entity)

	_ = c.freecache.Del(cachekey)
}

func decode(encodedData []byte, data interface{}) error {

	return gob.NewDecoder(bytes.NewBuffer(encodedData)).Decode(data)
}

func encode(data interface{}) ([]byte, error) {

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getCacheKeyAndType(entityValue string, entity interface{}) (cacheKey []byte, typeName string) {

	const (
		allUsers = "all_users"
		allRoles = "all_roles"
	)

	var prefix string
	switch entity.(type) {
	case types.User:
		prefix = db.EntityTypeUser
		typeName = db.EntityTypeUser
	case *types.User:
		prefix = db.EntityTypeUser
		typeName = db.EntityTypeUser

	case types.Users:
		prefix = allUsers
		typeName = db.EntityTypeUser + "s"
	case *types.Users:
		prefix = allUsers
		typeName = db.EntityTypeUser + "s"

	case types.Role:
		prefix = db.EntityTypeRole
		typeName = db.EntityTypeRole
	case *types.Role:
		prefix = db.EntityTypeRole
		typeName = db.EntityTypeRole

	case types.Roles:
		prefix = allRoles
		typeName = db.EntityTypeRole + "s"
	case *types.Roles:
		prefix = allRoles
		typeName = db.EntityTypeRole + "s"
	}
	return []byte(prefix + "_" + entityValue), typeName
}
