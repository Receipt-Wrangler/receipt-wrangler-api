import unittest
from email.message import Message
from unittest.mock import patch

from imap_client import ImapClient


class TestShouldSetUpClientCorrectly(unittest.TestCase):

    def test_constructor(self):
        client = ImapClient("host", "port", "username", "password", False, [], [])
        self.assertEqual(client.host, "host")
        self.assertEqual(client.port, "port")
        self.assertEqual(client.username, "username")
        self.assertEqual(client.password, "password")

    @patch('imap_client.IMAPClient')
    def test_catch_error_with_bad_connect(self, mock_imapclient):
        mock_imapclient.side_effect = Exception('Failed to connect')
        client = ImapClient("host", 993, "username", "password", False, [], [])
        with self.assertRaises(Exception) as context:
            client.connect()
        self.assertTrue('Failed to connect' in str(context.exception))

    def setUp(self):
        self.client = ImapClient('host', 'port', 'username', 'password', False, [], [])

    def test_get_formatted_to_or_from_data(self):
        message = Message()
        message['From'] = 'Test User <test@example.com>'
        result = self.client._get_formatted_to_or_from_data(message, 'From')
        self.assertEqual(result, {'name': 'Test User ', 'email': 'test@example.com'})

    def test_get_formatted_date(self):
        date = 'Wed, 20 Oct 2021 10:30:00 +0000'
        result = self.client.get_formatted_date(date)
        self.assertEqual(result, '2021-10-20T10:30:00.000000Z')

    def test_valid_mime_type(self):
        mime_type = 'image/jpeg'
        result = self.client.valid_mime_type(mime_type)
        self.assertTrue(result)

    def test_invalid_mime_type(self):
        mime_type = 'text/plain'
        result = self.client.valid_mime_type(mime_type)
        self.assertFalse(result)


if __name__ == '__main__':
    unittest.main()
