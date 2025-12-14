package cluster

type Shard struct {
	ID   uint64
	Name string
}

func (s Shard) Read(cmd []byte, stale bool) (val uint64, res []byte, err error) {
	return
}

func (s Shard) Apply(cmd []byte) (val uint64, res []byte, err error) {
	return
}

func ShardFind(name string) Shard {
	return Shard{ID: 1, Name: name}
}
