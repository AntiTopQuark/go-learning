package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// SelectMode 代表不同的负载均衡策略，简单起见，GeeRPC 仅实现 Random 和 RoundRobin 两种策略。
type SelectMode int

const (
	RandomSelect     SelectMode = iota // 0是随机策略
	RoundRobinSelect                   // 1是轮询策略
)

type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error) //Get(mode SelectMode) 根据负载均衡策略，选择一个服务实例
	GetAll() ([]string, error)
}

// 服务列表由手工维护的服务发现的结构体
type MultiServersDiscovery struct {
	r       *rand.Rand //产生随机数的实例
	mu      sync.Mutex
	servers []string
	index   int //index 记录 Round Robin 算法已经轮询到的位置，为了避免每次从 0 开始，初始化时随机设定一个值。

}

func (receiver *MultiServersDiscovery) Update(servers []string) error {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	receiver.servers = servers
	return nil
}

func (receiver *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	length := len(receiver.servers)
	if length == 0 {
		return "", errors.New("rpc 服务发现：没有可用服务")
	}
	switch mode {
	case RandomSelect:
		return receiver.servers[receiver.r.Intn(length)], nil
	case RoundRobinSelect:
		s := receiver.servers[receiver.index%length]
		receiver.index = (receiver.index + 1) % length
		return s, nil
	default:
		return "", errors.New("rpc 服务发现：不支持该模式")
	}
}

func (receiver *MultiServersDiscovery) GetAll() ([]string, error) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	servers := make([]string, len(receiver.servers))
	copy(servers, receiver.servers)
	return servers, nil
}

var _ Discovery = (*MultiServersDiscovery)(nil)

func (receiver *MultiServersDiscovery) Refresh() error {
	return nil
}

// NewMultiServerDiscovery creates a MultiServersDiscovery instance
func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}
