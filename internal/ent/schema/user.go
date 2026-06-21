package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "user",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("user"),
		entsql.WithComments(true),
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			MaxLen(26).
			NotEmpty().
			Unique(),

		field.Bytes("password_hash").
			NotEmpty().
			Sensitive().
			MaxLen(16).
			SchemaType(map[string]string{
				"mysql": "binary(16)",
			}),

		field.String("email").
			MaxLen(128).
			NotEmpty().
			Unique(),

		field.Uint8("status").
			Default(0),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
