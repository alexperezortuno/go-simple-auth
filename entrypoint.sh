#!/bin/sh
sqlite3 --version
# Nombre del archivo de la base de datos
cd /usr/local/bin/
DB_FILE="users.db"

# Comando SQL para crear una tabla si no existe
SQL_CREATE_TABLE="CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password TEXT NOT NULL
);"

# Crear la base de datos y la tabla
sqlite3 $DB_FILE "$SQL_CREATE_TABLE"

chmod +x /usr/local/bin/users.db

echo "Database and table created successfully."
chmod +x /usr/local/bin/go-simple-auth

/usr/local/bin/go-simple-auth
