package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/smtp"
	"os"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

type Config struct {
	Port string

	SMTPServer      string
	SMTPSenderEmail string
	SMTPPassword    string

	DestEmail string
}

func config() (Config, error) {
	cfg := Config{}

	fs := ff.NewFlagSet("smtpsender")

	fs.StringVar(&cfg.Port, 'p', "port", "12345", "the TCP port the server listens on")
	fs.StringVar(&cfg.SMTPSenderEmail, 'e', "smtp-sender-email", "", "")
	fs.StringVar(&cfg.SMTPServer, 's', "smtp-server", "", "")
	fs.StringVar(&cfg.SMTPPassword, 0, "smtp-password", "", "")
	fs.StringVar(&cfg.DestEmail, 'd', "dest-email", "", "")

	err := ff.Parse(fs,
		os.Args[1:],
		ff.WithEnvVarPrefix("SMTPSENDER"),
	)
	switch {
	case errors.Is(err, ff.ErrHelp):
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Flags(fs))
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return cfg, nil
}

func main() {
	cfg, err := config()
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		panic(err)
	}

	for {
		log.Println("waiting for tcp conn")
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			log.Println("handling connection")
			defer conn.Close()

			data, err := io.ReadAll(conn)
			if err != nil {
				log.Println("read all:", err)
			}
			log.Println("read:", string(data))

			auth := smtp.PlainAuth("", cfg.SMTPSenderEmail, cfg.SMTPPassword, cfg.SMTPServer)

			headers := make(map[string]string)
			headers["From"] = cfg.SMTPSenderEmail
			headers["To"] = cfg.DestEmail
			headers["Subject"] = "Notification"
			headers["Date"] = time.Now().Format(time.RFC3339)

			var message string
			for k, v := range headers {
				message += fmt.Sprintf("%s: %s\r\n", k, v)
			}
			message += "\r\n" + string(data)

			tlsCfg := &tls.Config{
				ServerName: cfg.SMTPServer,
			}

			cconn, err := tls.Dial("tcp", cfg.SMTPServer+":465", tlsCfg)
			if err != nil {
				log.Println("tls dial:", err)
			}

			c, err := smtp.NewClient(cconn, cfg.SMTPServer)
			if err != nil {
				log.Println("smtp new client:", err)
			}

			err = c.Auth(auth)
			if err != nil {
				log.Println("smtp client auth:", err)
			}

			err = c.Mail(cfg.SMTPSenderEmail)
			if err != nil {
				log.Println("smtp client mail:", err)
			}

			err = c.Rcpt(cfg.DestEmail)
			if err != nil {
				log.Println("smtp client rcpt:", err)
			}

			w, err := c.Data()
			if err != nil {
				log.Println("smtp client data:", err)
			}

			_, err = w.Write([]byte(message))
			if err != nil {
				log.Println("smtp client write:", err)
			}

			err = w.Close()
			if err != nil {
				log.Println("smtp client writer close:", err)
			}

			err = c.Quit()
			if err != nil {
				log.Println("smtp client quit:", err)
			}
		}()
	}
}
