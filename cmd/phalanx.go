package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	homedir "github.com/mitchellh/go-homedir"
	phalanxcluster "github.com/mosuka/phalanx/cluster"
	"github.com/mosuka/phalanx/gateway"
	"github.com/mosuka/phalanx/index"
	"github.com/mosuka/phalanx/logging"
	phalanxmetastore "github.com/mosuka/phalanx/metastore"
	"github.com/mosuka/phalanx/server"
	"github.com/mosuka/phalanx/util"
	"github.com/mosuka/phalanx/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Default gRPC address
const defaultEnvFile string = ".env"
const defaultConfigFile string = ""

const defaultHost string = "0.0.0.0"
const defaultBindPort int = 3000
const defaultGrpcPort int = 5000
const defaultHttpPort int = 8000

const defaultIndexMetastoreUri string = "file:///var/lib/phalanx/metastore"

const defaultCertificateFile string = ""
const defaultKeyFile string = ""
const defaultCommonName string = ""

const defaultLogLevel string = "INFO"
const defaultLogFile string = ""
const defaultLogMaxSize int = 500
const defaultLogMaxBackups int = 5
const defaultLogMaxAge int = 30
const defaultLogCompress bool = false

var (
	envFile    string
	configFile string

	host          string
	bindPort      int
	grpcPort      int
	httpPort      int
	seedAddresses []string
	roles         []string

	indexMetastoreUri string

	certificateFile string
	keyFile         string
	commonName      string

	corsAllowedMethods []string
	corsAllowedOrigins []string
	corsAllowedHeaders []string

	logLevel      string
	logFile       string
	logMaxSize    int
	logMaxBackups int
	logMaxAge     int
	logCompress   bool

	phalanxCmd = &cobra.Command{
		Use:   "phalanx",
		Short: "Phalanx",
		Long:  "Phalanx server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check version flag.
			version_flag, err := cmd.Flags().GetBool("version")
			if err != nil {
				return err
			}
			if version_flag {
				fmt.Printf("Phalanx version: %s\n", version.Version)
				return nil
			}

			host = viper.GetString("host")
			bindPort = viper.GetInt("bind_port")
			grpcPort = viper.GetInt("grpc_port")
			httpPort = viper.GetInt("http_port")
			roles = viper.GetStringSlice("roles")

			seedAddresses = viper.GetStringSlice("seed_addresses")

			indexMetastoreUri = viper.GetString("index_metastore_uri")

			certificateFile = viper.GetString("certificate_file")
			keyFile = viper.GetString("key_file")
			commonName = viper.GetString("common_name")

			corsAllowedMethods = viper.GetStringSlice("cors_allowed_methods")
			corsAllowedOrigins = viper.GetStringSlice("cors_allowed_origins")
			corsAllowedHeaders = viper.GetStringSlice("cors_allowed_headers")

			logLevel = viper.GetString("log_level")
			logFile = viper.GetString("log_file")
			logMaxSize = viper.GetInt("log_max_size")
			logMaxBackups = viper.GetInt("log_max_backups")
			logMaxAge = viper.GetInt("log_max_age")
			logCompress = viper.GetBool("log_compress")

			logger := logging.NewLogger(
				logLevel,
				logFile,
				logMaxSize,
				logMaxBackups,
				logMaxAge,
				logCompress,
			)

			host, err = util.ResolveHost(host)
			if err != nil {
				return err
			}

			nodeRoles := make([]phalanxcluster.NodeRole, 0)
			for _, role := range roles {
				nodeRoles = append(nodeRoles, phalanxcluster.NodeRole(phalanxcluster.NodeRole_value[role]))
			}
			nodeMetadata := phalanxcluster.NodeMetadata{
				GrpcPort: grpcPort,
				HttpPort: httpPort,
				Roles:    nodeRoles,
			}

			isSeedNode := len(seedAddresses) == 0

			// Create cluster
			cluster, err := phalanxcluster.NewCluster(host, bindPort, nodeMetadata, isSeedNode, logger)
			if err != nil {
				logger.Error("Failed to create node", zap.Error(err), zap.String("host", host), zap.Int("bind_port", bindPort))
				return err
			}

			// Create index metastore
			metastore, err := phalanxmetastore.NewMetastore(indexMetastoreUri, logger)
			if err != nil {
				logger.Error("failed to create metastore", zap.Error(err), zap.Any("uri", indexMetastoreUri))
				return err
			}
			if err := metastore.Start(); err != nil {
				logger.Error("failed to start metastore", zap.Error(err))
				return err
			}

			// Create index manager
			indexManager, err := index.NewManager(cluster, metastore, certificateFile, commonName, logger)
			if err != nil {
				logger.Error("failed to create index manager", zap.Error(err))
				return err
			}
			if err := indexManager.Start(); err != nil {
				logger.Error("failed to start index manager", zap.Error(err))
				return err
			}

			// Create indexService
			indexService, err := server.NewIndexService(indexManager, certificateFile, commonName, logger)
			if err != nil {
				logger.Error("failed to create index service", zap.Error(err))
				return err
			}

			// Create indexServer
			grpcAddress := fmt.Sprintf("%s:%d", host, grpcPort)
			indexServer, err := server.NewIndexServer(grpcAddress, certificateFile, keyFile, commonName, indexService, logger)
			if err != nil {
				logger.Error("failed to create index server", zap.Error(err))
				return err
			}

			httpAddress := fmt.Sprintf("%s:%d", host, httpPort)
			indexxGateway, err := gateway.NewIndexGatewayWithTLS(httpAddress, grpcAddress, certificateFile, keyFile, commonName, corsAllowedMethods, corsAllowedOrigins, corsAllowedHeaders, logger)
			if err != nil {
				logger.Error("failed to create index gateway", zap.Error(err))
				return err
			}

			// Start node
			if err := cluster.Start(); err != nil {
				return err
			}

			// Join cluster.
			if !isSeedNode {
				resolvedSeedAddresses := make([]string, 0)
				for _, seedAddress := range seedAddresses {
					seedHost, seedPort, err := net.SplitHostPort(seedAddress)
					if err != nil {
						logger.Error("failed to split seed address", zap.Error(err), zap.String("seed_address", seedAddress))
						return err
					}
					resolvedSeedHost, err := util.ResolveHost(seedHost)
					if err != nil {
						logger.Error("failed to resolve seed host", zap.Error(err), zap.String("seed_host", seedHost))
						return err
					}
					resolvedSeedAddresses = append(resolvedSeedAddresses, fmt.Sprintf("%s:%s", resolvedSeedHost, seedPort))
				}
				_, err := cluster.Join(resolvedSeedAddresses)
				if err != nil {
					logger.Error("failed to join to the cluster", zap.Error(err), zap.Any("seed_addresses", seedAddresses))
					return err
				}
			}

			// Start server
			if err := indexServer.Start(); err != nil {
				return err
			}

			// Start gateway
			if err := indexxGateway.Start(); err != nil {
				return err
			}

			// Make signal channel.
			quitCh := make(chan os.Signal, 1)
			signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)

			// Wait for receiving signal.
			<-quitCh

			// Leave the cluster.
			if err := cluster.Leave(10 * time.Second); err != nil {
				logger.Error("failed to leave cluster", zap.Error(err))
			}

			// Stop node
			if err := cluster.Stop(); err != nil {
				logger.Error("failed to stop node", zap.Error(err))
			}

			// Stop index manager.
			if err := indexManager.Stop(); err != nil {
				logger.Error("failed to stop index manager", zap.Error(err))
			}

			// Stop index metastore.
			if err := metastore.Stop(); err != nil {
				logger.Error("failed to stop index metastore", zap.Error(err))
			}

			// Stop server.
			err = indexxGateway.Stop()
			if err != nil {
				logger.Error("failed to stop gatway", zap.Error(err))
			}

			// Stop server.
			err = indexServer.Stop()
			if err != nil {
				logger.Error("failed to stop server", zap.Error(err))
			}

			return nil
		},
	}
)

