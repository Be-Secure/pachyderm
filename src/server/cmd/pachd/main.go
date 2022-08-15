package main

import (
	"context"
	gotls "crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"runtime/pprof"
	"syscall"

	adminclient "github.com/pachyderm/pachyderm/v2/src/admin"
	authclient "github.com/pachyderm/pachyderm/v2/src/auth"
	debugclient "github.com/pachyderm/pachyderm/v2/src/debug"
	eprsclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	identityclient "github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	logutil "github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	authmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	errorsmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/errors"
	loggingmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging"
	version_middleware "github.com/pachyderm/pachyderm/v2/src/internal/middleware/version"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/profileutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tls"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	licenseclient "github.com/pachyderm/pachyderm/v2/src/license"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	proxyclient "github.com/pachyderm/pachyderm/v2/src/proxy"
	adminserver "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/pfshttp"
	proxyserver "github.com/pachyderm/pachyderm/v2/src/server/proxy/server"
	"google.golang.org/grpc/health"

	identity_server "github.com/pachyderm/pachyderm/v2/src/server/identity/server"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/s3"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	txnserver "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	transactionclient "github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	_ "github.com/pachyderm/pachyderm/v2/src/internal/task/taskprotos"
)

var mode string
var readiness bool

type bootstrapper interface {
	EnvBootstrap(context.Context) error
}

func init() {
	flag.StringVar(&mode, "mode", "full", "Pachd currently supports four modes: full, enterprise, sidecar and paused. Full includes everything you need in a full pachd node. Enterprise runs the Enterprise Server. Sidecar runs only PFS, the Auth service, and a stripped-down version of PPS.  Paused runs all APIs other than PFS and PPS; it is intended to enable taking database backups.")
	flag.BoolVar(&readiness, "readiness", false, "Run readiness check.")
	flag.Parse()
}

func main() {
	log.SetFormatter(logutil.FormatterFunc(logutil.JSONPretty))
	// set GOMAXPROCS to the container limit & log outcome to stdout
	maxprocs.Set(maxprocs.Logger(log.Printf)) //nolint:errcheck
	log.Infof("version info: %v", version.Version)
	ctx := context.Background()

	switch {
	case readiness:
		cmdutil.Main(ctx, doReadinessCheck, &serviceenv.GlobalConfiguration{})
	case mode == "full", mode == "", mode == "$(MODE)":
		// Because of the way Kubernetes environment substitution works,
		// a reference to an unset variable is not replaced with the
		// empty string, but instead the reference is passed unchanged;
		// because of this, '$(MODE)' should be recognized as an unset —
		// i.e., default — mode.
		cmdutil.Main(ctx, doFullMode, &serviceenv.PachdFullConfiguration{})
	case mode == "enterprise":
		cmdutil.Main(ctx, doEnterpriseMode, &serviceenv.EnterpriseServerConfiguration{})
	case mode == "sidecar":
		cmdutil.Main(ctx, doSidecarMode, &serviceenv.PachdFullConfiguration{})
	case mode == "paused":
		cmdutil.Main(ctx, doPausedMode, &serviceenv.PachdFullConfiguration{})
	default:
		fmt.Printf("unrecognized mode: %s\n", mode)
	}
}

func newInternalServer(ctx context.Context, authInterceptor *authmw.Interceptor, loggingInterceptor *loggingmw.LoggingInterceptor) (*grpcutil.Server, error) {
	return grpcutil.NewServer(
		ctx,
		false,
		grpc.ChainUnaryInterceptor(
			errorsmw.UnaryServerInterceptor,
			tracing.UnaryServerInterceptor(),
			authInterceptor.InterceptUnary,
			loggingInterceptor.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			errorsmw.StreamServerInterceptor,
			tracing.StreamServerInterceptor(),
			authInterceptor.InterceptStream,
			loggingInterceptor.StreamServerInterceptor,
		),
	)
}

