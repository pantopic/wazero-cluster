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

// ShardFind returns the first shard with a name less than or equal to the supplied name
// This is useful for exact matches but also for partitioning. Example:
//
//	shard := cluster.ShardFind(`data-`+sha1(key))
//	val, res, err := shard.Read(key, true)
func ShardFind(name string) Shard {
	return Shard{ID: 1, Name: name}
}
