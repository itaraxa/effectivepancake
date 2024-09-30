package agentbuffer

import (
	"sync"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

type AgentBuffer struct {
	Data []*models.Metrics
	Size uint64
	mu   sync.Mutex
}

func Init() *AgentBuffer {
	return &AgentBuffer{Data: []*models.Metrics{}, Size: 0, mu: sync.Mutex{}}
}

func (ab *AgentBuffer) AddBack(data *models.Metrics) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	ab.Data = append(ab.Data, data)
	ab.Size++
}

func (ab *AgentBuffer) GetForward() (*models.Metrics, error) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.Size == 0 {
		return nil, errors.ErrBufferIsEmpty
	}
	if ab.Size == 1 {
		t := ab.Data[0]
		ab.Size--
		return t, nil
	}
	t := ab.Data[0]
	ab.Data = ab.Data[1:]
	ab.Size--
	return t, nil
}
