package storage

import (
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"nostr-citadel/pkg/config"
)

var DB *sqlx.DB

func InitDB() error {
	db, err := sqlx.Connect("sqlite3", config.Config.Database.Name)

	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	_, _ = db.Exec(`
						create table event
						(
							id         text      not null
								constraint event_pk
									primary key,
							pubkey     text      not null,
							created_at timestamp not null,
							kind       integer   not null,
							tags       text      not null,
							content    text      not null,
							sig        text      not null
						);
						
						create index event_pubkey_time_idx ON event (pubkey, created_at);
						create index event_kind_idx ON event (kind);
						
						create table whitelist
						(
							pubkey      text not null
								constraint whitelist_pk
									primary key,
							created_at  timestamp default CURRENT_TIMESTAMP not null,
							sync        bool      default false             not null,
							last_synced timestamp,
							admin       integer   default 0                 not null
						);
						
						create index whitelist_admin_index
							on whitelist (admin);
						
						create index whitelist_pubkey_admin_index
							on whitelist (pubkey, admin);
    `)

	db.Mapper = reflectx.NewMapperFunc("json", sqlx.NameMapper)
	_, _ = db.Exec("PRAGMA journal_mode=WAL")
	DB = db
	return nil
}
