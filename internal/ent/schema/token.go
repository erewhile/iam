package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/erewhile/iam/internal/ent/mixin"
	"github.com/erewhile/iam/internal/model"
	"github.com/google/uuid"
)

// Token holds the schema definition for the Token entity.
type Token struct {
	ent.Schema
}

func (Token) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "token",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("token"),
		entsql.WithComments(true),
	}
}

// Fields of the Token.
func (Token) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),

		field.UUID("jti", uuid.UUID{}).
			Immutable().
			Unique(),

		field.UUID("session_id", uuid.UUID{}),

		field.String("cookie_id").
			MaxLen(32).
			Optional(),

		field.Int("application_id").
			Optional().
			Nillable(),

		field.Uint8("type").
			GoType(model.TokenType(0)).
			Default(uint8(model.TokenTypeAccess)),

		field.Bytes("token_hash").
			Unique().
			MaxLen(32).
			SchemaType(map[string]string{
				"mysql": "binary(32)",
			}),

		field.String("ip").
			MaxLen(45).
			Optional(),

		field.Text("user_agent").
			MaxLen(1024).
			Optional(),

		field.Time("expires_at"),

		field.Time("revoked_at").
			Optional().
			Nillable(),
	}
}

func (Token) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
	}
}

func (Token) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id", "revoked_at"),
	}
}

// Edges of the Token.
func (Token) Edges() []ent.Edge {
	return nil
}
