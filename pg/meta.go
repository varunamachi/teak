package pg

import "github.com/varunamachi/teak"

func getDatabases() (dbs []string, err error) {
	dbs = make([]string, 0, 100)
	err = defDB.Select(
		dbs,
		`SELECT datname FROM pg_database WHERE datistemplate = false;`)
	return dbs, teak.LogErrorX("t.pg.meta", "Failed to get database list", err)
}

func getTables(db string) (tables []string, err error) {
	conn := NamedConn(db)
	if conn == nil {
		err = teak.Error("t.pg.meta", "No database with name %s found", db)
		return tables, err
	}
	tables = make([]string, 0, 100)
	err = defDB.Select(tables,
		`SELECT table_name FROM information_schema.tables 
			WHERE table_schema = 'public'`)
	return tables, teak.LogErrorX("t.pg.meta", "Failed to get tables list", err)
}

func getViews(db string) (views []string, err error) {
	conn := NamedConn(db)
	if conn == nil {
		err = teak.Error("t.pg.meta", "No database with name %s found", db)
		return views, err
	}
	views = make([]string, 0, 100)
	err = defDB.Select(views,
		`SELECT table_name FROM information_schema.views 
			WHERE table_schema = 'public`)
	return views, teak.LogErrorX("t.pg.meta", "Failed to get views list", err)
}
