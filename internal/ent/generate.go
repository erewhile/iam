package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature intercept,schema/snapshot,sql/exec,sql/lock,sql/modifier,feature/comment,sql/upsert ./schema --target ./db
