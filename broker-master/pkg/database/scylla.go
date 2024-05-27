package database

import (
	"context"
	"fmt"
	"sync"
	"therealbroker/config"
	"therealbroker/pkg/broker"
	"time"

	"github.com/gocql/gocql"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

var (
	scyllaDb      = &ScyllaDB{}
	onceScylla    = sync.Once{}
	errConnScylla error
)

type ScyllaDB struct {
	cfg            *config.Config
	log            *logrus.Logger
	session        *gocql.Session
	batch          *batchOperation
	lastMessageId  int
	handleMSgMutex sync.Mutex
}

func ConnectToScylla(log *logrus.Logger) (DB, error) {
	scyllaConfig := config.GetConfigInstance()
	onceScylla.Do(func() {
		cluster := gocql.NewCluster(scyllaConfig.ScyllaDB.Host)
		cluster.Port = scyllaConfig.ScyllaDB.Port
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: scyllaConfig.ScyllaDB.Username,
			Password: scyllaConfig.ScyllaDB.Password,
		}

		session, err := cluster.CreateSession()
		if err != nil {
			log.Fatalln(err)
		}
		scyllaDb = &ScyllaDB{
			cfg:            scyllaConfig,
			log:            log,
			session:        session,
			handleMSgMutex: sync.Mutex{},
			batch: &batchOperation{
				count:      0,
				batch:      session.NewBatch(gocql.UnloggedBatch),
				batchMutex: sync.Mutex{},
			},
		}

		err = scyllaDb.createKeyspace()
		if err != nil {
			errConnScylla = err
		}
		scyllaDb.log.Infof("scylla keyspace %s has been created successfully\n", scyllaDb.cfg.ScyllaDB.Keyspace)

		err = scyllaDb.createTable()
		if err != nil {
			errConnScylla = err
		} else {
			scyllaDb.log.Infoln("scylla messages table has been created successfully")
		}

		err = scyllaDb.loadLastId()
		if err != nil {
			errConnScylla = err
		} else {
			scyllaDb.log.Infoln("last message id has been found successfully")
		}

		go scyllaDb.scheduledBatchOperation()
	})
	return scyllaDb, errConnScylla
}

func (sd *ScyllaDB) createKeyspace() error {
	keyspace := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };",
		sd.cfg.ScyllaDB.Keyspace)
	err := sd.session.Query(keyspace).Exec()
	return err

}
func (sd *ScyllaDB) createTable() error {
	table := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s.messages (
        id INT,
        subject TEXT,
        body BLOB,
        expiration_time BIGINT,
        added_time TIMESTAMP,
        removed BOOLEAN,
        PRIMARY KEY (subject, id)
    );`, sd.cfg.ScyllaDB.Keyspace,
	)

	err := sd.session.Query(table).Exec()
	return err
}

func (sd *ScyllaDB) loadLastId() error {
	var lastId int

	query := fmt.Sprintf("SELECT MAX(id) FROM %s.messages;", sd.cfg.ScyllaDB.Keyspace)
	if err := sd.session.Query(query).Scan(&lastId); err != nil {
		if err == gocql.ErrNotFound {
			lastId = 0
		} else {
			return err
		}
	}
	sd.handleMSgMutex.Lock()
	sd.lastMessageId = lastId
	sd.handleMSgMutex.Unlock()
	return nil
}

func GetScyllaInstance() DB {
	return scyllaDb
}

func (sd *ScyllaDB) AddMessage(ctx context.Context, newMsg broker.Message, subject string) (int, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Add new message to scylla")
	defer span.Finish()

	sd.handleMSgMutex.Lock()
	sd.lastMessageId++
	var newId = sd.lastMessageId
	var expired = newMsg.Expiration == time.Duration(0)
	query := fmt.Sprintf(`
	INSERT INTO %s.messages (id, subject, body, expiration_time, added_time, removed) VALUES (?, ?, ?, ?, toTimestamp(now()), ?)
	`, sd.cfg.CassandraDB.Keyspace)
	sd.handleMSgMutex.Unlock()

	sd.addQueryToBatch(query, newId, subject, []byte(newMsg.Body), int64(newMsg.Expiration), expired)

	return newId, nil
}

func (sd *ScyllaDB) FetchMessage(ctx context.Context, id int, subject string) (broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Fetch message from scylla")
	defer span.Finish()

	query := fmt.Sprintf(`
		SELECT body, expiration_time FROM %s.messages WHERE subject = '%s' AND id = %d;
	`, sd.cfg.ScyllaDB.Keyspace, subject, id)

	rows := sd.session.Query(query).WithContext(ctx).Iter()

	var messages broker.Message
	var body []byte
	var expration_time int64
	for rows.Scan(&body, &expration_time) {
		messages = broker.Message{
			Body:       string(body),
			Expiration: time.Duration(expration_time),
		}
	}

	err := rows.Close()
	return messages, err
}

func (sd *ScyllaDB) DeleteMessage(subject string, id int) {
	span, _ := opentracing.StartSpanFromContext(context.Background(), "Delete message from scylla")
	defer span.Finish()

	query := fmt.Sprintf(`
	UPDATE %s.messages SET removed = true WHERE subject = '%s' AND id = %d;
	`, sd.cfg.ScyllaDB.Keyspace, subject, id)
	sd.addQueryToBatch(query)
}

func (sd *ScyllaDB) GetMessagesBySubject(ctx context.Context, subject string) ([]broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetMessages based on the given subject from scylla")
	defer span.Finish()

	query := fmt.Sprintf(`
		SELECT body, expiration_time FROM %s.messages WHERE subject = '%s';
	`, sd.cfg.ScyllaDB.Keyspace, subject)

	rows := sd.session.Query(query).Iter()

	var messages = make([]broker.Message, 0)
	var body []byte
	var expration_time int64
	for rows.Scan(&body, &expration_time) {
		messages = append(messages, broker.Message{
			Body:       string(body),
			Expiration: time.Duration(expration_time),
		})
	}

	err := rows.Close()
	return messages, err
}

func (sd *ScyllaDB) Close() error {
	if sd.session != nil {
		sd.session.Close()
	}
	return nil
}

func (sd *ScyllaDB) scheduledBatchOperation() {
	ticker := time.NewTicker(time.Duration(5 * sd.cfg.ScyllaDB.TimeThreshold))
	defer ticker.Stop()

	for range ticker.C {
		sd.batch.batchMutex.Lock()
		if sd.batch.count > 0 {
			sd.execBatch()
		}
		sd.batch.batchMutex.Unlock()
	}
}

func (sd *ScyllaDB) addQueryToBatch(query string, args ...interface{}) {
	sd.batch.batchMutex.Lock()
	defer sd.batch.batchMutex.Unlock()

	sd.batch.batch.Query(query, args...)
	sd.batch.count++
	if sd.batch.count == sd.cfg.ScyllaDB.BatchSize {
		sd.execBatch()
	}

}

func (sd *ScyllaDB) execBatch() {
	if sd.batch.count == 0 {
		return
	}

	err := sd.session.ExecuteBatch(sd.batch.batch)
	if err != nil {
		sd.log.WithError(err).Warn("could not execute batch operation")
		return
	}
	sd.batch.count = 0
	sd.batch.batch = sd.session.NewBatch(gocql.UnloggedBatch)
}
