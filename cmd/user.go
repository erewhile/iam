package cmd

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/database"
	"github.com/erewhile/iam/internal/database/data"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/pkg/password"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "iam user",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("user called")
	},
}

var (
	addUserEmail    string
	addUserUsername string
)

var userAddCmd = &cobra.Command{
	Use:          "add",
	Short:        "create a new user with a randomly generated password",
	SilenceUsage: true,
	RunE:         runUserAdd,
}

func runUserAdd(cmd *cobra.Command, args []string) error {
	setup()
	defer release()

	if err := database.Init(config.Get().Database); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer database.Close()

	client := database.GetDB()
	userRepo := repository.NewUserRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := userRepo.GetByID(ctx, model.UserSystemID)
	if err != nil {
		if !db.IsNotFound(err) {
			return fmt.Errorf("failed to get user info: %w", err)
		}
	} else {
		return errors.New("admin already exists")
	}

	passwordStr, err := randomPassword(8)
	if err != nil {
		return fmt.Errorf("failed to generate random password: %w", err)
	}

	hashed, err := password.Hash(passwordStr)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u, err := userRepo.Create(ctx, req.UserCreate{
		Email:    addUserEmail,
		Username: addUserUsername,
		Status:   model.UserStatusActive,
	}, hashed, model.UserSystem)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err := data.InitData(client); err != nil {
		return fmt.Errorf("failed to init system role data: %w", err)
	}

	fmt.Println("user created successfully, save the password now — it will not be shown again:")
	fmt.Printf("  id:       %v\n", u.ID)
	fmt.Printf("  email:    %s\n", u.Email)
	fmt.Printf("  username: %s\n", u.Username)
	fmt.Printf("  password: %s\n", passwordStr)

	return nil
}

func randomPassword(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("password length must be positive, got %d", n)
	}

	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}
func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userAddCmd)

	userAddCmd.Flags().StringVar(&addUserEmail, "email", "", "email of the new user (required)")
	userAddCmd.Flags().StringVar(&addUserUsername, "username", "", "username of the new user (required)")
	_ = userAddCmd.MarkFlagRequired("email")
	_ = userAddCmd.MarkFlagRequired("username")
}
