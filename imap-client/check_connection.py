import json
import sys

from imap_client import ImapClient


def read_system_email():
    raw_data = sys.stdin.read()
    json_data = json.loads(raw_data)

    return json_data


if __name__ == '__main__':
    credentials = read_system_email()
    client = ImapClient(credentials["host"], credentials["port"], credentials["username"], credentials["password"], [],
                        [])

    try:
        client.connect()
        exit(0)
    except Exception as e:
        print(e)
        exit(1)
