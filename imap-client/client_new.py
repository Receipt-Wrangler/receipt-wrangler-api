import json
import logging
import os
import sys

from client import valid_from_email
from imap_client import ImapClient
from utils import valid_subject

base_path = os.environ.get("BASE_PATH", "")


def init_logger():
    path = os.path.join(base_path, "logs", "imap-client.log")
    logging.basicConfig(filename=path, level=logging.INFO,
                        format='%(asctime)s %(levelname)s {%(pathname)s:%(lineno)d} %(message)s')


def read_group_settings():
    raw_data = sys.stdin.read()
    json_data = json.loads(raw_data)

    return json_data


def main():
    try:
        init_logger()
        group_settings_list = read_group_settings()
        all_subject_line_regexes = []
        all_email_whitelist = []
        all_unread_email_metadata = []
        unique_system_emails = {}
        metadata_group_settings_map = {}

        for group_settings in group_settings_list:
            unique_system_emails.update(group_settings["systemEmail"])
            all_subject_line_regexes.append(group_settings["subjectLineRegexes"])
            all_email_whitelist.append(group_settings["emailWhiteList"])
            metadata_group_settings_map[group_settings["id"]] = []

        if len(unique_system_emails) == 0:
            logging.error("No system emails found")
            print(metadata_group_settings_map)
            exit(0)

        for system_email in unique_system_emails:
            client = ImapClient(
                system_email["host"],
                system_email["port"],
                system_email["username"],
                system_email["password"],
                all_subject_line_regexes,
                all_email_whitelist
            )
            all_unread_email_metadata.append(client.get_unread_email_metadata())

        for metadata in all_unread_email_metadata:
            for group_settings in group_settings_list:
                if (valid_from_email(metadata["fromEmail"], group_settings["emailWhiteList"])
                        and valid_subject(metadata["subject"], group_settings["subjectLineRegexes"])):
                    metadata_group_settings_map[metadata].append(group_settings["id"])

        print(metadata_group_settings_map)
        exit(0)

    except Exception as e:
        logging.error(e)
        exit(1)


if __name__ == "__main__":
    main()
