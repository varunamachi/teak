package pg

import "github.com/varunamachi/teak"

func getDatabases() (dbs []string, err error) {
	dbs = make([]string, 0, 100)
	err = defDB.Select(
		dbs,
		`SELECT datname FROM pg_database WHERE datistemplate = false;`)
	return dbs, teak.LogErrorX("t.pg.meta", "Failed to get database list", err)
} 

func getTables() (tables []string, err error) {
	tables = make([]string, 0, 100)
	err = defDB.Select(tables,
		`SELECT table_name FROM information_schema.tables 
			WHERE table_schema = 'public'`)
	return tables, teak.LogErrorX("t.pg.meta", "Failed to get tables list", err)
}

func getViews() (views []string, err error) {
	views = make([]string, 0, 100)
	err = defDB.Select(views,
		`SELECT table_name FROM information_schema.views 
			WHERE table_schema = 'public`)
	return views, teak.LogErrorX("t.pg.meta", "Failed to get views list", err)
}
