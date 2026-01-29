package mongodb

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func URIBeautifier(uri string) string {
	// parse uri
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return "***"
	}
	fmt.Printf("[DEBUG] URL structure: %#v\n", u)
	// mask user info
	if u.User != nil {
		if _, set := u.User.Password(); !set {
			u.User = url.UserPassword(u.User.Username(), enterPassword())
		}
	} else {
		// skip enter password for localhost
		if strings.HasPrefix(u.Host, "localhost") || strings.HasPrefix(u.Host, "127.0.0.1") {
			fmt.Println("Username/Password is missing in the URI, but host is localhost, skipping user input.")
		} else {
			fmt.Println("Username/Password is missing in the URI. Required user input.")
			// no user info, ask for it
			user := enterUserName()
			password := enterPassword()
			// if no password, set only user
			if user != "" {
				u.User = url.User(user)
			}
			// set password if provided
			if password != "" {
				u.User = url.UserPassword(user, password)
			}
		}
	}
	// mask database name
	if u.Path == "" {
		u.Path = enterDatabaseName()
	}

	return u.String()
}

func enterUserName() (username string) {
	fmt.Print("Enter username: ")
	if _, err := fmt.Fscan(os.Stdin, &username); err != nil {
		return ""
	}
	return username
}

func enterPassword() (password string) {
	fmt.Print("Enter password: ")
	if _, err := fmt.Fscan(os.Stdin, &password); err != nil {
		return ""
	}
	return password
}

func enterDatabaseName() (dbname string) {
	fmt.Print("Enter database name: ")
	if _, err := fmt.Fscan(os.Stdin, &dbname); err != nil {
		return ""
	}
	return dbname
}

func newOrRenewConnection(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
	var (
		primaryStatus, secondaryStatus error
	)
	client, err := mongo.Connect(opts...)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		primaryStatus = err
		log.Printf("Primary ping failed: %v", err)
	}
	if err := client.Ping(ctx, readpref.Secondary()); err != nil {
		secondaryStatus = err
		log.Printf("Secondary ping failed: %v", err)
	}

	if primaryStatus != nil && secondaryStatus != nil {
		return nil, fmt.Errorf("unable to connect to primary or secondary, primary: %v, secondary: %v", primaryStatus, secondaryStatus)
	}
	return client, nil
}

func retryMechanism(current *mongo.Client, dbName string, operation func(*mongo.Database) error) (*mongo.Client, error) {

	// Helper function to identify retryable errors
	isRetryableError := func(err error) bool {
		if err == nil {
			return false
		}
		errStr := err.Error()
		return strings.Contains(errStr, "DeadlineExceeded") ||
			/* This is a normal case - ignore it: strings.Contains(errStr, "context deadline exceeded") || */
			strings.Contains(errStr, "connection") ||
			strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "network") ||
			strings.Contains(errStr, "closed") || strings.Contains(errStr, "disconnected") // may be transient
	}

	// try with case: normal write operation
	if err := operation(current.Database(dbName)); !isRetryableError(err) {
		return nil, err
	} else {
		log.Printf("Retrying MongoDB write operation due to error: %v --> Reconnecting...", err)
	}
	// implement retry logic here if needed
	renew, err := newOrRenewConnection(context.TODO())
	if err != nil {
		return nil, err
	}
	if err := operation(renew.Database(dbName)); !isRetryableError(err) {
		log.Printf("MongoDB write operation succeeded after reconnection.")
		return renew, err
	} else {
		log.Printf("MongoDB write operation failed again after reconnection: %v --> keep current connection", err)
		return nil, err
	}
}