func Execute() error {
	if err := phalanxCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}

func init() {
	cobra.OnInitialize(func() {
		// loads values from .env into the system
		godotenv.Load(envFile)

		if configFile != "" {
			// Set the path to the configuration file.
			viper.SetConfigFile(configFile)
		} else {
			// If the path to the configuration file is ommitted,
			// phalanx.yaml will be searched for in the /etc directory,
			// then the home directory, and will be set if found.
			home, err := homedir.Dir()
			if err != nil {
				// Failed to get home directory.
				// Exit the program.
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			viper.AddConfigPath("/etc")
			viper.AddConfigPath(home)
			viper.AddConfigPath("./")
			viper.SetConfigName("phalanx")
		}

		// Setup environment variables
		viper.SetEnvPrefix("PHALANX")
		viper.AutomaticEnv()

		// Read configuration file.
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				// The configuration file does not found in search path.
				// Skip reading the configuration file.
			default:
				// Failed to read the configuration file.
				// Exit the program.
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	})

	phalanxCmd.Flags().StringVar(&envFile, "env-file", defaultEnvFile, "path to the environment variables file. if omitted, .env in the current directory will be set.")
	phalanxCmd.Flags().StringVar(&configFile, "config-file", defaultConfigFile, "path to the configuration file. if omitted, phalanx.yaml will be searched for in the /etc directory, then the home directory, and will be set if found.")

	phalanxCmd.Flags().StringVar(&host, "host", defaultHost, "host address")
	phalanxCmd.Flags().IntVar(&bindPort, "bind-port", defaultBindPort, "Bind port")
	phalanxCmd.Flags().IntVar(&grpcPort, "grpc-port", defaultGrpcPort, "gRPC port")
	phalanxCmd.Flags().IntVar(&httpPort, "http-port", defaultHttpPort, "HTTP port")
	phalanxCmd.Flags().StringSliceVar(&seedAddresses, "seed-addresses", []string{}, "seed address (e.g. 192.168.1.10:2000, 192.168.1.11)")
	phalanxCmd.Flags().StringSliceVar(&roles, "roles", []string{string(phalanxcluster.NodeRole_name[1]), string(phalanxcluster.NodeRole_name[2])}, "node roles (ex: indexer,searcher)")

	phalanxCmd.Flags().StringVar(&indexMetastoreUri, "index-metastore-uri", defaultIndexMetastoreUri, "index metastore URI.")

	phalanxCmd.Flags().StringVar(&certificateFile, "certificate-file", defaultCertificateFile, "path to the client server TLS certificate file")
	phalanxCmd.Flags().StringVar(&keyFile, "key-file", defaultKeyFile, "path to the client server TLS key file")
	phalanxCmd.Flags().StringVar(&commonName, "common-name", defaultCommonName, "certificate common name")

	phalanxCmd.Flags().StringSliceVar(&corsAllowedMethods, "cors-allowed-methods", []string{}, "CORS allowed methods (e.g. GET,PUT,DELETE,POST)")
	phalanxCmd.Flags().StringSliceVar(&corsAllowedOrigins, "cors-allowed-origins", []string{}, "CORS allowed origins (e.g. http://localhost:8080,http://localhost:80)")
	phalanxCmd.Flags().StringSliceVar(&corsAllowedHeaders, "cors-allowed-headers", []string{}, "CORS allowed headers (e.g. content-type,x-some-key)")

	phalanxCmd.Flags().StringVar(&logLevel, "log-level", defaultLogLevel, "log level")
	phalanxCmd.Flags().StringVar(&logFile, "log-file", defaultLogFile, "log file")
	phalanxCmd.Flags().IntVar(&logMaxSize, "log-max-size", defaultLogMaxSize, "max size of a log file in megabytes")
	phalanxCmd.Flags().IntVar(&logMaxBackups, "log-max-backups", defaultLogMaxBackups, "max backup count of log files")
	phalanxCmd.Flags().IntVar(&logMaxAge, "log-max-age", defaultLogMaxAge, "max age of a log file in days")
	phalanxCmd.Flags().BoolVar(&logCompress, "log-compress", defaultLogCompress, "compress a log file")

	phalanxCmd.Flags().BoolP("version", "v", false, "show version")

	phalanxCmd.Flags().SortFlags = false

	_ = viper.BindPFlag("host", phalanxCmd.Flags().Lookup("host"))
	_ = viper.BindPFlag("bind_port", phalanxCmd.Flags().Lookup("bind-port"))
	_ = viper.BindPFlag("grpc_port", phalanxCmd.Flags().Lookup("grpc-port"))
	_ = viper.BindPFlag("http_port", phalanxCmd.Flags().Lookup("http-port"))
	_ = viper.BindPFlag("seed_addresses", phalanxCmd.Flags().Lookup("seed-addresses"))
	_ = viper.BindPFlag("roles", phalanxCmd.Flags().Lookup("roles"))

	_ = viper.BindPFlag("index_metastore_uri", phalanxCmd.Flags().Lookup("index-metastore-uri"))

	_ = viper.BindPFlag("certificate_file", phalanxCmd.Flags().Lookup("certificate-file"))
	_ = viper.BindPFlag("key_file", phalanxCmd.Flags().Lookup("key-file"))
	_ = viper.BindPFlag("common_name", phalanxCmd.Flags().Lookup("common-name"))

	_ = viper.BindPFlag("cors_allowed_methods", phalanxCmd.Flags().Lookup("cors-allowed-methods"))
	_ = viper.BindPFlag("cors_allowed_origins", phalanxCmd.Flags().Lookup("cors-allowed-origins"))
	_ = viper.BindPFlag("cors_allowed_headers", phalanxCmd.Flags().Lookup("cors-allowed-headers"))

	_ = viper.BindPFlag("log_level", phalanxCmd.Flags().Lookup("log-level"))
	_ = viper.BindPFlag("log_max_size", phalanxCmd.Flags().Lookup("log-max-size"))
	_ = viper.BindPFlag("log_max_backups", phalanxCmd.Flags().Lookup("log-max-backups"))
	_ = viper.BindPFlag("log_max_age", phalanxCmd.Flags().Lookup("log-max-age"))
	_ = viper.BindPFlag("log_compress", phalanxCmd.Flags().Lookup("log-compress"))
}
