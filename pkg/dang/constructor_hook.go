package dang

import "context"

type constructorHookKey struct{}

// ConstructorHook lets an embedded runtime adapt a newly constructed object.
// It runs after field initialization and the explicit new body, if present.
// The scope is the constructor's closure, for resolving the object's types.
type ConstructorHook func(context.Context, ValueScope, *Object) (*Object, error)

// WithConstructorHook installs a hook for this evaluation and its local calls.
// It does not change constructors or scopes shared with other evaluations.
func WithConstructorHook(ctx context.Context, hook ConstructorHook) context.Context {
	return context.WithValue(ctx, constructorHookKey{}, hook)
}

func applyConstructorHook(ctx context.Context, scope ValueScope, obj *Object) (*Object, error) {
	if hook, ok := ctx.Value(constructorHookKey{}).(ConstructorHook); ok && hook != nil {
		return hook(ctx, scope, obj)
	}
	return obj, nil
}
