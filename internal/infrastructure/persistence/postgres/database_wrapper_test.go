package postgres

import (
	"testing"
)

func TestExtractSQLOperationAndTable(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedOp    string
		expectedTable string
	}{
		// SELECT statements
		{
			name:          "simple SELECT",
			query:         "SELECT * FROM users",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "SELECT with column names",
			query:         "SELECT id, name, email FROM customers WHERE id = 1",
			expectedOp:    "SELECT",
			expectedTable: "customers",
		},
		{
			name:          "SELECT with schema prefix",
			query:         "SELECT * FROM public.users",
			expectedOp:    "SELECT",
			expectedTable: "public.users",
		},
		{
			name:          "SELECT with quoted table name",
			query:         `SELECT * FROM "user_profiles"`,
			expectedOp:    "SELECT",
			expectedTable: "user_profiles",
		},
		{
			name:          "SELECT with backtick quoted table",
			query:         "SELECT * FROM `users`",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "SELECT with single quote table",
			query:         "SELECT * FROM 'users'",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "SELECT with JOIN",
			query:         "SELECT u.id FROM users u JOIN orders o ON u.id = o.user_id",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "SELECT with subquery",
			query:         "SELECT * FROM (SELECT id FROM users)",
			expectedOp:    "SELECT",
			expectedTable: "",
		},
		{
			name:          "SELECT with lowercase",
			query:         "select * from users",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "SELECT with mixed case",
			query:         "SeLeCt * FrOm users",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "SELECT with leading whitespace",
			query:         "   SELECT * FROM users",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "SELECT with newlines",
			query:         "SELECT *\nFROM\nusers",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},

		// INSERT statements
		{
			name:          "simple INSERT",
			query:         "INSERT INTO users (id, name) VALUES (1, 'John')",
			expectedOp:    "INSERT",
			expectedTable: "users",
		},
		{
			name:          "INSERT with quoted table",
			query:         `INSERT INTO "user_profiles" (id, email) VALUES (1, 'test@example.com')`,
			expectedOp:    "INSERT",
			expectedTable: "user_profiles",
		},
		{
			name:          "INSERT with schema",
			query:         "INSERT INTO public.users (name) VALUES ('Alice')",
			expectedOp:    "INSERT",
			expectedTable: "public.users",
		},
		{
			name:          "INSERT with lowercase",
			query:         "insert into users values (1, 'Bob')",
			expectedOp:    "INSERT",
			expectedTable: "users",
		},

		// UPDATE statements
		{
			name:          "simple UPDATE",
			query:         "UPDATE users SET name = 'Charlie' WHERE id = 1",
			expectedOp:    "UPDATE",
			expectedTable: "users",
		},
		{
			name:          "UPDATE with quoted table",
			query:         `UPDATE "user_profiles" SET status = 'active'`,
			expectedOp:    "UPDATE",
			expectedTable: "user_profiles",
		},
		{
			name:          "UPDATE with schema",
			query:         "UPDATE public.users SET email = 'new@example.com'",
			expectedOp:    "UPDATE",
			expectedTable: "public.users",
		},
		{
			name:          "UPDATE with lowercase",
			query:         "update users set name = 'Diana'",
			expectedOp:    "UPDATE",
			expectedTable: "users",
		},

		// DELETE statements
		{
			name:          "simple DELETE",
			query:         "DELETE FROM users WHERE id = 1",
			expectedOp:    "DELETE",
			expectedTable: "users",
		},
		{
			name:          "DELETE with quoted table",
			query:         `DELETE FROM "user_profiles" WHERE status = 'inactive'`,
			expectedOp:    "DELETE",
			expectedTable: "user_profiles",
		},
		{
			name:          "DELETE with schema",
			query:         "DELETE FROM public.users WHERE created_at < NOW()",
			expectedOp:    "DELETE",
			expectedTable: "public.users",
		},
		{
			name:          "DELETE with lowercase",
			query:         "delete from users where age > 100",
			expectedOp:    "DELETE",
			expectedTable: "users",
		},

		// Edge cases
		{
			name:          "empty string",
			query:         "",
			expectedOp:    "",
			expectedTable: "",
		},
		{
			name:          "whitespace only",
			query:         "   \n\t  ",
			expectedOp:    "",
			expectedTable: "",
		},
		{
			name:          "no FROM clause in DELETE",
			query:         "DELETE users",
			expectedOp:    "DELETE",
			expectedTable: "",
		},
		{
			name:          "table name with underscore",
			query:         "SELECT * FROM user_profiles",
			expectedOp:    "SELECT",
			expectedTable: "user_profiles",
		},
		{
			name:          "table name with numbers",
			query:         "SELECT * FROM users2024",
			expectedOp:    "SELECT",
			expectedTable: "users2024",
		},
		{
			name:          "table name with schema and table",
			query:         "SELECT * FROM schema.table",
			expectedOp:    "SELECT",
			expectedTable: "schema.table",
		},
		{
			name:          "query with semicolon at end",
			query:         "SELECT * FROM users;",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "query with comment",
			query:         "SELECT * FROM users -- get all users",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "multiple tables (INSERT SELECT)",
			query:         "INSERT INTO users SELECT * FROM backup_users",
			expectedOp:    "INSERT",
			expectedTable: "users",
		},
		{
			name:          "SELECT with WHERE and multiple tables",
			query:         "SELECT u.* FROM users u, orders o WHERE u.id = o.user_id",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
		{
			name:          "unquoted identifier with parenthesis",
			query:         "SELECT * FROM (SELECT id FROM users)",
			expectedOp:    "SELECT",
			expectedTable: "",
		},
		{
			name:          "UPDATE with FROM clause",
			query:         "UPDATE users SET status = 'active' FROM orders WHERE users.id = orders.user_id",
			expectedOp:    "UPDATE",
			expectedTable: "users",
		},
		{
			name:          "invalid SQL",
			query:         "INVALID SYNTAX HERE",
			expectedOp:    "",
			expectedTable: "",
		},
		{
			name:          "only operation without table",
			query:         "SELECT",
			expectedOp:    "SELECT",
			expectedTable: "",
		},
		{
			name:          "SELECT with complex WHERE",
			query:         "SELECT id FROM users WHERE email LIKE '%@example.com' AND age > 18",
			expectedOp:    "SELECT",
			expectedTable: "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, table := extractSQLOperationAndTable(tt.query)

			if op != tt.expectedOp {
				t.Errorf("operation: expected %q, got %q", tt.expectedOp, op)
			}

			if table != tt.expectedTable {
				t.Errorf("table: expected %q, got %q", tt.expectedTable, table)
			}
		})
	}
}

// Benchmark tests
func BenchmarkExtractSQLOperationAndTable(b *testing.B) {
	queries := []string{
		"SELECT * FROM users",
		"INSERT INTO users (id, name) VALUES (1, 'John')",
		"UPDATE users SET name = 'Jane' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
		"SELECT u.id FROM users u JOIN orders o ON u.id = o.user_id",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			extractSQLOperationAndTable(q)
		}
	}
}
