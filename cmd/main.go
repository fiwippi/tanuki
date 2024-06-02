package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
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
	configPath := flag.String("config", "", "Path to config.json file. Leave blank to use the default config")
	flag.Usage = flagUsage
	flag.Parse()

	config := tanuki.DefaultServerConfig()
	if *configPath != "" {
		f, err := os.Open(*configPath)
		if err != nil {
			exit("Could not open config file", err)
		}
		if err := json.NewDecoder(f).Decode(&config); err != nil {
			exit("Could not decode config file", err)
		}
	}

	switch flag.Arg(0) {
	case "run":
		runServer(config)
	case "scan":
		scanLibrary(dialRPC(config))
	case "dump":
		dumpStore(dialRPC(config))
	case "user":
		modifyUser(dialRPC(config))
	default:
		panic("Invalid command")
	}
}

// Commands

func flagUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: tanuki [options] [command] args\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Commands:\n")
	fmt.Fprintf(out, "  run                                   Run the server\n")
	fmt.Fprintf(out, "  scan                                  Scan the library\n")
	fmt.Fprintf(out, "  dump                                  Dump the store's state\n")
	fmt.Fprintf(out, "  user add <name>                       Add a new user with the password provided via stdin\n")
	fmt.Fprintf(out, "  user delete <name>                    Delete an existing user\n")
	fmt.Fprintf(out, "  user edit name <old-name> <new-name>  Change a user's name\n")
	fmt.Fprintf(out, "  user edit pass <name>                 Change a user's password provided via stdin\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Examples:\n")
	fmt.Fprintf(out, "  $ tanuki -config /path/to/config.json run\n")
	fmt.Fprintf(out, "    // Run the server using a specific config\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "  $ tanuki scan\n")
	fmt.Fprintf(out, "  $ tanuki dump\n")
	fmt.Fprintf(out, "    // Scan the library, then dump the store's contents\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "  $ tanuki user edit name old-name new-name\n")
	fmt.Fprintf(out, "  $ echo \"new-password\" | tanuki user edit pass new-name\n")
	fmt.Fprintf(out, "    // Edit a user's name, then its password. Since the server we are \n")
	fmt.Fprintf(out, "    // connecting to exposes a custom RPC port, we also supply the\n")
	fmt.Fprintf(out, "    // config to the CLI, (which details the value of the port)\n")
}

func runServer(config tanuki.ServerConfig) {
	s, err := tanuki.NewServer(config)
	if err != nil {
		exit("Failed to create server", err)
	}

	if err := s.Start(); err != nil {
		exit("Failed to start server", err)
	}
	<-done()
	s.Stop()
}

func dialRPC(config tanuki.ServerConfig) *rpc.Client {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.RpcPort))
	if err != nil {
		exit("Could not dial the RPC port", err)
	}
	return rpc.NewClient(conn)
}

func scanLibrary(api *rpc.Client) {
	start := time.Now()
	if err := api.Call("Server.Scan", struct{}{}, &struct{}{}); err != nil {
		exit("Failed to scan library", err)
	}
	fmt.Printf("Scan complete in %s\n", time.Since(start).Round(time.Millisecond))
}

func dumpStore(api *rpc.Client) {
	output := new(string)
	if err := api.Call("Server.Dump", struct{}{}, output); err != nil {
		exit("Failed to dump store", err)
	}
	fmt.Print(*output)
}

func modifyUser(api *rpc.Client) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		exit("Failed to state os.Stdin", err)
	}
	canReadFromStdin := (stat.Mode() & os.ModeCharDevice) == 0

	var input []byte
	if canReadFromStdin {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			exit("Failed to read os.Stdin", err)
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
			exit("Failed to add user", err)
		}
		fmt.Println("Added user")
	case "delete":
		if err := api.Call("Server.DeleteUser", flag.Arg(2), &struct{}{}); err != nil {
			exit("Failed to delete user", err)
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
				exit("Failed to change username", err)
			}
			fmt.Println("Changed username")
		case "pass":
			req := tanuki.ChangePasswordRequest{
				Name:     flag.Arg(3),
				Password: string(input),
			}
			if err := api.Call("Server.ChangePassword", req, &struct{}{}); err != nil {
				exit("Failed to change password", err)
			}
			fmt.Println("Changed password")
		default:
			panic("Invalid command")
		}
	default:
		panic("Invalid command")
	}
}

// Helpers

func done() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return c
}

func exit(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
	os.Exit(1)
}
