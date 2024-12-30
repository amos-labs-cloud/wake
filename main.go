package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stianeikeland/go-rpio"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	pinNumber          int
	macAddress         string
	broadcastInterface string
	logLevel           string
	dryRun             bool
	logger             *slog.Logger
)

type Config struct {
	PinNumber          int    `mapstructure:"pin"`
	MacAddress         string `mapstructure:"mac_address"`
	BroadcastInterface string `mapstructure:"interface"`
	LogLevel           string `mapstructure:"log_level"`
}

func main() {
	viper.SetConfigName("wake")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/wake")

	viper.SetDefault("pin", 20)
	viper.SetDefault("mac_address", "11:22:33:44:55:66")
	viper.SetDefault("interface", "wlan0")
	viper.SetDefault("log.level", "info")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		os.Exit(1)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %s\n", err)
		os.Exit(1)
	}

	var rootCmd = &cobra.Command{
		Use:   "wake",
		Short: "a utility to monitor GPIO pin state and send a WOL packet",
		RunE:  run,
	}

	rootCmd.Flags().IntVar(&pinNumber, "pin", config.PinNumber, "GPIO pin number to monitor")
	rootCmd.Flags().StringVar(&macAddress, "mac-address", config.MacAddress, "MAC address to send WOL packet")
	rootCmd.Flags().StringVar(&broadcastInterface, "broadcast-interface", config.BroadcastInterface, "Broadcast interface name")
	rootCmd.Flags().StringVar(&logLevel, "log-level", config.LogLevel, "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run mode (do not send WOL packet)")

	rootCmd.MarkFlagsMutuallyExclusive("dry-run", "mac-address")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error executing root command:", err)
		os.Exit(1)
	}
}

func run(_ *cobra.Command, _ []string) error {
	level, err := parseLogLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", err)
	}
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	logger = logger.With(slog.Int("pin", pinNumber), slog.String("mac_address", macAddress), slog.String("interface", broadcastInterface))

	logger.Info("Starting wake utility")

	if err := rpio.Open(); err != nil {
		return fmt.Errorf("error opening GPIO: %s", err)
	}
	defer func() {
		err := rpio.Close()
		if err != nil {
			logger.Error("unable to close GPIO", slog.With("error", err))
		}
	}()

	pin := rpio.Pin(pinNumber)
	pin.Input()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	prevState := pin.Read()
	magicPacket, err := createMagicPacket(macAddress)
	if err != nil {
		return fmt.Errorf("error creating magic packet: %s", err)
	}

	iface, err := net.InterfaceByName(broadcastInterface)
	if err != nil {
		return fmt.Errorf("error getting interface: %s", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return fmt.Errorf("error getting interface addresses: %s", err)
	}

	var localAddr net.Addr
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			localAddr = addr
			break
		}
	}

	if localAddr == nil {
		return fmt.Errorf("no suitable local address found")
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 9,
	})
	if err != nil {
		return fmt.Errorf("error setting up UDP client: %s", err)
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			logger.Error("unable to close UDP connection", slog.String("error", err.Error()))
		}
	}(conn)

	for {
		select {
		case <-sigChan:
			logger.Info("Shutting down gracefully...")
			return nil
		default:
			currentState := pin.Read()
			if currentState != prevState {
				if currentState == rpio.Low {
					logger.Info("Signal received")
					if !dryRun {
						if _, err := conn.Write(magicPacket); err != nil {
							logger.Error("Error sending WOL packet", slog.With("error", err))
						}
					} else {
						logger.Info("Dry run mode: WOL packet not sent")
					}
				} else {
					logger.Info("Signal cleared")
				}
				prevState = currentState
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func createMagicPacket(mac string) ([]byte, error) {
	mac = strings.ReplaceAll(mac, ":", "")
	macBytes, err := hex.DecodeString(mac)
	if err != nil {
		return nil, fmt.Errorf("invalid MAC address: %s", err)
	}

	var packet []byte
	packet = append(packet, bytes.Repeat([]byte{0xFF}, 6)...)
	for i := 0; i < 16; i++ {
		packet = append(packet, macBytes...)
	}

	return packet, nil
}

func parseLogLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level: %s", level)
	}
}
