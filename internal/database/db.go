package db

import (
	"context"
	"fmt"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"reflect"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	natting "github.com/docker/go-connections/nat"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Connect() {
	config := config.GetConfig()
	connectedDb, err := gorm.Open(mysql.Open(config.ConnectionString), &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}

	db = connectedDb
}

func MakeMigrations() {
	db.AutoMigrate(&models.RefreshToken{}, &models.User{}, &models.Receipt{}, &models.Item{}, &models.FileData{}, &models.Tag{}, &models.Category{})
}

func GetDB() *gorm.DB {
	return db
}

func InitTestDb() string {
	envVals := make([]string, 0)
	envMap := make(map[string]string)
	envMap["MYSQL_ROOT_PASSWORD"] = "123456"
	envMap["MYSQL_USER"] = "wrangler-test"
	envMap["MYSQL_PASSWORD"] = "123456"
	envMap["MYSQL_DATABASE"] = "wrangler-test"

	for _, i := range reflect.ValueOf(envMap).MapKeys() {
		val := i.Interface().(string)
		envVals = append(envVals, val+"="+envMap[val])
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println("Unable to create docker client")
		panic(err)
	}

	hostBinding := nat.PortBinding{
		HostIP:   "localhost",
		HostPort: "9002",
	}
	containerPort, err := nat.NewPort("tcp", "3306")
	if err != nil {
		panic("Unable to get the port")
	}
	exposedPorts := map[natting.Port]struct{}{
		containerPort: struct{}{},
	}

	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	cont, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        "library/mariadb",
			Env:          envVals,
			ExposedPorts: exposedPorts,
		},
		&container.HostConfig{
			PortBindings: portBinding,
		}, nil, nil, "receipt-wrangler-test-db")
	if err != nil {
		panic(err)
	}

	cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	time.Sleep(5 * time.Second)
	fmt.Printf("Container %s is ready", cont.ID)

	return cont.ID
}

func TeardownTestDb(containerId string) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	dur, _ := time.ParseDuration("5s")
	err = cli.ContainerStop(context.Background(), containerId, &dur)
	if err != nil {
		panic(err)
	}

	err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
}
