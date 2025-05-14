import datetime
import email
import logging
import os
import re
from mailbox import Message

from imapclient import IMAPClient

from utils import valid_from_email, valid_subject

base_path = os.environ.get("BASE_PATH", "")


class ImapClient:
    host = None
    port = None
    username = None
    password = None
    subject_line_regexes = None
    email_whitelist = None
    client = None

    def __init__(self, host, port, username, password, use_starttls, subject_line_regexes, email_whitelist):
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.use_starttls = use_starttls
        self.subject_line_regexes = subject_line_regexes or []
        self.email_whitelist = email_whitelist or []

    def connect(self):
        if self.use_starttls:
            self.client = IMAPClient(self.host, self.port, ssl=False)
            self.client.starttls()
        else:
            self.client = IMAPClient(self.host, self.port)

        self.client.login(self.username, self.password)

    def get_unread_email_metadata(self):
        unread_emails = self._get_unread_emails()
        return self._messages_to_email_metadata(unread_emails)

    def _get_unread_emails(self):
        if self.client is None:
            self.connect()

        self.client.select_folder('INBOX')

        messages = self.client.search(['UNSEEN'])
        response = self.client.fetch(messages, ['FLAGS', 'RFC822'])
        return response or None

    def _messages_to_email_metadata(self, response):
        if response is None:
            return []

        results = []
        for message_id, data in response.items():
            formatted_data = self._get_formatted_message_data(
                data)
            if len(formatted_data) > 0:
                formatted_data[message_id] = message_id
                results.append(formatted_data)

        return results

    def _get_formatted_message_data(self, data):
        message_data = email.message_from_bytes(data[b"RFC822"])

        from_data = self._get_formatted_to_or_from_data(message_data, "From")
        if from_data["email"] is None:
            return {}

        to_data = self._get_formatted_to_or_from_data(message_data, "To")
        if to_data["email"] is None:
            return {}

        subject = message_data.get("Subject")
        should_process = self._valid_from_email(from_data["email"]) and self._valid_subject(subject)

        # TODO: V5 - Could we set this message to unread if not process?
        if not should_process:
            return {}

        formatted_date = self.get_formatted_date(message_data.get("Date"))

        result = {
            "date": formatted_date,
            "subject": subject,
            "to": to_data["email"],
            "fromName": from_data["name"],
            "fromEmail": from_data["email"],
            "attachments": self._get_attachments(message_data),
            "groupSettingsIds": [],
        }

        if len(result["attachments"]) == 0:
            return {}

        logging.info(f"Formatted message data: {result}")
        return result

    def _get_formatted_to_or_from_data(self, message_data: Message, key: str):
        result = {
            "name": None,
            "email": None
        }

        from_data = message_data.get(key).split("<")
        if len(from_data) == 2:
            result["name"] = from_data[0]
            result["email"] = from_data[1].replace("<", "").replace(">", "")

        if len(from_data) == 1:
            result["email"] = from_data[0]

        logging.info(f"Formatted from data: {result}")

        return result

    def get_formatted_date(self, date):
        date_parts = date.split(
            "(")  # Fixes case when date comes back in utc , so (UTC) is appended
        logging.info(date[0])
        date = datetime.datetime.strptime(
            date_parts[0].strip(), "%a, %d %b %Y %H:%M:%S %z")
        utc_date = date.replace(tzinfo=datetime.timezone.utc)
        formatted_date = utc_date.strftime("%Y-%m-%dT%H:%M:%S.%fZ")

        return formatted_date

    def _valid_from_email(self, from_email):
        return valid_from_email(from_email, self.email_whitelist)

    def _valid_subject(self, subject):
        return valid_subject(subject, self.subject_line_regexes)

    def _get_attachments(self, message_data: Message):
        result = []
        for part in message_data.walk():
            if part.get_content_maintype() == 'multipart':
                continue
            if part.get('Content-Disposition') is None:
                continue

            filename = part.get_filename()
            mime_type = part.get_content_type()

            logging.info(f"Filename: {filename} mime_type: {mime_type}")

            if len(filename) > 0 and self.valid_mime_type(mime_type):
                filePath = os.path.join(base_path, "temp", filename)
                with open(filePath, 'wb') as f:
                    f.write(part.get_payload(decode=True))

                size = os.path.getsize(filePath)
                data = {
                    "filename": filename,
                    "fileType": mime_type,
                    "size": size,
                }

                result.append(data)

        return result

    def valid_mime_type(self, mime_type):
        image_mime_types_regex = r"^(image\/(jpeg|png|heic|bmp|webp|tiff)|application\/pdf)$"
        match = re.search(image_mime_types_regex, mime_type or "")
        return match is not None
