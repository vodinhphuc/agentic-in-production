// admin-password reads ADMIN_PASSWORD from stdin and writes its bcrypt hash to .env.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
)

func main() {
	fmt.Print("admin password (will not echo to .env literally; we hash it): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	pw := strings.TrimSpace(scanner.Text())
	if pw == "" {
		fmt.Fprintln(os.Stderr, "empty password")
		os.Exit(1)
	}
	hash, err := auth.HashPassword(pw)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	envPath := "../.env" // run from backend/
	updateEnv(envPath, "ADMIN_PASSWORD_HASH", hash)
	fmt.Println("OK — ADMIN_PASSWORD_HASH written to", envPath)
}

// updateEnv writes key='val' (single-quoted) so bcrypt's $-laden hashes survive
// `set -a; source .env`. docker-compose treats single quotes as literal value
// delimiters and strips them, so the postgres vars in docker-compose.yml are
// unaffected.
func updateEnv(path, key, val string) {
	b, _ := os.ReadFile(path)
	lines := strings.Split(string(b), "\n")
	quoted := key + "='" + val + "'"
	found := false
	for i, ln := range lines {
		if strings.HasPrefix(ln, key+"=") {
			lines[i] = quoted
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, quoted)
	}
	_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
