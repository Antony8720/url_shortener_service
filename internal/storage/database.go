package storage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/Antony8720/url-shortener/internal/app/violationerror"
)

type DatabaseStorage struct{
	db *pgxpool.Pool
}

type DatabaseURL struct {
	userID uuid.UUID `db:"user_id"`
	short  string    `db:"short_url"`
	long   string    `db:"long_url"`
}

func NewDatabaseStorage(DBAddress string) (*DatabaseStorage, error){
	pgxConfig, err := pgxpool.ParseConfig(DBAddress)
	if err != nil {
		return &DatabaseStorage{}, err
	}
	pgxConnPool, err := pgxpool.ConnectConfig(context.Background(), pgxConfig)
	if err != nil {
		return &DatabaseStorage{}, err
	}
	query := `CREATE TABLE IF NOT EXISTS database_url
			 (
			 id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
			 user_id uuid NOT NULL,
			 short_url text NOT NULL,
			 long_url text NOT NULL,
			 PRIMARY KEY (id));
			 CREATE UNIQUE INDEX IF NOT EXISTS long_url_unique_idx on database_url(long_url);`
	_, err = pgxConnPool.Exec(context.Background(), query) 
	if err != nil {
		return &DatabaseStorage{}, err
	}
	return &DatabaseStorage{db: pgxConnPool}, nil
}

func (dbs *DatabaseStorage) Get(short string) (string, bool) {
	var url DatabaseURL
	err := dbs.db.QueryRow(context.Background(),
						  `SELECT user_id, short_url, long_url
						   FROM database_url 
						   WHERE short_url = $1::text`, short).Scan(&url.userID, &url.short, &url.long)
	if err != nil {
		return "", false
	}
	return url.long, true		
}

func (dbs *DatabaseStorage) GetHistory(userID uuid.UUID) (map[string]string, error) {
	res := make(map[string]string)
	rows, err := dbs.db.Query(context.Background(), 
							`SELECT user_id, short_url, long_url 
							 FROM database_url 
							 WHERE user_id = $1::uuid`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var url DatabaseURL
		err = rows.Scan(&url.userID, &url.short, &url.long)
		if err != nil {
			return nil, err
		}
		res[url.short] = url.long
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func(dbs *DatabaseStorage) Set(userID uuid.UUID, short, long string) error {
	query := "INSERT INTO database_url(user_id, short_url, long_url) VALUES ($1::uuid, $2::text, $3::text)"
	_, err := dbs.db.Exec(context.Background(), query, userID, short, long)
	if err != nil {
		var pgError *pgconn.PgError
		if !errors.As(err, &pgError){
			return err
		}
		pgErr, ok := err.(*pgconn.PgError)
		if !ok {
			return err
		}
		if pgErr.Code != pgerrcode.UniqueViolation{
			return err
		}
		var ndb DatabaseURL
		if err := dbs.db.QueryRow(context.Background(),
								  "SELECT user_id, short_url, long_url FROM database_url WHERE long_url = $1::text", long,
								 ).Scan(&ndb.userID, &ndb.short, &ndb.long); err != nil{
									return err
								 }	
		
		return &violationerror.UniqueViolationError{
			Err: err,
			UserID: ndb.userID,
			Short: ndb.short,
			Long: ndb.long,
		}						 
	}
	return nil
}