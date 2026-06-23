package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/erewhile/iam/pkg/totp"
)

type Account struct {
	Label     string `json:"label"`
	Secret    string `json:"secret"`
	Algorithm string `json:"algorithm"`
	Period    int64  `json:"period"`
	Digits    int    `json:"digits"`
	Issuer    string `json:"issuer"`
}

func main() {
	secretFlag := flag.String("s", "", "directly provide secret")
	configPath := flag.String("c", "~/.totp.json", "config file path")
	flag.Parse()

	var targetSecret string
	var conf totp.Config
	var label string

	if *secretFlag != "" {
		targetSecret = *secretFlag
		conf = totp.Config{Algorithm: totp.SHA1, Digits: totp.Digits6, Timestep: totp.Timestep30s}
		label = "manual"
	} else {
		path, err := expandPath(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config path error: %v\n", err)
			os.Exit(1)
		}

		warnIfWorldReadable(path)

		accounts, err := loadConfig(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config error: %v\n", err)
			os.Exit(1)
		}

		target := flag.Arg(0)
		if target == "" {
			printAccountList(accounts)
			return
		}

		acc, found := findAccount(accounts, target)
		if !found {
			fmt.Fprintf(os.Stderr, "label %q not found.\n", target)
			os.Exit(1)
		}
		if strings.TrimSpace(acc.Secret) == "" {
			fmt.Fprintf(os.Stderr, "account %q has no secret configured.\n", acc.Label)
			os.Exit(1)
		}

		targetSecret = acc.Secret
		conf = totp.Config{
			Algorithm:  totp.Algorithm(acc.Algorithm),
			Digits:     acc.Digits,
			Timestep:   acc.Period,
			MinKeySize: 10,
		}
		label = displayLabel(acc)
	}

	if _, err := totp.New(conf, nil); err != nil {
		fmt.Fprintf(os.Stderr, "invalid TOTP config: %v\n", err)
		os.Exit(1)
	}

	conf = conf.WithDefaults()

	fmt.Printf("account: %s\n", label)
	fmt.Println("press Ctrl+C to stop.")
	fmt.Println("--------------------------")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println()
		os.Exit(0)
	}()

	for {
		now := time.Now()
		code, err := totp.Generate(targetSecret, now, conf)
		if err != nil {
			fmt.Printf("\nerror: %v\n", err)
			os.Exit(1)
		}
		remaining := conf.Timestep - (now.Unix() % conf.Timestep)
		fmt.Printf("\rcode: \033[1;32m%s\033[0m  Expires in: %2ds ", code, remaining)
		time.Sleep(time.Second)
	}
}

func displayLabel(acc Account) string {
	if acc.Issuer != "" {
		return fmt.Sprintf("%s (%s)", acc.Label, acc.Issuer)
	}
	return acc.Label
}

func findAccount(accounts []Account, target string) (Account, bool) {
	for _, acc := range accounts {
		if strings.EqualFold(acc.Label, target) {
			return acc, true
		}
	}
	return Account{}, false
}

func printAccountList(accounts []Account) {
	fmt.Println("usage: totp <label>")
	if len(accounts) == 0 {
		fmt.Println("no accounts configured.")
		return
	}
	fmt.Println("available labels:")
	for _, acc := range accounts {
		fmt.Printf(" - %s\n", displayLabel(acc))
	}
}

func loadConfig(path string) ([]Account, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var accs []Account
	if err := json.Unmarshal(data, &accs); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return accs, nil
}

func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("could not resolve home directory: %w", err)
	}
	return filepath.Join(usr.HomeDir, strings.TrimPrefix(path, "~")), nil
}

func warnIfWorldReadable(path string) {
	if runtime.GOOS == "windows" {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.Mode().Perm()&0o077 != 0 {
		fmt.Fprintf(os.Stderr,
			"warning: %s is readable by other users (mode %04o). It contains plaintext TOTP secrets.\nConsider: chmod 600 %s\n",
			path, info.Mode().Perm(), path)
	}
}
