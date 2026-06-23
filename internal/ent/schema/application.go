package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/erewhile/iam/internal/ent/mixin"
)

// Application holds the schema definition for the Application entity.
type Application struct {
	ent.Schema
}

func (Application) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "application",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("application"),
		entsql.WithComments(true),
	}
}

// Fields of the Application.
func (Application) Fields() []ent.Field {
	return []ent.Field{
		field.String("client_id").
			Unique().
			MaxLen(36).
			NotEmpty(),

		field.String("client_secret").
			NotEmpty().
			MinLen(32).
			MaxLen(64),

		field.String("name").
			NotEmpty().
			MaxLen(64),

		field.JSON("redirect_uris", []string{}),
	}
}

func (Application) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Edges of the Application.
func (Application) Edges() []ent.Edge {
	return nil
}

func (Application) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("client_id").Unique(),
	}
}
