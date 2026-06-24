package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/erewhile/iam/internal/ent/mixin"
	"github.com/erewhile/iam/internal/model"
	"github.com/google/uuid"
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

		field.UUID("uuid", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Unique(),

		field.Bytes("password_hash").
			NotEmpty().
			Sensitive().
			MaxLen(255).
			SchemaType(map[string]string{
				"mysql": "varbinary(255)",
			}),

		field.String("email").
			MinLen(6).
			MaxLen(128).
			NotEmpty().
			Unique(),

		field.Uint8("status").
			GoType(model.UserStatus(0)).
			Default(uint8(model.UserStatusPending)),

		field.Bool("is_system").
			Immutable().
			Default(false),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roles", Role.Type).
			Through("user_roles", UserRole.Type),
	}
}
