package wazero_cluster

import (
	"context"
	"fmt"
)

// Name is the name of this host module.
const Name = "pantopic/wazero-cluster"

var (
	ctxKeyNamespace = Name + `/namespace`
	ctxKeyResource  = Name + `/resource`
)

type resolver struct {
	namespace string
	resource  string
}

func (r resolver) ContextCopy(dst, src context.Context) context.Context {
	dst = context.WithValue(dst, ctxKeyNamespace, r.namespace)
	dst = context.WithValue(dst, ctxKeyResource, r.resource)
	return dst
}

func NewResolver(namespace, resource string) resolver {
	return resolver{
		namespace: namespace,
		resource:  resource,
	}
}

func NamespaceFrom(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyNamespace).(string); ok {
		return v
	}
	panic(fmt.Sprintf("Namespace missing from context"))
}

func ResourceFrom(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyResource).(string); ok {
		return v
	}
	panic(fmt.Sprintf("Namespace missing from context"))
}

func ResolveFrom(ctx context.Context) (string, string) {
	return NamespaceFrom(ctx), ResourceFrom(ctx)
}
