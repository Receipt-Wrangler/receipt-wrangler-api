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

    print(emailsToProcess)


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

    for k, v in message_data.items():
        print(k, v)

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
    if message_data.is_multipart():
        # Iterate through each part
        for part in message_data:
            # If the part is an attachment
            print(part, "part")
            if part() == 'attachment':
                # Extract filename
                filename = part.get_filename()
                if filename:
                    # Open the file in write-binary mode and save it
                    result.append({
                        "filename": filename,
                        "payload": part.get_payload(decode=True)
                    })
                    # with open(filename, 'wb') as f:
                    #     f.write(part.get_payload(decode=True))
                    print(f"Saved attachment as {filename}")

    return result


def read_config():
    path = "config/config.dev.json"
    f = open(path, "r")
    data = json.load(f)
    f.close()

    return data


if __name__ == "__main__":
    main()
