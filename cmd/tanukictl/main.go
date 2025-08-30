package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"

	"github.com/lmittmann/tint"

	"github.com/fiwippi/tanuki/v2"
)

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.DateTime,
		}),
	))
}

func main() {
	defaultConfig := tanuki.DefaultServerConfig()
	host := flag.String("host", defaultConfig.Host, "Host address of tanuki")
	port := flag.String("port", strconv.Itoa(int(defaultConfig.RpcPort)), "Port tanuki's RPC handler is listening on")
	printVersion := flag.Bool("version", false, "Output version information and exit")
	flag.Usage = flagUsage
	flag.Parse()

	if err := run(*host, *port, *printVersion); err != nil {
		slog.Error("Failed to run tanukictl", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(host, port string, printVersion bool) error {
	if printVersion {
		fmt.Printf("tanukictl %s\n", tanuki.Version)
		return nil
	}

	socketAddr := net.JoinHostPort(host, port)
	conn, err := net.Dial("tcp", socketAddr)
	if err != nil {
		return fmt.Errorf("dial tcp: %w", err)
	}
	rpc := rpc.NewClient(conn)

	switch flag.Arg(0) {
	case "scan":
		return scanLibrary(rpc)
	case "dump":
		return dumpStore(rpc)
	case "user":
		return modifyUser(rpc)
	default:
		slog.Error("Invalid command", slog.String("command", flag.Arg(0)))
		flagUsage()
		return nil
	}
}

// Commands

func flagUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: tanukictl [options] [command] args\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Commands:\n")
	fmt.Fprintf(out, "  scan                                  Scan the library\n")
	fmt.Fprintf(out, "  dump                                  Dump the store's state\n")
	fmt.Fprintf(out, "  user add <name>                       Add a new user with the password provided via stdin\n")
	fmt.Fprintf(out, "  user delete <name>                    Delete an existing user\n")
	fmt.Fprintf(out, "  user edit name <old-name> <new-name>  Change a user's name\n")
	fmt.Fprintf(out, "  user edit pass <name>                 Change a user's password provided via stdin\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "  $ tanukictl -port 5000 scan\n")
	fmt.Fprintf(out, "  $ tanukictl -port 5000 dump\n")
	fmt.Fprintf(out, "    // Scan the library, then dump the store's contents\n")
	fmt.Fprintf(out, "    // We connect to a tanuki instance listening on a\n")
	fmt.Fprintf(out, "    // standard host but a non-standard port (5000)\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "  $ tanukictl user edit name old-name new-name\n")
	fmt.Fprintf(out, "  $ echo \"new-password\" | tanukictl user edit pass new-name\n")
	fmt.Fprintf(out, "    // Edit a user's name, then their password\n")
}

func scanLibrary(api *rpc.Client) error {
	start := time.Now()
	if err := api.Call("Server.Scan", struct{}{}, &struct{}{}); err != nil {
		return fmt.Errorf("scan library: %w", err)
	}
	fmt.Printf("Scan complete in %s\n", time.Since(start).Round(time.Millisecond))
	return nil
}

func dumpStore(api *rpc.Client) error {
	output := new(string)
	if err := api.Call("Server.Dump", struct{}{}, output); err != nil {
		return fmt.Errorf("dump store: %w", err)
	}
	fmt.Print(*output)
	return nil
}

func modifyUser(api *rpc.Client) error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("stat os.Stdin: %w", err)
	}
	canReadFromStdin := (stat.Mode() & os.ModeCharDevice) == 0

	var input []byte
	if canReadFromStdin {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read os.Stdin: %w", err)
		}
		input = bytes.TrimRight(input, "\n")
	}

	switch flag.Arg(1) {
	case "add":
		u := tanuki.User{
			Name: flag.Arg(2),
			Pass: string(input),
		}
		if err := api.Call("Server.AddUser", u, &struct{}{}); err != nil {
			return fmt.Errorf("add user: %w", err)
		}
		fmt.Println("Added user")
	case "delete":
		if err := api.Call("Server.DeleteUser", flag.Arg(2), &struct{}{}); err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		fmt.Println("Deleted user")
	case "edit":
		switch flag.Arg(2) {
		case "name":
			req := tanuki.ChangeUsernameRequest{
				OldName: flag.Arg(3),
				NewName: flag.Arg(4),
			}
			if err := api.Call("Server.ChangeUsername", req, &struct{}{}); err != nil {
				return fmt.Errorf("change username: %w", err)
			}
			fmt.Println("Changed username")
		case "pass":
			req := tanuki.ChangePasswordRequest{
				Name:     flag.Arg(3),
				Password: string(input),
			}
			if err := api.Call("Server.ChangePassword", req, &struct{}{}); err != nil {
				return fmt.Errorf("change password: %w", err)
			}
			fmt.Println("Changed password")
		default:
			slog.Error("Invalid command", slog.String("command", flag.Arg(2)))
			flagUsage()
		}
	default:
		slog.Error("Invalid command", slog.String("command", flag.Arg(1)))
		flagUsage()
	}

	return nil
}
