package registry

import (
	"fmt"

	"github.com/denzelpenzel/magic-chain/internal/core"
)

const (
	noEntryErr = "could not find data topic type %v"
)

type Registry struct {
	topics map[core.TopicType]*core.DataTopic
}

func New() *Registry {
	topics := map[core.TopicType]*core.DataTopic{
		core.BlockHeader: {
			DataType:    core.BlockHeader,
			ProcessType: core.Subscribe,
			Constructor: NewHeaderTraversal,
		},
		core.Log: {},
	}

	return &Registry{topics}
}

func (r *Registry) GetDataTopic(tt core.TopicType) (*core.DataTopic, error) {
	if _, exists := r.topics[tt]; !exists {
		return nil, fmt.Errorf(noEntryErr, tt)
	}
	return r.topics[tt], nil
}
