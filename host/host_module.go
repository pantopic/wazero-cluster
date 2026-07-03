package wazero_cluster

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/logbn/zongzi"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Name is the name of this host module.
const Name = "pantopic/wazero-cluster-registry"

var (
	ctxKeyMeta  = Name + `/meta`
	ctxKeyAgent = Name + `/agent`
)

type meta struct {
	ptrShardID uint32
	ptrVal     uint32
	ptrDataMax uint32
	ptrDataLen uint32
	ptrData    uint32
	ptrErrMax  uint32
	ptrErrLen  uint32
	ptrErr     uint32
}

type hostModule struct {
	sync.RWMutex

	module api.Module

	resolveNamespace func(context.Context) string
	resolveResource  func(context.Context) string
}

func New(opts ...Option) *hostModule {
	h := &hostModule{
		resolveNamespace: func(ctx context.Context) string {
			return `default`
		},
		resolveResource: func(ctx context.Context) string {
			return `default`
		},
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *hostModule) Name() string {
	return Name
}

// Register instantiates the host module, making it available to all module instances in this runtime
func (h *hostModule) Register(ctx context.Context, r wazero.Runtime) (err error) {
	builder := r.NewHostModuleBuilder(Name)
	register := func(name string, fn func(ctx context.Context, m api.Module, stack []uint64)) {
		builder = builder.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(fn), nil, nil).Export(name)
	}
	for name, fn := range map[string]any{
		"__cluster_registry_shard_find": func(ctx context.Context, agent *zongzi.Agent, namespace, resourceName, shardName string) (id uint64) {
			agent.StateLocal(func(s *zongzi.State) {
				shard, _ := s.ShardFindByName(fmt.Sprintf(`%s.%s.%s`, namespace, resourceName, shardName))
				id = shard.ID
			})
			return
		},
	} {
		switch fn := fn.(type) {
		case func(ctx context.Context, agent *zongzi.Agent, namespace, resourceID, name string) (id uint64):
			register(name, func(ctx context.Context, m api.Module, stack []uint64) {
				meta := h.meta(ctx)
				val := fn(ctx,
					h.agent(ctx),
					h.resolveNamespace(ctx),
					h.resolveResource(ctx),
					string(getData(m, meta)))
				setShardID(m, meta, val)
			})
		default:
			log.Panicf("Method signature implementation missing: %#v", fn)
		}
	}
	h.module, err = builder.Instantiate(ctx)
	return
}

// InitContext retrieves the meta page from the wasm module
func (h *hostModule) InitContext(ctx context.Context, m api.Module) (context.Context, error) {
	stack, err := m.ExportedFunction(`__cluster_registry`).Call(ctx)
	if err != nil {
		return ctx, err
	}
	meta := &meta{}
	ptr := uint32(stack[0])
	for i, v := range []*uint32{
		&meta.ptrShardID,
		&meta.ptrVal,
		&meta.ptrDataMax,
		&meta.ptrDataLen,
		&meta.ptrData,
		&meta.ptrErrMax,
		&meta.ptrErrLen,
		&meta.ptrErr,
	} {
		*v = readUint32(m, ptr+uint32(4*i))
	}
	return context.WithValue(ctx, ctxKeyMeta, meta), nil
}

// ContextCopy populates dst context with the meta page from src context.
func (h *hostModule) ContextCopy(dst, src context.Context) context.Context {
	dst = context.WithValue(dst, ctxKeyMeta, get[*meta](src, ctxKeyMeta))
	dst = context.WithValue(dst, ctxKeyAgent, h.agent(src))
	return dst
}

func (h *hostModule) meta(ctx context.Context) *meta {
	return get[*meta](ctx, ctxKeyMeta)
}

func (h *hostModule) agent(ctx context.Context) *zongzi.Agent {
	return get[*zongzi.Agent](ctx, ctxKeyAgent)
}

func get[T any](ctx context.Context, key string) T {
	v := ctx.Value(key)
	if v == nil {
		log.Panicf("Context item missing %s", key)
	}
	return v.(T)
}

func getData(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrData, meta.ptrDataLen, meta.ptrDataMax)
}

func setShardID(m api.Module, meta *meta, val uint64) {
	writeUint64(m, meta.ptrVal, val)
}

func read(m api.Module, ptrData, ptrLen, ptrMax uint32) (buf []byte) {
	buf, ok := m.Memory().Read(ptrData, readUint32(m, ptrMax))
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range", ptrData, ptrLen)
	}
	return buf[:readUint32(m, ptrLen)]
}

func readUint32(m api.Module, ptr uint32) (val uint32) {
	val, ok := m.Memory().ReadUint32Le(ptr)
	if !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
	return
}

func readUint64(m api.Module, ptr uint32) (val uint64) {
	val, ok := m.Memory().ReadUint64Le(ptr)
	if !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
	return
}

func writeUint32(m api.Module, ptr uint32, val uint32) {
	if ok := m.Memory().WriteUint32Le(ptr, val); !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
}

func writeUint64(m api.Module, ptr uint32, val uint64) {
	if ok := m.Memory().WriteUint64Le(ptr, val); !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
}
