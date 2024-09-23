package ratelimiter

type Ratelimiter interface {
	SetRules(ruleStr string) error
	Inc(key string) error
	Allow(key string) (bool, error)
}

func New() *MasterManager {
	return &MasterManager{
		ratelimiters: make(map[string]*MasterRatelimiter),
	}
}

func NewRemote(addr string) *SlaveManager {
	return &SlaveManager{
		remoteAddr: addr,
	}
}
