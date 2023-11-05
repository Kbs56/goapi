package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func (conn *PostgresDB) getAllUsers() ([]*User, error) {
	sqlStatement := `
  select * from customer
  `
	rows, err := conn.db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (conn *PostgresDB) createUser(u *User) (*User, error) {
	sqlStatement := fmt.Sprintf(
		"INSERT INTO CUSTOMER (FIRST_NAME, LAST_NAME, EMAIL_ADDRESS) VALUES ('%s', '%s', '%s') RETURNING ID",
		u.FirstName,
		u.LastName,
		u.Email,
	)
	var id int
	if err := conn.db.QueryRow(sqlStatement).Scan(&id); err != nil {
		return nil, err
	}
	u.Id = id
	return u, nil
}

func (conn *PostgresDB) getUser(id int) (*User, error) {
	u := User{}
	err := conn.db.QueryRow("select * from customer where id = $1", id).
		Scan(&u.Id, &u.FirstName, &u.LastName, &u.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		} else {
			panic(err)
		}
	}
	return &u, nil
}

func (conn *PostgresDB) updateEmail(u *User) (*User, error) {
	sqlStatement := fmt.Sprintf(
		"UPDATE CUSTOMER SET EMAIL_ADDRESS = '%s' WHERE ID = %d RETURNING FIRST_NAME, LAST_NAME",
		u.Email,
		u.Id,
	)
	if err := conn.db.QueryRow(sqlStatement).Scan(&u.FirstName, &u.LastName); err != nil {
		return nil, err
	}
	return u, nil
}

func (conn *PostgresDB) deleteUser(id int) error {
	sqlStatement := fmt.Sprintf("DELETE FROM CUSTOMER WHERE ID = %d", id)
	res, err := conn.db.Exec(sqlStatement)
	if err != nil {
		return err
	}
	numRows, err := res.RowsAffected()
	if numRows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func ConnectDB() (*PostgresDB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
		os.Getenv("pghost"),
		os.Getenv("pgport"),
		os.Getenv("pguser"),
		os.Getenv("pgpass"),
		os.Getenv("pgdbname"),
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{
		db: db,
	}, nil
}
