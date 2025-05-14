import json
import logging
import sys

from imap_client import ImapClient
from imap_log import init_logger
from utils import valid_subject, valid_from_email


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
        unique_system_emails = []

        for group_settings in group_settings_list:
            unique_system_emails.append(group_settings["systemEmail"])
            all_subject_line_regexes = all_subject_line_regexes + group_settings["subjectLineRegexes"]
            all_email_whitelist = all_email_whitelist + group_settings["emailWhiteList"]

        if len(unique_system_emails) == 0:
            logging.error("No system emails found")
            print(json.dumps([]))
            exit(0)

        unique_system_emails_dict = {email["id"]: email for email in unique_system_emails}
        unique_system_emails = list(unique_system_emails_dict.values())

        for system_email in unique_system_emails:
            client = ImapClient(
                system_email["host"],
                system_email["port"],
                system_email["username"],
                system_email["password"],
                system_email["useStartTLS"],
                all_subject_line_regexes,
                all_email_whitelist
            )
            all_unread_email_metadata = all_unread_email_metadata + client.get_unread_email_metadata()

        for metadata in all_unread_email_metadata:
            for group_settings in group_settings_list:
                if (valid_from_email(metadata["fromEmail"], group_settings["emailWhiteList"])
                        and valid_subject(metadata["subject"], group_settings["subjectLineRegexes"])):
                    metadata["groupSettingsIds"].append(group_settings["id"])

        logging.info(f"All metadata found: {all_unread_email_metadata}")
        print(json.dumps(all_unread_email_metadata))
        exit(0)

    except Exception as e:
        logging.error(e)
        exit(1)


if __name__ == "__main__":
    main()
