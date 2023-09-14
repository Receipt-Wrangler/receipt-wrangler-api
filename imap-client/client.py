from email import policy
from email.parser import BytesParser
from mailbox import Message
from imapclient import IMAPClient
import json
import email


def main():
    config = read_config()
    emailSettings = config["emailSettings"]
    emailsToProcess = []
    for settings in emailSettings:
        emailData = get_latest_email(settings)
        emailsToProcess.append(emailData)


def get_latest_email(settings):
    results = {}
    with IMAPClient(host=settings["host"]) as client:
        client.login(settings["username"], settings["password"])
        client.select_folder('INBOX')

        # messages = client.search(['UNSEEN'])
        messages = client.search(['ALL'])
        response = client.fetch(messages, ['FLAGS', 'RFC822'])

        for message_id, data in response.items():
            formatted_data = get_formatted_message_data(data)
            results[message_id] = formatted_data

    return results


def get_formatted_message_data(data):
    message_data = email.message_from_bytes(data[b"RFC822"])
    fromData = message_data.get("From").split("<")
    fromName = fromData[0]
    fromEmail = fromData[1].replace("<", "").replace(">", "")

    result = {
        "date": message_data.get("Date"),
        "subject": message_data.get("Subject"),
        "to": message_data.get("To"),
        "fromName": fromName,
        "fromEmail": fromEmail,
        "attachments": get_attachments(message_data),
    }

    return result


def get_attachments(message_data: Message):
    result = []
    for part in message_data.walk():
        if part.get_content_maintype() == 'multipart':
            continue
        if part.get('Content-Disposition') is None:
            continue

        fileName = part.get_filename()
        if bool(fileName):
            filePath = f"./temp/{fileName}"
            with open(filePath, 'wb') as f:
                f.write(part.get_payload(decode=True))

            result.append({
                "fileName": fileName,
            })

    return result


def read_config():
    path = "config/config.dev.json"
    f = open(path, "r")
    data = json.load(f)
    f.close()

    return data


if __name__ == "__main__":
    main()
