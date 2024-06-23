package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"hostManager/internal/transport/client"
)

const (
	DefaultServerAddr = "localhost:50051"
	DefaultTTL        = 10 * time.Second
)

var (
	serverAddr = DefaultServerAddr
	TTL        = DefaultTTL
)

var gRPCClient *client.GRPCClient

var rootCmd = &cobra.Command{
	Use: "host-manager",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		gRPCClient = client.NewGRPCClient(serverAddr)
	},
}

var setHostname = &cobra.Command{
	Use:   "set-hostname <hostname>",
	Short: "set hostname on machine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := gRPCClient.Close(); err != nil {
				log.Fatal().Msg("failed to close gRPC cli")
			}
		}()

		hostname := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), TTL)
		defer cancel()

		if err := gRPCClient.SetHostname(ctx, hostname); err != nil {
			log.Fatal().Err(err).Msg("failed to set hostname")
		}

		fmt.Printf("set hostname %s\n", hostname)
	},
}

var listDNSService = &cobra.Command{
	Use:   "list-dns-services",
	Short: "show list all DNS services",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := gRPCClient.Close(); err != nil {
				log.Fatal().Msg("failed to close gRPC cli")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), TTL)
		defer cancel()

		services, err := gRPCClient.ListDNSServers(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list DNS servers")
		}

		for _, service := range services {
			fmt.Println(service)
		}
	},
}

var addDNSServer = &cobra.Command{
	Use:   "add-dns-server <servername>",
	Short: "add a DNS server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := gRPCClient.Close(); err != nil {
				log.Fatal().Msg("failed to close gRPC cli")
			}
		}()
		servername := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), TTL)
		defer cancel()

		err := gRPCClient.AddDNSServer(ctx, servername)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add DNS server")
		}

		fmt.Printf("add server %s\n", servername)
	},
}

var removeDNSServer = &cobra.Command{
	Use:   "remove-dns-server <servername>",
	Short: "remove a DNS server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := gRPCClient.Close(); err != nil {
				log.Fatal().Msg("failed to close gRPC cli")
			}
		}()
		servername := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), TTL)
		defer cancel()

		err := gRPCClient.RemoveDNSServer(ctx, servername)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to remove DNS server")
		}

		fmt.Printf("remove server %s\n", servername)
	},
}

func main() {
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server-addr", DefaultServerAddr, "grpc addr")
	rootCmd.PersistentFlags().DurationVar(&TTL, "TTL", DefaultTTL, "request timeout")

	rootCmd.AddCommand(setHostname)
	rootCmd.AddCommand(listDNSService)
	rootCmd.AddCommand(addDNSServer)
	rootCmd.AddCommand(removeDNSServer)

	rootCmd.Execute()
}
