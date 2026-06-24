package cmd

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/database"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/pkg/password"
	"github.com/spf13/cobra"
)

// userCmd represents the user command
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
	Use:   "add",
	Short: "create a new user with a randomly generated password",
	RunE:  runUserAdd,
}

func runUserAdd(cmd *cobra.Command, args []string) error {
	setup()
	defer release()

	if err := database.Init(config.Get().Database); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer database.Close()

	passwordStr, err := randomPassword(16)
	if err != nil {
		return fmt.Errorf("failed to generate random password: %w", err)
	}

	hashed, err := password.Hash(passwordStr)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userRepo := repository.NewUserRepository(database.GetDB())
	u, err := userRepo.Create(ctx, req.UserCreate{
		Email:    addUserEmail,
		Username: addUserUsername,
		Status:   model.UserStatusActive,
	}, hashed, model.UserSystem)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
