package data

import "github.com/erewhile/iam/internal/ent/db"

func InitData(client *db.Client) error {
	if err := initRole(client); err != nil {
		return err
	}
	if err := initUserRole(client); err != nil {
		return err
	}
	return nil
}
