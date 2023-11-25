package pgsql

import (
	"github.com/forscht/ddrv/pkg/migrate"
)

var migrations = []migrate.Migration{
	{
		ID: 1,
		Up: migrate.Queries([]string{
			fsTable,
			nodeTable,
			fsParentIdx,
			fsNameIdx,
			rootInsert,
			statFunction,
			lsFunction,
			treeFunction,
			touchFunction,
			mkdirFunction,
			mvFunction,
			rmFunction,
			resetFunction,
			parserootFunction,
			validnameFunction,
			sanitizeFPath,
			parseSizeFunction,
			basenameFunction,
			dirnameFunction,
		}),
		Down: migrate.Queries([]string{dropFs}),
	},
	{
		ID: 2,
		Up: migrate.Queries([]string{
			`
				CREATE TABLE temp_node
				(
				    id    BIGINT PRIMARY KEY NOT NULL,
				    file  UUID               NOT NULL REFERENCES fs (id) ON DELETE CASCADE,
				    url   VARCHAR(255)       NOT NULL,
				    size  INTEGER            NOT NULL,
				    iv    VARCHAR(255)       NOT NULL DEFAULT '',
				    mtime TIMESTAMP          NOT NULL DEFAULT NOW()
				);
				
				INSERT INTO temp_node (id, file, url, size, iv, mtime)
				SELECT CAST(
				               (REGEXP_MATCHES(url, '/([0-9]+)/[A-Za-z0-9_-]+$', 'g'))[1]
				           AS BIGINT) AS id,
				       file,
				       url,
				       size,
				       iv,
				       mtime
				FROM node;
				
				DROP TABLE node;
				
				ALTER TABLE temp_node RENAME TO node;
				
				alter table public.node rename constraint temp_node_pkey to node_pkey;
				
				alter table public.node rename constraint temp_node_file_fkey to node_file_fkey;
			`,
		}),
		Down: migrate.Queries([]string{}),
	},
	{
		ID:   3,
		Up:   migrate.Queries([]string{`CREATE INDEX idx_node_file ON node (file);`}),
		Down: migrate.Queries([]string{`DROP INDEX idx_node_file;`}),
	},
	{
		ID:   4,
		Up:   migrate.Queries([]string{`CREATE INDEX idx_node_size ON node (size);`}),
		Down: migrate.Queries([]string{`DROP INDEX idx_node_size;`}),
	},
	{
		ID:   5,
		Up:   migrate.Queries([]string{`ALTER TABLE your_table_name ADD COLUMN mid VARCHAR(255), ADD COLUMN ex INT, ADD COLUMN is INT, ADD COLUMN hm VARCHAR(255);`}),
		Down: migrate.Queries([]string{`ALTER TABLE your_table_name DROP COLUMN mid, DROP COLUMN ex, DROP COLUMN is, DROP COLUMN hm;`}),
	},
}
