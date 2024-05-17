package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"therealbroker/config"
	"therealbroker/pkg/broker"
	"time"

	_ "github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

var (
	pgDatabase = &PostgresDB{}
	oncePg     = &sync.Once{}
	errConnPg  error
)

type PostgresDB struct {
	cfg          *config.Config
	log          *logrus.Logger
	conn         *sql.DB
	deletionList []string
	// insertMessages []broker.Message
	sync.RWMutex
}

func ConnectToPg(ctx context.Context, cfg *config.Config, logger *logrus.Logger) (DB, error) {

	connString := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		cfg.PostgresDB.Host, cfg.PostgresDB.Port, cfg.PostgresDB.Username, cfg.PostgresDB.Password, cfg.PostgresDB.DbName)

	oncePg.Do(func() {
		//	Connect to database
		conn, errConnPg := sql.Open("postgres", connString)
		if errConnPg != nil {
			fmt.Println(errConnPg)
			return
		}
		conn.SetMaxOpenConns(90)
		conn.SetConnMaxIdleTime(45)
		conn.SetConnMaxIdleTime(1 * time.Second)

		pgDatabase = &PostgresDB{
			cfg:          cfg,
			log:          logger,
			conn:         conn,
			deletionList: make([]string, 0),
		}
		//	Create Table
		errConnPg = pgDatabase.createTable()
		if errConnPg != nil {
			pgDatabase.log.WithError(errConnPg).Warn("could not create messages table")
			return
		}
		pgDatabase.log.Infoln("messages table has been created successfully")

		//	Create Index
		errConnPg = pgDatabase.createIndex()
		if errConnPg != nil {
			pgDatabase.log.WithError(errConnPg).Warn("could not create index on messages subject")
			return
		}
		pgDatabase.log.Infoln("messages index has been created successfully")

		errConnPg = pgDatabase.updateExpiredMessages()
		if errConnPg != nil {
			pgDatabase.log.WithError(errConnPg).Warn("could not update removed column for expired messages after starting")
			return
		}
		pgDatabase.log.Infoln("expired messages has been marked successfully")

		go pgDatabase.scheduledBatchDeletion()
	})
	return pgDatabase, errConnPg
}

func GetDatabaseInstance() DB {
	return pgDatabase
}

func (pd *PostgresDB) createTable() error {
	table := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		subject VARCHAR(255) NOT NULL,
		body BYTEA,
		expiration_time BIGINT NOT NULL,
		added_time TIMESTAMP NOT NULL,
		removed BOOL
	);
	`
	_, err := pd.conn.Exec(table)
	return err
}

func (pd *PostgresDB) createIndex() error {
	index := `CREATE UNIQUE INDEX IF NOT EXISTS idx_subject ON messages (id, subject);`
	_, err := pd.conn.Exec(index)
	return err
}

func (pd *PostgresDB) updateExpiredMessages() error {
	query := `
        UPDATE messages
        SET removed = TRUE
        WHERE added_time + (expiration_time * INTERVAL '1 second') < NOW()
        AND removed = FALSE;
    `
	_, err := pd.conn.Exec(query)
	return err
}

func (pd *PostgresDB) Close() error {
	if pd.conn != nil {
		return pd.conn.Close()
	}
	return nil
}

func (pd *PostgresDB) AddMessage(ctx context.Context, msg broker.Message, subject string) (int, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Add new message to postgresql")
	defer span.Finish()

	pd.Lock()
	defer pd.Unlock()
	var insertID = -1
	var expired = msg.Expiration == time.Duration(0)
	query := `INSERT INTO messages (subject, body, expiration_time, added_time, removed) VALUES ($1, $2, $3, NOW(), $4) RETURNING id;`

	row := pd.conn.QueryRow(query, subject, []byte(msg.Body), int64(msg.Expiration), expired)
	err := row.Scan(&insertID)
	if err != nil {
		pd.log.WithError(err).Warn("failed in inserting message")
		return insertID, err
	}
	return insertID, nil
}

func (pd *PostgresDB) FetchMessage(ctx context.Context, id int, subject string) (broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Fetch message from postgresql")
	defer span.Finish()

	pd.RLock()
	stringId := strconv.Itoa(id)
	for _, delId := range pd.deletionList {

		if delId == stringId {

			pd.RUnlock()
			return broker.Message{}, broker.ErrExpiredID
		}
	}
	pd.RUnlock()

	query := fmt.Sprintf("SELECT body, expiration_time, removed FROM messages WHERE id = %d AND subject = '%s';", id, subject)
	rows, err := pd.conn.Query(query)
	if err != nil {
		pd.log.WithError(err).Warn("failed in retrieving message")
		return broker.Message{}, err
	}
	defer rows.Close()

	var msgBdy []byte
	var expirationTime int
	var removed bool
	if rows.Next() {
		if err := rows.Scan(&msgBdy, &expirationTime, &removed); err != nil {
			pd.log.WithError(err).Warn("failed in scanning fetched data from database")
			return broker.Message{}, err
		}

		if err := rows.Err(); err != nil {
			return broker.Message{}, err
		}
	} else {
		return broker.Message{}, broker.ErrInvalidID
	}

	if removed {
		return broker.Message{}, broker.ErrExpiredID
	}

	return broker.Message{
		Body:       string(msgBdy),
		Expiration: time.Duration(expirationTime),
	}, nil
}

func (pd *PostgresDB) GetMessagesBySubject(ctx context.Context, subject string) ([]broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetMessages based on the given subject from postgresql")
	defer span.Finish()

	var messages = make([]broker.Message, 0)
	query := fmt.Sprintf("SELECT id, body FROM messages WHERE subject = '%s' AND removed = false;", subject)
	rows, err := pd.conn.Query(query)
	if err != nil {
		pd.log.WithError(err).Warn("failed in retrieving messages with the given subject")
		return nil, err
	}
	for rows.Next() {
		var id int
		var body []byte
		rows.Scan(&id, &body)
		messages = append(messages, broker.Message{
			Body: string(body),
		})
	}

	return messages, nil
}

func (pd *PostgresDB) DeleteMessage(subject string, id int) {
	span, _ := opentracing.StartSpanFromContext(context.Background(), "Delete message from postgresql")
	defer span.Finish()

	pd.Lock()
	pd.deletionList = append(pd.deletionList, strconv.Itoa(id))
	pd.Unlock()
}

func (pd *PostgresDB) scheduledBatchDeletion() {
	ticker := time.NewTicker(time.Duration(5 * time.Second))

	for range ticker.C {
		pd.Lock()
		if len(pd.deletionList) > 0 {
			query := fmt.Sprintf("UPDATE messages SET removed = true WHERE id in (%v)", strings.Join(pd.deletionList, ", "))
			_, err := pd.conn.Exec(query)
			if err != nil {
				pd.log.WithError(err).Warn("can not update 'removed' field for items in deletion list")
			}
			pd.deletionList = pd.deletionList[:0]
		}
		pd.Unlock()
	}
}
