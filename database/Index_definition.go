package database

import (
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IndexDefinition struct {
	Keys    interface{}
	Options *options.IndexOptions
}