func newExternalServer(ctx context.Context, authInterceptor *authmw.Interceptor, loggingInterceptor *loggingmw.LoggingInterceptor) (*grpcutil.Server, error) {
	return grpcutil.NewServer(
		ctx,
		true,
		// Add an UnknownServiceHandler to catch the case where the user has a client with the wrong major version.
		// Weirdly, GRPC seems to run the interceptor stack before the UnknownServiceHandler, so this is never called
		// (because the version_middleware interceptor throws an error, or the auth interceptor does).
		grpc.UnknownServiceHandler(func(srv interface{}, stream grpc.ServerStream) error {
			method, _ := grpc.MethodFromServerStream(stream)
			return errors.Errorf("unknown service %v", method)
		}),
		grpc.ChainUnaryInterceptor(
			errorsmw.UnaryServerInterceptor,
			version_middleware.UnaryServerInterceptor,
			tracing.UnaryServerInterceptor(),
			authInterceptor.InterceptUnary,
			loggingInterceptor.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			errorsmw.StreamServerInterceptor,
			version_middleware.StreamServerInterceptor,
			tracing.StreamServerInterceptor(),
			authInterceptor.InterceptStream,
			loggingInterceptor.StreamServerInterceptor,
		),
	)
}

func setup(config interface{}, service string) (env serviceenv.ServiceEnv, err error) {
	switch logLevel := os.Getenv("LOG_LEVEL"); logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info", "":
		log.SetLevel(log.InfoLevel)
	default:
		log.Errorf("Unrecognized log level %s, falling back to default of \"info\"", logLevel)
		log.SetLevel(log.InfoLevel)
	}
	// must run InstallJaegerTracer before InitWithKube (otherwise InitWithKube
	// may create a pach client before tracing is active, not install the Jaeger
	// gRPC interceptor in the client, and not propagate traces)
	if endpoint := tracing.InstallJaegerTracerFromEnv(); endpoint != "" {
		log.Printf("connecting to Jaeger at %q", endpoint)
	} else {
		log.Printf("no Jaeger collector found (JAEGER_COLLECTOR_SERVICE_HOST not set)")
	}
	env = serviceenv.InitWithKube(serviceenv.NewConfiguration(config))
	if env.Config().LogFormat == "text" {
		log.SetFormatter(logutil.FormatterFunc(logutil.Pretty))
	}
	profileutil.StartCloudProfiler(service, env.Config())
	debug.SetGCPercent(env.Config().GCPercent)
	if env.Config().EtcdPrefix == "" {
		env.Config().EtcdPrefix = col.DefaultPrefix
	}
	return env, err
}

func setupDB(ctx context.Context, env serviceenv.ServiceEnv) error {
	// TODO: currently all pachds attempt to apply migrations, we should coordinate this
	if err := dbutil.WaitUntilReady(context.Background(), log.StandardLogger(), env.GetDBClient()); err != nil {
		return err
	}
	if err := migrations.ApplyMigrations(context.Background(), env.GetDBClient(), migrations.MakeEnv(nil, env.GetEtcdClient()), clusterstate.DesiredClusterState); err != nil {
		return err
	}
	if err := migrations.BlockUntil(context.Background(), env.GetDBClient(), clusterstate.DesiredClusterState); err != nil {
		return err
	}
	return nil
}

func doReadinessCheck(ctx context.Context, config interface{}) error {
	env := serviceenv.InitPachOnlyEnv(serviceenv.NewConfiguration(config))
	return env.GetPachClient(ctx).Health()
}

