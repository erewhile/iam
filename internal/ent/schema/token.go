package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
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
			Unique(),

		field.UUID("session_id", uuid.UUID{}),

		field.String("token_hash").
			Unique().
			MaxLen(64).
			SchemaType(map[string]string{
				"mysql": "char(64)",
			}),

		field.String("ip").
			MaxLen(45).
			Optional(),

		field.Text("user_agent").
			Optional(),

		field.Time("expires_at"),

		field.Time("revoked_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Token.
func (Token) Edges() []ent.Edge {
	return nil
}
