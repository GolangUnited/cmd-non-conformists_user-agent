package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	config "github.com/golang-unitied-school/useragent/config"
	api "github.com/golang-unitied-school/useragent/internal/api/v1"
	dbFace "github.com/golang-unitied-school/useragent/internal/interfaces"
	user "github.com/golang-unitied-school/useragent/internal/repositories/users"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var appCfg config.Config

func init() {

	log.Println("get app config..")

	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}
	appCfg = *config.GetConfig()

}

func initDatabase(cfg config.DatabaseConfig) dbFace.UserDataManager {

	var dbConn dbFace.UserDataManager

	//here you can describe any variant og db connections (db face implementation)
	switch cfg.DB_TYPE {
	case "Postgres":
		dbConn = new(user.PGSQL)
	default:
		log.Fatal("Database implementation not found")
	}
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=%s",
		cfg.DB_HOST,
		cfg.DB_PORT,
		cfg.DB_NAME,
		cfg.DB_USER,
		cfg.DB_PASS,
		cfg.OTHER_P["DB_SSLMODE"],
		cfg.OTHER_P["DB_TZ"])

	log.Println("start db connection..")
	dbConn.Init(dsn)
	log.Println("connection opened!")
	return dbConn
}

func main() {
	conf := config.GetConfig()

	dbConn := initDatabase(conf.DBConfig)

	log.Println("starting grpc server...")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := grpc.NewServer()
	grpcsrv := &api.UserAgent{DBConn: dbConn}
	api.RegisterUserAgentServer(srv, grpcsrv)

	go func() {

		var port = "8080"
		if conf.TCPPort != "" {
			port = conf.TCPPort
		}
		log.Printf("starting server on port %s", port)
		listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("serving..")
		if err := srv.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	log.Print("Server Stopping..")

	err := dbConn.Close()
	if err != nil {
		log.Fatal(err)
	}

	srv.GracefulStop()

	log.Print("Server stopped")

	defer func() {
		dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
		srv.Stop()
	}()
}