func doEnterpriseMode(ctx context.Context, config interface{}) (retErr error) {
	eg, ctx := errgroup.WithContext(ctx)
	defer func() {
		if retErr != nil {
			log.WithError(retErr).Print("failed to start server")
			_ = pprof.Lookup("goroutine").WriteTo(os.Stderr, 2) // swallow error, not much we can do if we can't write to stderr
		}
	}()

	env, err := setup(config, "pachyderm-pachd-enterprise")
	if err != nil {
		return err
	}
	if err := setupDB(ctx, env); err != nil {
		return err
	}
	if !env.Config().EnterpriseMember {
		env.InitDexDB()
	}

	// Setup External Pachd GRPC Server.
	authInterceptor := authmw.NewInterceptor(env.AuthServer)
	loggingInterceptor := loggingmw.NewLoggingInterceptor(env.Logger(), loggingmw.WithLogFormat(env.Config().LogFormat))
	externalServer, err := newExternalServer(ctx, authInterceptor, loggingInterceptor)
	if err != nil {
		return err
	}
	// Setup Internal Pachd GRPC Server.
	internalServer, err := newInternalServer(ctx, authInterceptor, loggingInterceptor)
	if err != nil {
		return err
	}
	var bootstrappers []bootstrapper
	if err := logGRPCServerSetup("External + Internal Enterprise Servers", func() error {
		txnEnv := txnenv.New()
		licenseEnv := licenseserver.EnvFromServiceEnv(env)
		if err := logGRPCServerSetup("License API", func() error {
			licenseAPIServer, err := licenseserver.New(licenseEnv)
			if err != nil {
				return err
			}
			licenseclient.RegisterAPIServer(externalServer.Server, licenseAPIServer)
			licenseclient.RegisterAPIServer(internalServer.Server, licenseAPIServer)
			bootstrappers = append(bootstrappers, licenseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		enterpriseEnv := eprsserver.EnvFromServiceEnv(env, path.Join(env.Config().EtcdPrefix, env.Config().EnterpriseEtcdPrefix), txnEnv)
		if err := logGRPCServerSetup("Enterprise API", func() error {
			enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
				enterpriseEnv,
				true,
			)
			if err != nil {
				return err
			}
			eprsclient.RegisterAPIServer(externalServer.Server, enterpriseAPIServer)
			eprsclient.RegisterAPIServer(internalServer.Server, enterpriseAPIServer)
			env.SetEnterpriseServer(enterpriseAPIServer)
			licenseEnv.EnterpriseServer = enterpriseAPIServer
			bootstrappers = append(bootstrappers, enterpriseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Identity API", func() error {
			idAPIServer := identity_server.NewIdentityServer(
				identity_server.EnvFromServiceEnv(env),
				true,
			)
			identityclient.RegisterAPIServer(externalServer.Server, idAPIServer)
			identityclient.RegisterAPIServer(internalServer.Server, idAPIServer)
			env.SetIdentityServer(idAPIServer)
			bootstrappers = append(bootstrappers, idAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Auth API", func() error {
			authAPIServer, err := authserver.NewAuthServer(
				authserver.EnvFromServiceEnv(env, txnEnv),
				true,
				true,
				true,
			)
			if err != nil {
				return err
			}
			authclient.RegisterAPIServer(externalServer.Server, authAPIServer)
			authclient.RegisterAPIServer(internalServer.Server, authAPIServer)
			env.SetAuthServer(authAPIServer)
			enterpriseEnv.AuthServer = authAPIServer
			bootstrappers = append(bootstrappers, authAPIServer)
			return nil
		}); err != nil {
			return err
		}
		healthServer := health.NewServer()
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		if err := logGRPCServerSetup("Health", func() error {
			grpc_health_v1.RegisterHealthServer(externalServer.Server, healthServer)
			grpc_health_v1.RegisterHealthServer(internalServer.Server, healthServer)
			return nil
		}); err != nil {
			return err
		}

		if err := logGRPCServerSetup("Admin API", func() error {
			adminclient.RegisterAPIServer(externalServer.Server, adminserver.NewAPIServer(adminserver.EnvFromServiceEnv(env)))
			adminclient.RegisterAPIServer(internalServer.Server, adminserver.NewAPIServer(adminserver.EnvFromServiceEnv(env)))
			return nil
		}); err != nil {
			return err
		}

		if err := logGRPCServerSetup("Version API", func() error {
			versionpb.RegisterAPIServer(externalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			versionpb.RegisterAPIServer(internalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			return nil
		}); err != nil {
			return err
		}
		txnEnv.Initialize(env, nil)
		if _, err := internalServer.ListenTCP("", env.Config().PeerPort); err != nil {
			return err
		}
		for _, b := range bootstrappers {
			if err := b.EnvBootstrap(ctx); err != nil {
				return errors.EnsureStack(err)
			}
		}
		if _, err := externalServer.ListenTCP("", env.Config().Port); err != nil {
			return err
		}
		healthServer.Resume()
		return nil
	}); err != nil {
		return err
	}
	// Create the goroutines for the servers.
	// Any server error is considered critical and will cause Pachd to exit.
	// The first server that errors will have its error message logged.
	eg.Go(maybeIgnoreErrorFunc("External Enterprise GRPC Server", true, func() error {
		return externalServer.Wait()
	}))
	eg.Go(maybeIgnoreErrorFunc("Internal Enterprise GRPC Server", true, func() error {
		return internalServer.Wait()
	}))
	return errors.EnsureStack(eg.Wait())
}

func doSidecarMode(ctx context.Context, config interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2) //nolint:errcheck
		}
	}()
	env, err := setup(config, "pachyderm-pachd-sidecar")
	if err != nil {
		return err
	}
	authInterceptor := authmw.NewInterceptor(env.AuthServer)
	loggingInterceptor := loggingmw.NewLoggingInterceptor(env.Logger())
	server, err := newInternalServer(ctx, authInterceptor, loggingInterceptor)
	if err != nil {
		return err
	}
	txnEnv := txnenv.New()
	if err := logGRPCServerSetup("Auth API", func() error {
		authAPIServer, err := authserver.NewAuthServer(
			authserver.EnvFromServiceEnv(env, txnEnv),
			false,
			false,
			false,
		)
		if err != nil {
			return err
		}
		authclient.RegisterAPIServer(server.Server, authAPIServer)
		env.SetAuthServer(authAPIServer)
		return nil
	}); err != nil {
		return err
	}
	if err := logGRPCServerSetup("PFS API", func() error {
		pfsEnv, err := pfs_server.EnvFromServiceEnv(env, txnEnv)
		if err != nil {
			return err
		}
		pfsAPIServer, err := pfs_server.NewSidecarAPIServer(*pfsEnv)
		if err != nil {
			return err
		}
		pfsclient.RegisterAPIServer(server.Server, pfsAPIServer)
		env.SetPfsServer(pfsAPIServer)
		return nil
	}); err != nil {
		return err
	}
	if err := logGRPCServerSetup("PPS API", func() error {
		ppsAPIServer, err := pps_server.NewSidecarAPIServer(
			pps_server.EnvFromServiceEnv(env, txnEnv, nil),
			env.Config().Namespace,
			env.Config().PPSWorkerPort,
			env.Config().PeerPort,
		)
		if err != nil {
			return err
		}
		ppsclient.RegisterAPIServer(server.Server, ppsAPIServer)
		env.SetPpsServer(ppsAPIServer)
		return nil
	}); err != nil {
		return err
	}
	e := eprsserver.EnvFromServiceEnv(env, path.Join(env.Config().EtcdPrefix, env.Config().EnterpriseEtcdPrefix), txnEnv)
	if err := logGRPCServerSetup("Enterprise API", func() error {
		enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
			e,
			false,
		)
		if err != nil {
			return err
		}
		eprsclient.RegisterAPIServer(server.Server, enterpriseAPIServer)
		env.SetEnterpriseServer(enterpriseAPIServer)
		return nil
	}); err != nil {
		return err
	}
	var transactionAPIServer txnserver.APIServer
	if err := logGRPCServerSetup("Transaction API", func() error {
		transactionAPIServer, err = txnserver.NewAPIServer(
			env,
			txnEnv,
		)
		if err != nil {
			return err
		}
		transactionclient.RegisterAPIServer(server.Server, transactionAPIServer)
		return nil
	}); err != nil {
		return err
	}
	if err := logGRPCServerSetup("Health", func() error {
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(server.Server, healthServer)
		return nil
	}); err != nil {
		return err
	}
	if err := logGRPCServerSetup("Debug", func() error {
		debugclient.RegisterDebugServer(server.Server, debugserver.NewDebugServer(
			env,
			env.Config().PachdPodName,
			nil,
			env.GetDBClient(),
		))
		return nil
	}); err != nil {
		return err
	}
	txnEnv.Initialize(env, transactionAPIServer)
	// The sidecar only needs to serve traffic on the peer port, as it only serves
	// traffic from the user container (the worker binary and occasionally user
	// pipelines)
	if _, err := server.ListenTCP("", env.Config().PeerPort); err != nil {
		return err
	}
	return server.Wait()
}

func doFullMode(ctx context.Context, config interface{}) (retErr error) {
	eg, ctx := errgroup.WithContext(ctx)
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		if retErr != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2) //nolint:errcheck
		}
	}()
	env, err := setup(config, "pachyderm-pachd-full")
	if err != nil {
		return err
	}
	if err := setupDB(ctx, env); err != nil {
		return err
	}
	if !env.Config().EnterpriseMember {
		env.InitDexDB()
	}
	var reporter *metrics.Reporter
	if env.Config().Metrics {
		reporter = metrics.NewReporter(env)
	}
	requireNoncriticalServers := !env.Config().RequireCriticalServersOnly
	// Setup External Pachd GRPC Server.
	authInterceptor := authmw.NewInterceptor(env.AuthServer)
	loggingInterceptor := loggingmw.NewLoggingInterceptor(env.Logger())
	externalServer, err := newExternalServer(ctx, authInterceptor, loggingInterceptor)
	if err != nil {
		return err
	}
	// Setup Internal Pachd GRPC Server.
	internalServer, err := newInternalServer(ctx, authInterceptor, loggingInterceptor)
	if err != nil {
		return err
	}
	var bootstrappers []bootstrapper
	if err := logGRPCServerSetup("External + Internal Pachd Servers", func() error {
		txnEnv := txnenv.New()
		licenseEnv := licenseserver.EnvFromServiceEnv(env)
		if err := logGRPCServerSetup("License API", func() error {
			licenseAPIServer, err := licenseserver.New(licenseEnv)
			if err != nil {
				return err
			}
			licenseclient.RegisterAPIServer(externalServer.Server, licenseAPIServer)
			licenseclient.RegisterAPIServer(internalServer.Server, licenseAPIServer)
			bootstrappers = append(bootstrappers, licenseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		enterpriseEnv := eprsserver.EnvFromServiceEnv(env,
			path.Join(env.Config().EtcdPrefix, env.Config().EnterpriseEtcdPrefix),
			txnEnv,
			eprsserver.WithMode(eprsserver.FullMode),
			eprsserver.WithUnpausedMode(os.Getenv("UNPAUSED_MODE")))

		if err := logGRPCServerSetup("Enterprise API", func() error {
			enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
				enterpriseEnv,
				true,
			)
			if err != nil {
				return err
			}
			eprsclient.RegisterAPIServer(externalServer.Server, enterpriseAPIServer)
			eprsclient.RegisterAPIServer(internalServer.Server, enterpriseAPIServer)
			bootstrappers = append(bootstrappers, enterpriseAPIServer)
			env.SetEnterpriseServer(enterpriseAPIServer)
			licenseEnv.EnterpriseServer = enterpriseAPIServer
			return nil
		}); err != nil {
			return err
		}
		if !env.Config().EnterpriseMember {
			if err := logGRPCServerSetup("Identity API", func() error {
				idAPIServer := identity_server.NewIdentityServer(
					identity_server.EnvFromServiceEnv(env),
					true,
				)
				identityclient.RegisterAPIServer(externalServer.Server, idAPIServer)
				identityclient.RegisterAPIServer(internalServer.Server, idAPIServer)
				env.SetIdentityServer(idAPIServer)
				bootstrappers = append(bootstrappers, idAPIServer)
				return nil
			}); err != nil {
				return err
			}
		}
		if err := logGRPCServerSetup("Auth API", func() error {
			authAPIServer, err := authserver.NewAuthServer(
				authserver.EnvFromServiceEnv(env, txnEnv),
				true, requireNoncriticalServers, true,
			)
			if err != nil {
				return err
			}
			authclient.RegisterAPIServer(externalServer.Server, authAPIServer)
			authclient.RegisterAPIServer(internalServer.Server, authAPIServer)
			env.SetAuthServer(authAPIServer)
			enterpriseEnv.AuthServer = authAPIServer
			bootstrappers = append(bootstrappers, authAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("PFS API", func() error {
			pfsEnv, err := pfs_server.EnvFromServiceEnv(env, txnEnv)
			if err != nil {
				return err
			}
			pfsAPIServer, err := pfs_server.NewAPIServer(*pfsEnv)
			if err != nil {
				return err
			}
			pfsclient.RegisterAPIServer(externalServer.Server, pfsAPIServer)
			pfsclient.RegisterAPIServer(internalServer.Server, pfsAPIServer)
			env.SetPfsServer(pfsAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("PPS API", func() error {
			ppsAPIServer, err := pps_server.NewAPIServer(
				pps_server.EnvFromServiceEnv(env, txnEnv, reporter),
			)
			if err != nil {
				return err
			}
			ppsclient.RegisterAPIServer(externalServer.Server, ppsAPIServer)
			ppsclient.RegisterAPIServer(internalServer.Server, ppsAPIServer)
			env.SetPpsServer(ppsAPIServer)
			return nil
		}); err != nil {
			return err
		}

		var transactionAPIServer txnserver.APIServer
		if err := logGRPCServerSetup("Transaction API", func() error {
			transactionAPIServer, err = txnserver.NewAPIServer(
				env,
				txnEnv,
			)
			if err != nil {
				return err
			}
			transactionclient.RegisterAPIServer(externalServer.Server, transactionAPIServer)
			transactionclient.RegisterAPIServer(internalServer.Server, transactionAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Admin API", func() error {
			adminServer := adminserver.NewAPIServer(adminserver.EnvFromServiceEnv(env))
			adminclient.RegisterAPIServer(externalServer.Server, adminServer)
			adminclient.RegisterAPIServer(internalServer.Server, adminServer)
			return nil
		}); err != nil {
			return err
		}
		healthServer := health.NewServer()
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		if err := logGRPCServerSetup("Health", func() error {
			grpc_health_v1.RegisterHealthServer(externalServer.Server, healthServer)
			grpc_health_v1.RegisterHealthServer(internalServer.Server, healthServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Version API", func() error {
			versionpb.RegisterAPIServer(externalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			versionpb.RegisterAPIServer(internalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Debug", func() error {
			debugServer := debugserver.NewDebugServer(
				env,
				env.Config().PachdPodName,
				nil,
				env.GetDBClient(),
			)
			debugclient.RegisterDebugServer(externalServer.Server, debugServer)
			debugclient.RegisterDebugServer(internalServer.Server, debugServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Proxy API", func() error {
			proxyServer := proxyserver.NewAPIServer(proxyserver.Env{
				Listener: env.GetPostgresListener(),
			})
			proxyclient.RegisterAPIServer(externalServer.Server, proxyServer)
			proxyclient.RegisterAPIServer(internalServer.Server, proxyServer)
			return nil
		}); err != nil {
			return err
		}
		txnEnv.Initialize(env, transactionAPIServer)
		if _, err := internalServer.ListenTCP("", env.Config().PeerPort); err != nil {
			return err
		}
		for _, b := range bootstrappers {
			if err := b.EnvBootstrap(ctx); err != nil {
				return errors.EnsureStack(err)
			}
		}
		if _, err := externalServer.ListenTCP("", env.Config().Port); err != nil {
			return err
		}
		healthServer.Resume()
		return nil
	}); err != nil {
		return err
	}
	// Create the goroutines for the servers.
	// Any server error is considered critical and will cause Pachd to exit.
	// The first server that errors will have its error message logged.
	eg.Go(maybeIgnoreErrorFunc("External Pachd GRPC Server", true, func() error {
		return externalServer.Wait()
	}))
	eg.Go(maybeIgnoreErrorFunc("External Pachd GRPC Server", true, func() error {
		return externalServer.Wait()
	}))
	eg.Go(maybeIgnoreErrorFunc("Internal Pachd GRPC Server", true, func() error {
		return internalServer.Wait()
	}))
	eg.Go(maybeIgnoreErrorFunc("S3 Server", requireNoncriticalServers, func() error { return s3Server{env.GetPachClient, env.Config().S3GatewayPort}.listenAndServe(ctx) }))
	eg.Go(maybeIgnoreErrorFunc("Prometheus Server", requireNoncriticalServers, func() error { return prometheusServer{port: env.Config().PrometheusPort}.listenAndServe(ctx) }))
	eg.Go(maybeIgnoreErrorFunc("HTTP PFS Server", requireNoncriticalServers, func() error {
		s := http.Server{
			Addr: fmt.Sprintf(":%v", env.Config().PFSHTTPPort),
			Handler: &pfshttp.Server{
				Client: env.GetPachClient(context.TODO()),
				Logger: log.NewEntry(log.StandardLogger()).WithField("source", "pfshttp"),
			},
		}
		return s.ListenAndServe()
	}))
	go func(c chan os.Signal) {
		<-c
		log.Println("terminating; waiting for pachd server to gracefully stop")
		var g, _ = errgroup.WithContext(ctx)
		g.Go(func() error { externalServer.Server.GracefulStop(); return nil })
		g.Go(func() error { internalServer.Server.GracefulStop(); return nil })
		if err := g.Wait(); err != nil {
			log.Errorf("error waiting for pachd server to gracefully stop: %v", err)
		} else {
			log.Println("gRPC server gracefully stopped")
		}
	}(interruptChan)
	return errors.EnsureStack(eg.Wait())
}

func doPausedMode(ctx context.Context, config interface{}) (retErr error) {
	eg, ctx := errgroup.WithContext(ctx)
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		if retErr != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2) //nolint:errcheck
		}
	}()
	log.Println("starting up in paused mode")
	env, err := setup(config, "pachyderm-pachd-paused")
	if err != nil {
		return err
	}
	if err := setupDB(ctx, env); err != nil {
		return err
	}
	if !env.Config().EnterpriseMember {
		env.InitDexDB()
	}
	requireNoncriticalServers := !env.Config().RequireCriticalServersOnly
	// Setup External Pachd GRPC Server.
	authInterceptor := authmw.NewInterceptor(env.AuthServer)
	loggingInterceptor := loggingmw.NewLoggingInterceptor(env.Logger())
	externalServer, err := newExternalServer(ctx, authInterceptor, loggingInterceptor)
	if err != nil {
		return err
	}
	// Setup Internal Pachd GRPC Server.
	internalServer, err := newInternalServer(ctx, authInterceptor, loggingInterceptor)
	if err != nil {
		return err
	}
	if err := logGRPCServerSetup("External + Internal Pachd Services", func() error {
		txnEnv := txnenv.New()
		licenseEnv := licenseserver.EnvFromServiceEnv(env)
		if err := logGRPCServerSetup("License API", func() error {
			licenseAPIServer, err := licenseserver.New(licenseEnv)
			if err != nil {
				return err
			}
			licenseclient.RegisterAPIServer(externalServer.Server, licenseAPIServer)
			licenseclient.RegisterAPIServer(internalServer.Server, licenseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		enterpriseEnv := eprsserver.EnvFromServiceEnv(env,
			path.Join(env.Config().EtcdPrefix,
				env.Config().EnterpriseEtcdPrefix),
			txnEnv, eprsserver.WithMode(eprsserver.PausedMode),
			eprsserver.WithUnpausedMode(os.Getenv("UNPAUSED_MODE")))
		if err := logGRPCServerSetup("Enterprise API", func() error {
			enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
				enterpriseEnv,
				true,
			)
			if err != nil {
				return err
			}
			eprsclient.RegisterAPIServer(externalServer.Server, enterpriseAPIServer)
			eprsclient.RegisterAPIServer(internalServer.Server, enterpriseAPIServer)
			env.SetEnterpriseServer(enterpriseAPIServer)
			licenseEnv.EnterpriseServer = enterpriseAPIServer
			// Stop workers because unpaused pachds in the process
			// of rolling may have started them back up.
			if err := enterpriseEnv.StopWorkers(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Admin API", func() error {
			adminServer := adminserver.NewAPIServer(adminserver.EnvFromServiceEnv(env))
			adminclient.RegisterAPIServer(externalServer.Server, adminServer)
			adminclient.RegisterAPIServer(internalServer.Server, adminServer)
			return nil
		}); err != nil {
			return err
		}
		if !env.Config().EnterpriseMember {
			if err := logGRPCServerSetup("Identity API", func() error {
				idAPIServer := identity_server.NewIdentityServer(
					identity_server.EnvFromServiceEnv(env),
					true,
				)
				identityclient.RegisterAPIServer(externalServer.Server, idAPIServer)
				identityclient.RegisterAPIServer(internalServer.Server, idAPIServer)
				env.SetIdentityServer(idAPIServer)
				return nil
			}); err != nil {
				return err
			}
		}
		if err := logGRPCServerSetup("Auth API", func() error {
			authAPIServer, err := authserver.NewAuthServer(
				authserver.EnvFromServiceEnv(env, txnEnv),
				true, requireNoncriticalServers, true,
			)
			if err != nil {
				return err
			}
			authclient.RegisterAPIServer(externalServer.Server, authAPIServer)
			authclient.RegisterAPIServer(internalServer.Server, authAPIServer)
			env.SetAuthServer(authAPIServer)
			return nil
		}); err != nil {
			return err
		}
		healthServer := health.NewServer()
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		if err := logGRPCServerSetup("Health", func() error {
			grpc_health_v1.RegisterHealthServer(externalServer.Server, healthServer)
			grpc_health_v1.RegisterHealthServer(internalServer.Server, healthServer)
			return nil
		}); err != nil {
			return err
		}
		var transactionAPIServer txnserver.APIServer
		if err := logGRPCServerSetup("Transaction API", func() error {
			transactionAPIServer, err = txnserver.NewAPIServer(
				env,
				txnEnv,
			)
			if err != nil {
				return err
			}
			transactionclient.RegisterAPIServer(externalServer.Server, transactionAPIServer)
			transactionclient.RegisterAPIServer(internalServer.Server, transactionAPIServer)
			return nil
		}); err != nil {
			return err
		}
		txnEnv.Initialize(env, transactionAPIServer)
		if _, err := internalServer.ListenTCP("", env.Config().PeerPort); err != nil {
			return err
		}
		if _, err := externalServer.ListenTCP("", env.Config().Port); err != nil {
			return err
		}
		healthServer.Resume()
		return nil
	}); err != nil {
		return err
	}
	// Create the goroutines for the servers.
	// Any server error is considered critical and will cause Pachd to exit.
	// The first server that errors will have its error message logged.
	eg.Go(maybeIgnoreErrorFunc("External Pachd GRPC Server", true, func() error {
		return externalServer.Wait()
	}))
	eg.Go(maybeIgnoreErrorFunc("Internal Pachd GRPC Server", true, func() error {
		return internalServer.Wait()
	}))
	eg.Go(maybeIgnoreErrorFunc("S3 Server", requireNoncriticalServers, func() error { return s3Server{env.GetPachClient, env.Config().S3GatewayPort}.listenAndServe(ctx) }))
	eg.Go(maybeIgnoreErrorFunc("Prometheus Server", requireNoncriticalServers, func() error { return prometheusServer{port: env.Config().PrometheusPort}.listenAndServe(ctx) }))
	go func(c chan os.Signal) {
		<-c
		log.Println("terminating; waiting for paused pachd server to gracefully stop")
		var g, _ = errgroup.WithContext(ctx)
		g.Go(func() error { externalServer.Server.GracefulStop(); return nil })
		g.Go(func() error { internalServer.Server.GracefulStop(); return nil })
		if err := g.Wait(); err != nil {
			log.Errorf("error waiting for paused pachd server to gracefully stop: %v", err)
		} else {
			log.Println("gRPC server gracefully stopped")
		}
	}(interruptChan)
	return errors.EnsureStack(eg.Wait())
}

func logGRPCServerSetup(name string, f func() error) (retErr error) {
	log.Printf("started setting up %v GRPC Server", name)
	defer func() {
		if retErr != nil {
			retErr = errors.Wrapf(retErr, "error setting up %v GRPC Server", name)
		} else {
			log.Printf("finished setting up %v GRPC Server", name)
		}
	}()
	return f()
}

// maybeIgnoreErrorFunc returns a function that runs f; if f returns an HTTP server
// closed error it returns nil; if required is false it returns nil; otherwise
// it returns whatever f returns.
func maybeIgnoreErrorFunc(name string, required bool, f func() error) func() error {
	return func() error {
		if err := f(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			if !required {
				log.Errorf("error setting up and/or running %v: %v", name, err)
				return nil
			}
			return errors.Wrapf(err, "error setting up and/or running %v (use --require-critical-servers-only deploy flag to ignore errors from noncritical servers)", name)
		}
		return nil
	}
}

type s3Server struct {
	clientFactory s3.ClientFactory
	port          uint16
}

// listenAndServe listens until ctx is cancelled; it then gracefully shuts down
// the server, returning once all requests have been handled.
func (ss s3Server) listenAndServe(ctx context.Context) error {
	var (
		router = s3.Router(s3.NewMasterDriver(), ss.clientFactory)
		srv    = s3.Server(ss.port, router)
		errCh  = make(chan error, 1)
	)
	srv.BaseContext = func(net.Listener) context.Context { return ctx }
	go func() {
		var (
			certPath, keyPath, err = tls.GetCertPaths()
		)
		if err != nil {
			log.Warnf("s3gateway TLS disabled: %v", err)
			errCh <- errors.EnsureStack(srv.ListenAndServe())
		}
		cLoader := tls.NewCertLoader(certPath, keyPath, tls.CertCheckFrequency)
		// Read TLS cert and key
		err = cLoader.LoadAndStart()
		if err != nil {
			errCh <- errors.Wrapf(err, "couldn't load TLS cert for s3gateway: %v", err)
		}
		srv.TLSConfig = &gotls.Config{GetCertificate: cLoader.GetCertificate}
		errCh <- errors.EnsureStack(srv.ListenAndServeTLS(certPath, keyPath))
	}()
	select {
	case <-ctx.Done():
		// NOTE: using context.Background here means that shutdown will
		// wait until all requests terminate.
		log.Info("terminating S3 server due to cancelled context")
		return errors.EnsureStack(srv.Shutdown(context.Background()))
	case err := <-errCh:
		return err
	}
}

type prometheusServer struct {
	port uint16
}

// listenAndServe listens until ctx is cancelled; it then gracefully shuts down
// the server, returning once all requests have been handled.
func (ps prometheusServer) listenAndServe(ctx context.Context) error {
	var (
		mux = http.NewServeMux()
		srv = http.Server{
			Addr:        fmt.Sprintf(":%v", ps.port),
			Handler:     mux,
			BaseContext: func(net.Listener) context.Context { return ctx },
		}
		errCh = make(chan error, 1)
	)
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		// NOTE: using context.Background here means that shutdown will
		// wait until all requests terminate.
		log.Info("terminating S3 server due to cancelled context")
		return errors.EnsureStack(srv.Shutdown(context.Background()))
	case err := <-errCh:
		return err
	}

}
