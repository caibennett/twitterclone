if [ ! -f app.db ]; then
    echo "Applying schema..."
    cat sql/schema.sql | sqlite3 "app.db"
    echo "Applied schema"
else
    echo "No need to apply schema, file exists"
fi