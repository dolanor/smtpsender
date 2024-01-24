# SMTP Sender

The spamming solution for cheap IoT bastards.


## Docker

The image is available as `dolanor/smtpsender:latest`
You need to configure the environment as:

```env
SMTPSENDER_SMTP_SENDER_EMAIL=sender@example.com
SMTPSENDER_SMTP_PASSWORD=password
SMTPSENDER_SMTP_SERVER=mail.example.com
SMTPSENDER_DEST_EMAIL=destination@example.com
SMTPSENDER_PORT=12345
```
