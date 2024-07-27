package core

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

type Env string

const (
	Development Env = "development"
	Production  Env = "production"
	Local       Env = "local"

	MaxBackoffSec = 5

	BucketMinutes = 60
	TXCacheTime   = time.Minute * 30
)

const (
	UnknownType = "unknown"
)

type CtxKey uint8

const (
	Logger CtxKey = iota
	Clients
	State
)

type ClientConfig struct {
	L1RpcEndpoint string
	PollInterval  int
	NumOfRetries  int
	StartHeight   *big.Int
	EndHeight     *big.Int
}

type Event struct {
	Timestamp time.Time

	Value  *types.Transaction
	Source string
}

type TopicType uint8

const (
	BlockHeader TopicType = iota + 1
	Log
)

func (rt TopicType) String() string {
	switch rt {
	case BlockHeader:
		return "block_header"

	case Log:
		return "log"
	}

	return UnknownType
}

type ProcessType uint8

const (
	Read ProcessType = iota + 1
	Subscribe
)

type DataTopic struct {
	DataType    TopicType
	ProcessType ProcessType
	Constructor interface{}
}

func (ct ProcessType) String() string {
	switch ct {
	case Read:
		return "reader"

	case Subscribe:
		return "subscriber"
	}

	return UnknownType
}

const (
	UnknownCompType = "unknown process type %s provided"
	CouldNotCastErr = "could not cast process initializer function to %s constructor type"
)
