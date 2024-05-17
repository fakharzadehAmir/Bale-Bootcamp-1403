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
	cassandraDb      = &CassandraDB{}
	onceCassandra    = sync.Once{}
	errConnCassandra error
)

type batchOperation struct {
	count      int
	batch      *gocql.Batch
	batchMutex sync.Mutex
}
type CassandraDB struct {
	cfg            *config.Config
	log            *logrus.Logger
	session        *gocql.Session
	batch          *batchOperation
	lastMessageId  int
	handleMSgMutex sync.Mutex
}

func ConnectToCassandra(log *logrus.Logger) (DB, error) {
	cassandraConfig := config.GetConfigInstance()
	onceCassandra.Do(func() {
		cluster := gocql.NewCluster(cassandraConfig.CassandraDB.Host)
		cluster.Port = cassandraConfig.CassandraDB.Port
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cassandraConfig.CassandraDB.Username,
			Password: cassandraConfig.CassandraDB.Password,
		}

		session, err := cluster.CreateSession()
		if err != nil {
			log.Fatalln(err)
		}
		cassandraDb = &CassandraDB{
			cfg:            cassandraConfig,
			log:            log,
			session:        session,
			handleMSgMutex: sync.Mutex{},
			batch: &batchOperation{
				count:      0,
				batch:      session.NewBatch(gocql.UnloggedBatch),
				batchMutex: sync.Mutex{},
			},
		}

		err = cassandraDb.createKeyspace()
		if err != nil {
			errConnCassandra = err
		}
		cassandraDb.log.Infof("cassandra keyspace %s has been created successfully\n", cassandraDb.cfg.CassandraDB.Keyspace)

		err = cassandraDb.createTable()
		if err != nil {
			errConnCassandra = err
		} else {
			cassandraDb.log.Infoln("cassandra messages table has been created successfully")
		}

		err = cassandraDb.loadLastId()
		if err != nil {
			errConnCassandra = err
		} else {
			cassandraDb.log.Infoln("last message id has been found successfully")
		}

		go cassandraDb.scheduledBatchOperation()
	})
	return cassandraDb, errConnCassandra
}

func (cd *CassandraDB) createKeyspace() error {
	keyspace := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };",
		cd.cfg.CassandraDB.Keyspace)
	err := cd.session.Query(keyspace).Exec()
	return err

}
func (cd *CassandraDB) createTable() error {
	table := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s.messages (
        id INT,
        subject TEXT,
        body BLOB,
        expiration_time BIGINT,
        added_time TIMESTAMP,
        removed BOOLEAN,
        PRIMARY KEY (subject, id)
    );`, cd.cfg.CassandraDB.Keyspace,
	)

	err := cd.session.Query(table).Exec()
	return err
}

func (cd *CassandraDB) loadLastId() error {
	var lastId int

	query := fmt.Sprintf("SELECT MAX(id) FROM %s.messages;", cd.cfg.CassandraDB.Keyspace)
	if err := cd.session.Query(query).Scan(&lastId); err != nil {
		if err == gocql.ErrNotFound {
			lastId = 0
		} else {
			return err
		}
	}
	cd.handleMSgMutex.Lock()
	cd.lastMessageId = lastId
	cd.handleMSgMutex.Unlock()
	return nil
}

func GetCassandraInstance() DB {
	return cassandraDb
}

func (cd *CassandraDB) AddMessage(ctx context.Context, newMsg broker.Message, subject string) (int, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Add new message to cassandra")
	defer span.Finish()

	cd.handleMSgMutex.Lock()
	cd.lastMessageId++
	var newId = cd.lastMessageId
	var expired = newMsg.Expiration == time.Duration(0)
	query := fmt.Sprintf(`
	INSERT INTO %s.messages (id, subject, body, expiration_time, added_time, removed) VALUES (?, ?, ?, ?, toTimestamp(now()), ?)
	`, cd.cfg.CassandraDB.Keyspace)
	cd.handleMSgMutex.Unlock()

	cd.addQueryToBatch(query, newId, subject, []byte(newMsg.Body), int64(newMsg.Expiration), expired)

	return newId, nil
}

func (cd *CassandraDB) FetchMessage(ctx context.Context, id int, subject string) (broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Fetch message from cassandra")
	defer span.Finish()

	query := fmt.Sprintf(`
		SELECT body, expiration_time FROM %s.messages WHERE subject = '%s' AND id = %d;
	`, cd.cfg.CassandraDB.Keyspace, subject, id)

	rows := cd.session.Query(query).WithContext(ctx).Iter()

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

func (cd *CassandraDB) DeleteMessage(subject string, id int) {
	span, _ := opentracing.StartSpanFromContext(context.Background(), "Delete message from cassandra")
	defer span.Finish()

	query := fmt.Sprintf(`
	UPDATE %s.messages SET removed = true WHERE subject = '%s' AND id = %d;
	`, cd.cfg.CassandraDB.Keyspace, subject, id)
	cd.addQueryToBatch(query)
}

func (cd *CassandraDB) GetMessagesBySubject(ctx context.Context, subject string) ([]broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetMessages based on the given subject from cassandra")
	defer span.Finish()

	query := fmt.Sprintf(`
		SELECT body, expiration_time FROM %s.messages WHERE subject = '%s';
	`, cd.cfg.CassandraDB.Keyspace, subject)

	rows := cd.session.Query(query).Iter()

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

func (cd *CassandraDB) Close() error {
	if cd.session != nil {
		cd.session.Close()
	}
	return nil
}

func (cd *CassandraDB) scheduledBatchOperation() {
	ticker := time.NewTicker(time.Duration(5 * cd.cfg.CassandraDB.TimeThreshold))
	defer ticker.Stop()

	for range ticker.C {
		cd.batch.batchMutex.Lock()
		if cd.batch.count > 0 {
			cd.execBatch()
		}
		cd.batch.batchMutex.Unlock()
	}
}

func (cd *CassandraDB) addQueryToBatch(query string, args ...interface{}) {
	cd.batch.batchMutex.Lock()
	defer cd.batch.batchMutex.Unlock()

	cd.batch.batch.Query(query, args...)
	cd.batch.count++
	if cd.batch.count == cd.cfg.CassandraDB.BatchSize {
		cd.execBatch()
	}

}

func (cd *CassandraDB) execBatch() {
	if cd.batch.count == 0 {
		return
	}

	err := cd.session.ExecuteBatch(cd.batch.batch)
	if err != nil {
		cd.log.WithError(err).Warn("could not execute batch operation")
		return
	}
	cd.batch.count = 0
	cd.batch.batch = cd.session.NewBatch(gocql.UnloggedBatch)
}
