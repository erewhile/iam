package mixin

import (
	"context"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/erewhile/iam/internal/ent/db"
)

type softDeleteKey struct{}

func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

func isSoftDeleteSkipped(ctx context.Context) bool {
	skip, _ := ctx.Value(softDeleteKey{}).(bool)
	return skip
}

type SoftDeleteMixin struct {
	mixin.Schema
}

func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (d SoftDeleteMixin) fieldName() string {
	return d.Fields()[0].Descriptor().Name
}

func (d SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			if isSoftDeleteSkipped(ctx) {
				return nil
			}
			p, ok := q.(interface{ WhereP(...func(*sql.Selector)) })
			if !ok {
				return nil
			}
			p.WhereP(sql.FieldIsNull(d.fieldName()))
			return nil
		}),
	}
}

func (d SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if !m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
					return next.Mutate(ctx, m)
				}

				if isSoftDeleteSkipped(ctx) {
					return next.Mutate(ctx, m)
				}

				mx, ok := m.(interface {
					SetOp(ent.Op)
					Client() *db.Client
					SetDeletedAt(time.Time)
					WhereP(...func(*sql.Selector))
				})
				if !ok {
					return next.Mutate(ctx, m)
				}

				mx.WhereP(sql.FieldIsNull(d.fieldName()))
				mx.SetOp(ent.OpUpdate)
				mx.SetDeletedAt(time.Now())

				return mx.Client().Mutate(ctx, m)
			})
		},
	}
}
