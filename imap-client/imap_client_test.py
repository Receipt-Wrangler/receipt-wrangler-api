import unittest

from imapclient import exceptions as imapclientexceptions

from imap_client import ImapClient


class TestShouldSetUpClientCorrectly(unittest.TestCase):

    def test_constructor(self):
        client = ImapClient("host", "port", "username", "password")
        self.assertEqual(client.host, "host")
        self.assertEqual(client.port, "port")
        self.assertEqual(client.username, "username")
        self.assertEqual(client.password, "password")

    def test_catch_error_with_bad_connect(self):
        client = ImapClient("host", 993, "username", "password")
        client.connect()

        self.assertRaises(imapclientexceptions.LoginError)


if __name__ == '__main__':
    unittest.main()
