from email import policy
from email.parser import BytesParser
from mailbox import Message
import re
import sys
from imapclient import IMAPClient
import json
import email


def main():
    config = read_config()
    group_settings = read_group_settings()
    emailSettings = config["emailSettings"]
    group_settings_to_process = get_group_settings_to_process(
        group_settings, emailSettings)

    try:
        emailsToProcess = []
        for settings in emailSettings:
            emailData = get_latest_email(settings, group_settings_to_process)
            emailsToProcess.append(emailData)

        # print(json.dumps(emailsToProcess, indent=4))
        print(json.dumps(group_settings_to_process, indent=4))
    except Exception as e:
        print(e)
        sys.exit(1)
    exit(0)


def get_group_settings_to_process(group_settings: list, emailSettings):
    email_settings_emails = list(
        map(lambda setting: setting["username"], emailSettings))
    group_settings_to_process = list(filter(
        lambda y: y["emailToRead"] in email_settings_emails, group_settings))
    return group_settings_to_process


def read_group_settings():
    raw_data = sys.stdin.read()
    json_data = json.loads(raw_data)

    return json_data


def get_latest_email(settings, group_settings_to_process):
    results = {}
    with IMAPClient(host=settings["host"]) as client:
        client.login(settings["username"], settings["password"])
        client.select_folder('INBOX')

        # messages = client.search(['UNSEEN'])
        messages = client.search(['ALL'])
        response = client.fetch(messages, ['FLAGS', 'RFC822'])

        for message_id, data in response.items():
            formatted_data = get_formatted_message_data(
                data, group_settings_to_process)
            if formatted_data:
                results[message_id] = formatted_data

    return results


def get_formatted_message_data(data, group_settings_to_process):
    message_data = email.message_from_bytes(data[b"RFC822"])
    fromData = message_data.get("From").split("<")
    fromName = fromData[0]
    fromEmail = fromData[1].replace("<", "").replace(">", "")
    toEmail = message_data.get("To")
    subject = message_data.get("Subject")

    should_process = False
    for group_setting in group_settings_to_process:
        if group_setting["emailToRead"] == toEmail:
            should_process = should_process_email(
                subject, fromEmail, group_setting)
            return False

    if not should_process:
        return None

    result = {
        "date": message_data.get("Date"),
        "subject": subject,
        "to": message_data.get("To"),
        "fromName": fromName,
        "fromEmail": fromEmail,
        "attachments": get_attachments(message_data),
    }

    return result


def should_process_email(subject, from_email, group_setting):
    whitelist_emails = group_setting["emailWhiteList"]
    subject_line_regexes = group_setting["subjectLineRegexes"]
    # check that form email is in the whitelist emails
    # if there are no emails, then bypass
    valid_email = valid_from_email(from_email, whitelist_emails)

    # iterate over regexes and check that subject matches at least one of the regexes
    # if there are no regexes, then bypass
    valid_subject_line = valid_subject(subject, subject_line_regexes)

    return valid_email and valid_subject_line


def valid_from_email(from_email, whitelist_emails):
    if len(whitelist_emails) == 0:
        return True

    whitelist_email_addresses = list(
        map(lambda emails: emails["email"], whitelist_emails))
    if from_email not in whitelist_email_addresses:
        return False


def valid_subject(subject, subject_line_regexes):
    if len(subject_line_regexes) == 0:
        return True

    for subject_line_regex in subject_line_regexes:
        matches = re.search(subject_line_regex.regex, subject)
        if matches:
            return True

    return False


def get_attachments(message_data: Message):
    result = []
    for part in message_data.walk():
        if part.get_content_maintype() == 'multipart':
            continue
        if part.get('Content-Disposition') is None:
            continue

        filename = part.get_filename()
        mime_type = part.get_content_type()

        if bool(fileName) and valid_mime_type(mime_type):
            filePath = f"./temp/{fileName}"
            with open(filePath, 'wb') as f:
                f.write(part.get_payload(decode=True))

            result.append({
                "filename": fileName,
            })

    return result


def valid_mime_type(mime_type):
    image_mime_types_regex = r"image\/.*"
    index = image_mime_types_regex.find(mime_type)
    return index > -1


def read_config():
    path = "config/config.dev.json"
    f = open(path, "r")
    data = json.load(f)
    f.close()

    return data


if __name__ == "__main__":
    main()
