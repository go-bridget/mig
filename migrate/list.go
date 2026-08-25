package migrate

import "context"

// List returns one Migration per file in the migrations filesystem, in lexical
// order of filename, along with the rows of the migrations table whose file is
// gone. It applies nothing.
//
// A migration the table has a row for carries the StatementIndex and Status of
// that row; one it has no row for carries a StatementIndex of -1 and an empty
// Status, which is what says it has never run.
//
// List creates the migrations table if it is missing.
func (m *Manager) List(ctx context.Context) ([]Migration, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	return m.report(ctx)
}
