package db

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/config"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"time"
)

type DB interface {
	AddTasks(tasks []Task) error
	DeleteTask(taskId uuid.UUID) error
	GetShards(bucketId int64) ([]Shard, error)
	GetTaskById(taskId uuid.UUID) (*Task, error)
	GetUnpickedTasksByShardId(shardId uuid.UUID) ([]Task, error)
	UpdateTaskStatus(taskId uuid.UUID, status Status) error
}

type db struct {
	client       *sqlx.DB
	maxShardSize int64
}

var EMPTY_SHARD Shard = Shard{}

func (d *db) AddTasks(tasks []Task) error {
	tx := d.client.MustBegin()
	data := map[int64][]Task{}
	for _, task := range tasks {
		executionTime := task.TimeForExecution.Unix()
		window := executionTime - (executionTime % 60)
		if _, ok := data[window]; !ok {
			data[window] = []Task{}
		}
		data[window] = append(data[window], task)
	}
	for window, tasks := range data {
		shard := EMPTY_SHARD
		err := d.client.Get(&shard, "SELECT * FROM shards where bucket_id = $1 ORDER BY created_on DESC LIMIT 1", window)
		if err != nil && err != sql.ErrNoRows{
			tx.Rollback()
			return errors.Wrap(err,"Error in getting shards by bucket id")
		}
		count := int64(0)
		remainingSlotsInCurrentShard := int64(0)
		if shard != EMPTY_SHARD {
			err = d.client.Get(&count, "SELECT COUNT(*) FROM tasks where shard_id = $1", shard.Id)
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err,"Error in getting tasks by shard id")
			}
			remainingSlotsInCurrentShard = d.maxShardSize - count
		}
		for _, task := range tasks {
			if shard == EMPTY_SHARD || remainingSlotsInCurrentShard == 0 {
				shard = Shard{
					Id:        uuid.NewV4(),
					BucketId:  window,
					CreatedOn: time.Now(),
				}
				_, err := tx.NamedExec("INSERT INTO shards(id,bucket_id,created_on) VALUES(:id,:bucket_id,:created_on)", &shard)
				if err != nil {
					tx.Rollback()
					return errors.Wrap(err,"Error in inserting shards")
				}
				remainingSlotsInCurrentShard = d.maxShardSize
			}
			task.Id = uuid.NewV4()
			task.ShardId = shard.Id
			task.Status = UNPICKED
			_, err := tx.NamedExec("INSERT INTO tasks(id, shard_id, status,source,destination,time_for_execution,content) VALUES (:id, :shard_id, :status,:source,:destination,:time_for_execution,:content)", &task)
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err,"Errors in inserting task")
			}
			remainingSlotsInCurrentShard--
		}
	}
	return tx.Commit()
}

func (d *db) GetTaskById(taskId uuid.UUID) (*Task, error) {
	task := Task{}
	err := d.client.Select(&task, "SELECT * FROM tasks WHERE id = $1", taskId)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (d *db) DeleteTask(taskId uuid.UUID) error {
	tx := d.client.MustBegin()
	tx.MustExec("DELETE FROM tasks WHERE id = $1", taskId)
	return tx.Commit()
}

func (d *db) GetShards(bucketId int64) ([]Shard, error) {
	var shards []Shard
	err := d.client.Select(&shards, "SELECT * FROM shards WHERE bucket_id = $1", bucketId)
	if err != nil {
		return nil, err
	}
	return shards, nil
}

func (d *db) GetUnpickedTasksByShardId(shardId uuid.UUID) ([]Task, error) {
	var tasks []Task
	err := d.client.Select(&tasks, "SELECT * FROM tasks WHERE shard_id = $1 AND status = $2", shardId, UNPICKED)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (d *db) UpdateTaskStatus(taskId uuid.UUID, status Status) error {
	tx := d.client.MustBegin()
	tx.MustExec("UPDATE tasks SET status = $1 WHERE id = $2", status, taskId)
	return tx.Commit()
}

func populateTasksTable(dbClient DB) error {
	var tasks []Task
	j,k,current := 0, 1,time.Now()
	for i := 0; i < 100; i++ {
		if j == 5 {
			j = 0
			k++
		}
		tasks = append(tasks, Task{
			Source:           fmt.Sprintf("source-%d", i),
			Destination:      fmt.Sprintf("target-%d", i),
			TimeForExecution: current.Add(time.Duration(60 * k) * time.Second),
			Content:          "Sample Content",
		})
		j++
	}
	return dbClient.AddTasks(tasks)
}

func runMigrations(host string, port int, username string, password string, dbName string, connection DB) error {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", username, password, host, port, dbName))
	defer db.Close()
	migrationsPath := "file://migrations"
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return errors.Wrap(err, "Migration object creation failed")
	}
	m.Up()
	return populateTasksTable(connection)
}

func InitDB(host string, port int, username string, password string, dbName string,maxShardSize int64, connectionTimeout, maxConnectionLifetime, maxIdleConnections, maxOpenConnections int) (DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s connect_timeout=%d sslmode=disable",
		host, port, username, password, dbName, connectionTimeout)
	dbConnection, err := sqlx.Connect(config.DBDriver(), connectionString)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to connect to postgres db")
	}
	dbConnection.SetConnMaxLifetime(time.Duration(maxConnectionLifetime))
	dbConnection.SetMaxIdleConns(maxIdleConnections)
	dbConnection.SetMaxOpenConns(maxOpenConnections)
	db := &db{
		client: dbConnection,
		maxShardSize: maxShardSize,
	}
	err = runMigrations(host, port, username, password, dbName, db)
	if err != nil {
		return nil, errors.Wrap(err, "Error occurred while running migrations")
	}
	return db, nil
}
