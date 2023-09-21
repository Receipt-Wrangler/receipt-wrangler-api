import datetime
from itertools import chain
from mailbox import Message
import os
from imapclient import IMAPClient
from imapclient import exceptions
import logging
import re
import sys
import json
import email

base_path = os.environ.get("BASE_PATH", "")


def main():
    init_logger()
    config = read_config()
    group_settings = read_group_settings()
    emailSettings = config["emailSettings"]
    group_settings_to_process = get_group_settings_to_process(
        group_settings, emailSettings)

    try:
        emailsToProcess = []
        for settings in emailSettings:
            emailData = get_unread_emails_to_process(
                settings, group_settings_to_process)
            emailsToProcess.append(emailData)

        results = list(chain.from_iterable(emailsToProcess))
        json_results = json.dumps(results)

        logging.info(f"Results: {json_results}")

        print(json_results)
    except exceptions.LoginError as e:
        logging.error(e)
    except Exception as e:
        logging.error(e)

    exit(0)


def init_logger():
    path = os.path.join(base_path, "logs", "imap-client.log")
    logging.basicConfig(filename=path, level=logging.INFO,
                        format='%(asctime)s %(levelname)s {%(pathname)s:%(lineno)d} %(message)s')


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


def get_unread_emails_to_process(settings, group_settings_to_process):
    results = []
    with IMAPClient(host=settings["host"]) as client:
        client.login(settings["username"], settings["password"])
        client.select_folder('INBOX')

        messages = client.search(['UNSEEN'])
        response = client.fetch(messages, ['FLAGS', 'RFC822'])

        for message_id, data in response.items():
            formatted_data = get_formatted_message_data(
                data, group_settings_to_process)
            if len(formatted_data) > 0:
                formatted_data[message_id] = message_id
                results.append(formatted_data)

    return results


def get_formatted_message_data(data, group_settings_to_process):
    message_data = email.message_from_bytes(data[b"RFC822"])
    fromData = message_data.get("From").split("<")
    fromName = fromData[0]
    fromEmail = fromData[1].replace("<", "").replace(">", "")
    toEmail = message_data.get("To")
    subject = message_data.get("Subject")

    should_process = False
    group_settings_ids = []
    for group_setting in group_settings_to_process:
        if group_setting["emailToRead"] == toEmail:
            should_process = should_process_email(
                subject, fromEmail, group_setting)
            if should_process:
                group_settings_ids.append(group_setting["id"])

    if not should_process:
        return {}

    date = datetime.datetime.strptime(message_data.get(
        "Date"), "%a, %d %b %Y %H:%M:%S %z")
    utc_date = date.replace(tzinfo=datetime.timezone.utc)
    formatted_date = utc_date.strftime("%Y-%m-%dT%H:%M:%S.%fZ")

    result = {
        "date": formatted_date,
        "subject": subject,
        "to": message_data.get("To"),
        "fromName": fromName,
        "fromEmail": fromEmail,
        "attachments": get_attachments(message_data),
        "groupSettingsIds": group_settings_ids,
    }

    if (len(result["attachments"]) == 0):
        return {}

    logging.info(f"Formatted message data: {result}")
    return result


def should_process_email(subject, from_email, group_setting):
    whitelist_emails = group_setting["emailWhiteList"]
    subject_line_regexes = group_setting["subjectLineRegexes"]

    valid_email = valid_from_email(from_email, whitelist_emails)
    valid_subject_line = valid_subject(subject, subject_line_regexes)

    should_process = valid_email and valid_subject_line

    logging.info(
        f"Should process email: '{subject}' from: '{from_email}' {should_process}. Valid email: {valid_email}. Valid subject line: {valid_subject_line} ")
    return should_process


def valid_from_email(from_email, whitelist_emails):
    if len(whitelist_emails) == 0:
        return True

    whitelist_email_addresses = list(
        map(lambda emails: emails["email"], whitelist_emails))
    if from_email not in whitelist_email_addresses:
        return False

    return True


def valid_subject(subject, subject_line_regexes):
    if len(subject_line_regexes) == 0:
        return True

    for subject_line_regex in subject_line_regexes:
        regex = re.compile(subject_line_regex["regex"])
        matches = regex.search(subject)
        logging.info(
            f"Found match: {matches} on email subject: '{subject}' with regex: '{subject_line_regex['regex']}'")
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

        logging.info(f"Filename: {filename} mime_type: {mime_type}")

        if len(filename) > 0 and valid_mime_type(mime_type):
            filePath = f"./temp/{filename}"
            with open(filePath, 'wb') as f:
                f.write(part.get_payload(decode=True))

            size = os.path.getsize(filePath)

            result.append({
                "filename": filename,
                "fileType": mime_type,
                "size": size,
            })

    return result


def valid_mime_type(mime_type):
    image_mime_types_regex = r"image\/.*"
    match = re.search(image_mime_types_regex, mime_type or "")
    return match is not None


def read_config():
    env = os.environ.get("ENV", "dev")
    path = os.path.join(base_path, "config", f"config.{env}.json")
    f = open(path, "r")
    data = json.load(f)
    f.close()

    return data


if __name__ == "__main__":
    main()
