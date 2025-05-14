import json
import logging
import sys

from imap_client import ImapClient
from imap_log import init_logger


def read_system_email():
    raw_data = sys.stdin.read()
    json_data = json.loads(raw_data)

    return json_data


if __name__ == '__main__':
    init_logger()
    credentials = read_system_email()
    client = ImapClient(
        credentials["host"],
        credentials["port"],
        credentials["username"],
        credentials["password"],
        credentials["useStartTLS"],
        [],
        []
    )

    try:
        client.connect()
        exit(0)
    except Exception as e:
        logging.error(e)
        print(e)
        exit(1)
